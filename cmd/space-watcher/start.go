/*
 *  Copyright 2021 qitoi
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	twitter11 "github.com/dghubble/go-twitter/twitter"
	"go.uber.org/zap"

	"github.com/qitoi/space-watcher/bot"
	"github.com/qitoi/space-watcher/db"
	"github.com/qitoi/space-watcher/logger"
	"github.com/qitoi/space-watcher/oauth1"
	twitter2 "github.com/qitoi/space-watcher/twitter"
)

type watcher struct {
	config    *Config
	logger    *zap.SugaredLogger
	clientV11 *twitter11.Client
	clientV2  *twitter2.Client
	dbClient  *db.Client
}

func Start(config *Config) error {
	var err error
	if err := CheckValidConfig(config); err != nil {
		return err
	}

	log, err := getLogger(config)
	if err != nil {
		return err
	}
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// signal handler (usr1: reopen log file)
	startSignalHandler(log)

	dbClient, err := db.Open("./space-watcher.db")
	defer dbClient.Close()

	// twitter api v1.1 client
	auth := oauth1.NewAuth(config.Twitter.ConsumerKey, config.Twitter.ConsumerSecret)
	httpClient := auth.GetHttpClient(context.Background(), config.Twitter.AccessToken, config.Twitter.AccessSecret)
	clientV11 := twitter11.NewClient(httpClient)

	// twitter api v2 client
	clientV2 := twitter2.NewClient(config.Twitter.BearerToken)

	w := &watcher{
		config:    config,
		logger:    log.Sugar(),
		clientV11: clientV11,
		clientV2:  clientV2,
		dbClient:  dbClient,
	}

	// monitoring target = followings
	ids, err := w.getFollowings(config.Twitter.UserID)
	if err != nil {
		return err
	}
	creatorIDs := make([]string, len(ids))
	for i, id := range ids {
		creatorIDs[i] = strconv.FormatInt(id, 10)
	}

	w.logger.Infow("target users", "users", creatorIDs)

	// start http server for health check
	if config.HealthCheck.Enabled {
		w.startHealthCheckServer(*w.config.HealthCheck.Port)
	}

	w.startWatch(ctx, creatorIDs)

	return nil
}

func getLogger(config *Config) (*logger.Logger, error) {
	var err error
	infoLog := logger.Wrap(os.Stdout)
	if config.Logger != nil && config.Logger.Info != nil {
		infoLog, err = logger.OpenFile(*config.Logger.Info, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
	}

	errorLog := logger.Wrap(os.Stderr)
	if config.Logger != nil && config.Logger.Error != nil {
		errorLog, err = logger.OpenFile(*config.Logger.Error, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
	}

	return logger.New(infoLog, errorLog), nil
}

func (w *watcher) startWatch(ctx context.Context, creatorIDs []string) {
	// interval [s]
	baseInterval := w.config.Bot.WatchInterval
	interval := baseInterval

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		spaces, users, rate, err := w.getSpaces(ctx, creatorIDs)
		if err != nil {
			w.logger.Errorw("watch spaces error", "error", err)
		}
		w.logger.Infow("watch spaces result", "spaces", spaces, "users", users, "rate", rate)

		if spaces != nil && users != nil {
			err = w.processSpaces(spaces, users)
			if err != nil {
				w.logger.Errorw("notify space error", "error", err)
			}
		}

		if rate != nil {
			resetTime := rate.Reset.Sub(time.Now()).Seconds()
			intervalForReset := int64(math.Ceil(resetTime / float64(rate.Remaining+1)))
			nextInterval := interval

			if intervalForReset != interval {
				nextInterval = intervalForReset
			}

			if nextInterval < baseInterval {
				nextInterval = baseInterval
			}

			if nextInterval != interval {
				interval = nextInterval
				ticker.Reset(time.Duration(interval) * time.Second)
			}
		}
	}
}

func (w *watcher) getFollowings(userID int64) ([]int64, error) {
	friendsResp, _, err := w.clientV11.Friends.IDs(&twitter11.FriendIDParams{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return friendsResp.IDs, nil
}

func (w *watcher) getSpaces(ctx context.Context, creatorIDs []string) ([]twitter2.Space, map[string]twitter2.User, *twitter2.RateLimit, error) {
	resp, rate, err := w.clientV2.GetSpacesByCreatorIDs(
		ctx,
		twitter2.SpacesByCreatorIDsRequest{
			UserIDs:     creatorIDs,
			Expansions:  []string{"creator_id"},
			SpaceFields: []string{"id", "title", "creator_id", "state", "started_at", "scheduled_start", "created_at", "updated_at"},
			UserFields:  []string{"id", "name", "username"},
		})
	if err != nil {
		return nil, nil, rate, err
	}

	spaces := make([]twitter2.Space, 0)
	if resp.Data != nil {
		spaces = resp.Data
	}

	users := make(map[string]twitter2.User)
	if resp.Includes != nil && resp.Includes.Users != nil {
		for _, u := range *resp.Includes.Users {
			users[u.ID] = u
		}
	}

	return spaces, users, rate, nil
}

func (w *watcher) processSpaces(spaces []twitter2.Space, users map[string]twitter2.User) error {
	ch := make(chan error)
	var wg sync.WaitGroup

	wg.Add(len(spaces))
	for _, space := range spaces {
		s := space
		u := users[s.CreatorID]
		go func() {
			defer wg.Done()
			ch <- w.notifySpace(&s, &u)
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var err error
	for e := range ch {
		if e != nil {
			w.logger.Errorw("notify error", "error", e)
			err = e
		}
	}

	return err
}

func (w *watcher) notifySpace(space *twitter2.Space, user *twitter2.User) error {
	currentStatus, err := w.getNotificationStatus(space)
	if err != nil {
		return err
	}

	if notified, err := w.dbClient.CheckNotified(space.ID, currentStatus); err != nil {
		return err
	} else if notified {
		return nil
	}

	w.logger.Infow("notify", "space", *space, "user", *user, "status", currentStatus)

	_, err = w.tweetSpace(currentStatus, space, user)
	if err != nil {
		return err
	}

	switch currentStatus {
	case db.SpaceNotificationStatus_SCHEDULE:
		return w.dbClient.RegisterSchedule(space.ID, user.ID, user.Username, space.Title, *space.ScheduledStart, *space.CreatedAt)
	case db.SpaceNotificationStatus_SCHEDULE_REMIND:
		return w.dbClient.RegisterScheduleRemind(space.ID, user.ID, user.Username, space.Title, *space.ScheduledStart, *space.CreatedAt)
	case db.SpaceNotificationStatus_START:
		return w.dbClient.RegisterStart(space.ID, user.ID, user.Username, space.Title, *space.StartedAt, *space.CreatedAt)
	}

	return nil
}

func (w *watcher) getNotificationStatus(space *twitter2.Space) (db.SpaceNotificationStatus, error) {
	if space.State == nil {
		return db.SpaceNotificationStatus_NONE, errors.New("invalid space info")
	}
	switch *space.State {
	case "scheduled":
		if space.ScheduledStart == nil {
			return db.SpaceNotificationStatus_NONE, errors.New("invalid space info")
		}

		// リマインド通知が有効で、リマインド時間を過ぎていればリマインド
		start := *space.ScheduledStart
		if w.config.Bot.Notification.ScheduleRemind.Enabled {
			reminderTime := start.Add(-time.Duration(*w.config.Bot.Notification.ScheduleRemind.Before) * time.Second)
			if time.Now().After(reminderTime) {
				return db.SpaceNotificationStatus_SCHEDULE_REMIND, nil
			}
		}

		// スケジュール作成通知が有効
		if w.config.Bot.Notification.Schedule.Enabled {
			return db.SpaceNotificationStatus_SCHEDULE, nil
		}

		break
	case "live":
		// 開始済み
		return db.SpaceNotificationStatus_START, nil
	}

	return db.SpaceNotificationStatus_NONE, nil
}

func (w *watcher) tweetSpace(status db.SpaceNotificationStatus, space *twitter2.Space, user *twitter2.User) (int64, error) {
	var template string
	switch status {
	case db.SpaceNotificationStatus_SCHEDULE:
		template = *w.config.Bot.Notification.Schedule.Message
		break
	case db.SpaceNotificationStatus_SCHEDULE_REMIND:
		template = *w.config.Bot.Notification.ScheduleRemind.Message
		break
	case db.SpaceNotificationStatus_START:
		template = *w.config.Bot.Notification.Start.Message
		break
	default:
		return 0, errors.New("invalid notification status")
	}

	message, err := bot.GetTweetMessage(template, space, user)
	if err != nil {
		return 0, err
	}
	tweet, _, err := w.clientV11.Statuses.Update(message, nil)
	if err != nil {
		return 0, err
	}
	w.logger.Infow("tweet completed", "message", message)
	return tweet.ID, nil
}

func (w *watcher) startHealthCheckServer(port int) {
	go func() {
		address := fmt.Sprintf(":%d", port)
		w.logger.Infow("start http server for health check", "address", address)
		err := http.ListenAndServe(address, http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
			w.logger.Debugw("health check access", "uri", r.RequestURI, "remote_addr", r.RemoteAddr)
			res.WriteHeader(http.StatusOK)
		}))
		if err != nil {
			w.logger.Errorw("http server for health check failed", "address", address, "error", err)
		}
	}()
}

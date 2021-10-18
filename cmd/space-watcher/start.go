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
	"strconv"
	"sync"
	"time"

	twitter11 "github.com/dghubble/go-twitter/twitter"

	"github.com/qitoi/space-watcher/bot"
	"github.com/qitoi/space-watcher/db"
	"github.com/qitoi/space-watcher/oauth1"
	twitter2 "github.com/qitoi/space-watcher/twitter"
)

func (c *Command) Start() error {
	if err := CheckValidConfig(*c.Config); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbClient, err := db.Open("./space-watcher.db")
	defer dbClient.Close()

	// twitter api v1.1 client
	auth := oauth1.NewAuth(c.Config.Twitter.ConsumerKey, c.Config.Twitter.ConsumerSecret)
	httpClient := auth.GetHttpClient(context.Background(), c.Config.Twitter.AccessToken, c.Config.Twitter.AccessSecret)
	clientV11 := twitter11.NewClient(httpClient)

	// twitter api v2 client
	clientV2 := twitter2.NewClient(c.Config.Twitter.BearerToken)

	// monitoring target = followings
	ids, err := c.getFollowings(clientV11, c.Config.Twitter.UserID)
	if err != nil {
		return err
	}
	creatorIDs := make([]string, len(ids))
	for i, id := range ids {
		creatorIDs[i] = strconv.FormatInt(id, 10)
	}

	c.Logger.Infow("target users", "users", creatorIDs)

	// start http server for health check
	if c.Config.HealthCheck.Enabled {
		c.StartHealthCheckServer(*c.Config.HealthCheck.Port)
	}

	// interval [s]
	baseInterval := c.Config.Bot.WatchInterval
	interval := baseInterval

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		spaces, users, rate, err := c.watchSpaces(ctx, clientV2, creatorIDs)
		if err != nil {
			c.Logger.Errorw("watch spaces error", "error", err)
		}
		c.Logger.Infow("watch spaces result", "spaces", spaces, "users", users, "rate", rate)

		if spaces != nil && users != nil {
			err = c.processSpaces(dbClient, clientV11, spaces, users)
			if err != nil {
				c.Logger.Errorw("notify space error", "error", err)
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

	return nil
}

func (c *Command) getFollowings(clientV11 *twitter11.Client, userID int64) ([]int64, error) {
	friendsResp, _, err := clientV11.Friends.IDs(&twitter11.FriendIDParams{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return friendsResp.IDs, nil
}

func (c *Command) watchSpaces(ctx context.Context, clientV2 *twitter2.Client, creatorIDs []string) ([]twitter2.Space, map[string]twitter2.User, *twitter2.RateLimit, error) {
	resp, rate, err := clientV2.GetSpacesByCreatorIDs(
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

func (c *Command) processSpaces(dbClient *db.Client, clientV11 *twitter11.Client, spaces []twitter2.Space, users map[string]twitter2.User) error {
	ch := make(chan error)
	var wg sync.WaitGroup

	wg.Add(len(spaces))
	for _, space := range spaces {
		s := space
		u := users[s.CreatorID]
		go func() {
			defer wg.Done()
			ch <- c.notifySpace(dbClient, clientV11, &s, &u)
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var err error
	for e := range ch {
		if e != nil {
			c.Logger.Errorw("notify error", "error", e)
			err = e
		}
	}

	return err
}

func (c *Command) notifySpace(dbClient *db.Client, clientV11 *twitter11.Client, space *twitter2.Space, user *twitter2.User) error {
	currentStatus, err := c.getNotificationStatus(space)
	if err != nil {
		return err
	}

	if notified, err := dbClient.CheckNotified(space.ID, currentStatus); err != nil {
		return err
	} else if notified {
		return nil
	}

	c.Logger.Infow("notify", "space", *space, "user", *user, "status", currentStatus)

	_, err = c.tweetSpace(clientV11, currentStatus, space, user)
	if err != nil {
		return err
	}

	switch currentStatus {
	case db.SpaceNotificationStatus_SCHEDULE:
		return dbClient.RegisterSchedule(space.ID, user.ID, user.Username, space.Title, *space.ScheduledStart, *space.CreatedAt)
	case db.SpaceNotificationStatus_SCHEDULE_REMIND:
		return dbClient.RegisterScheduleRemind(space.ID, user.ID, user.Username, space.Title, *space.ScheduledStart, *space.CreatedAt)
	case db.SpaceNotificationStatus_START:
		return dbClient.RegisterStart(space.ID, user.ID, user.Username, space.Title, *space.StartedAt, *space.CreatedAt)
	}

	return nil
}

func (c *Command) getNotificationStatus(space *twitter2.Space) (db.SpaceNotificationStatus, error) {
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
		if c.Config.Bot.Notification.ScheduleRemind.Enabled {
			reminderTime := start.Add(-time.Duration(*c.Config.Bot.Notification.ScheduleRemind.Before) * time.Second)
			if time.Now().After(reminderTime) {
				return db.SpaceNotificationStatus_SCHEDULE_REMIND, nil
			}
		}

		// スケジュール作成通知が有効
		if c.Config.Bot.Notification.Schedule.Enabled {
			return db.SpaceNotificationStatus_SCHEDULE, nil
		}

		break
	case "live":
		// 開始済み
		return db.SpaceNotificationStatus_START, nil
	}

	return db.SpaceNotificationStatus_NONE, nil
}

func (c *Command) tweetSpace(clientV11 *twitter11.Client, status db.SpaceNotificationStatus, space *twitter2.Space, user *twitter2.User) (int64, error) {
	var template string
	switch status {
	case db.SpaceNotificationStatus_SCHEDULE:
		template = *c.Config.Bot.Notification.Schedule.Message
		break
	case db.SpaceNotificationStatus_SCHEDULE_REMIND:
		template = *c.Config.Bot.Notification.ScheduleRemind.Message
		break
	case db.SpaceNotificationStatus_START:
		template = *c.Config.Bot.Notification.Start.Message
		break
	default:
		return 0, errors.New("invalid notification status")
	}

	message, err := bot.GetTweetMessage(template, space, user)
	if err != nil {
		return 0, err
	}
	tweet, _, err := clientV11.Statuses.Update(message, nil)
	if err != nil {
		return 0, err
	}
	c.Logger.Infow("tweet completed", "message", message)
	return tweet.ID, nil
}

func (c *Command) StartHealthCheckServer(port int) {
	go func() {
		address := fmt.Sprintf(":%d", port)
		c.Logger.Infow("start http server for health check", "address", address)
		err := http.ListenAndServe(address, http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
			c.Logger.Infow("health check access", "uri", r.RequestURI, "remote_addr", r.RemoteAddr)
			res.WriteHeader(http.StatusOK)
		}))
		if err != nil {
			c.Logger.Infow("http server for health check failed", "address", address, "error", err)
		}
	}()
}

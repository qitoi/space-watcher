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
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	twitter11 "github.com/dghubble/go-twitter/twitter"
	"google.golang.org/protobuf/types/known/timestamppb"

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

	// start http server for health check
	if c.Config.HealthCheck != nil && *c.Config.HealthCheck.Enabled {
		StartHealthCheckServer(*c.Config.HealthCheck.Port)
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
			err = c.notify(dbClient, clientV11, spaces, users)
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
			SpaceFields: []string{"id", "title", "creator_id", "started_at", "created_at"},
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

func (c *Command) notify(dbClient *db.Client, clientV11 *twitter11.Client, spaces []twitter2.Space, users map[string]twitter2.User) error {
	ch := make(chan error)
	var wg sync.WaitGroup

	wg.Add(len(spaces))
	for _, space := range spaces {
		s := space
		u := users[s.CreatorID]
		go func() {
			defer wg.Done()
			if !dbClient.CheckNotified(s.ID) {
				c.Logger.Infow("notify", "space", s, "user", u)

				tweetID, err := c.tweetSpace(clientV11, s, u)
				if err != nil {
					ch <- err
					return
				}
				ch <- dbClient.RegisterNotified(&db.Space{
					Id:            s.ID,
					CreatorId:     u.ID,
					ScreenName:    u.Username,
					Title:         s.Title,
					NotifyTweetId: tweetID,
					StartedAt:     timestamppb.New(*s.StartedAt),
					CreatedAt:     timestamppb.New(*s.CreatedAt),
				})
			}
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

func (c *Command) tweetSpace(clientV11 *twitter11.Client, space twitter2.Space, user twitter2.User) (int64, error) {
	message, err := bot.GetTweetMessage(c.Config.Bot.Message, space, user)
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

func StartHealthCheckServer(port int) {
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", port), http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
			res.WriteHeader(http.StatusOK)
		}))
	}()
}

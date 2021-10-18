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

package db

import (
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	bucketSpace = "space"
)

type Client struct {
	db *bolt.DB
}

func Open(path string) (*Client, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketSpace))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		db: db,
	}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) GetNotifiedStatus(spaceID string) (SpaceNotificationStatus, error) {
	key := spaceID
	var status SpaceNotificationStatus
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketSpace))
		if b == nil {
			return errors.New("bucket not found: " + bucketSpace)
		}

		data := b.Get([]byte(key))

		if data == nil {
			status = SpaceNotificationStatus_NONE
			return nil
		}

		var s Space
		err := proto.Unmarshal(data, &s)
		if err != nil {
			return err
		}

		status = s.NotificationStatus
		return nil
	})
	return status, err
}

func (c *Client) CheckNotified(spaceID string, status SpaceNotificationStatus) (bool, error) {
	prevStatus, err := c.GetNotifiedStatus(spaceID)
	if err != nil {
		return false, err
	}
	return status <= prevStatus, nil
}

func (c *Client) RegisterSchedule(spaceID, creatorID, screenName, title string, scheduledStart, createdAt time.Time) error {
	return c.register(&Space{
		Id:                 spaceID,
		CreatorId:          creatorID,
		ScreenName:         screenName,
		Title:              title,
		NotificationStatus: SpaceNotificationStatus_SCHEDULE,
		ScheduledStart:     timestamppb.New(scheduledStart),
		StartedAt:          nil,
		CreatedAt:          timestamppb.New(createdAt),
	})
}

func (c *Client) RegisterScheduleRemind(spaceID, creatorID, screenName, title string, scheduledStart, createdAt time.Time) error {
	return c.register(&Space{
		Id:                 spaceID,
		CreatorId:          creatorID,
		ScreenName:         screenName,
		Title:              title,
		NotificationStatus: SpaceNotificationStatus_SCHEDULE_REMIND,
		ScheduledStart:     timestamppb.New(scheduledStart),
		StartedAt:          nil,
		CreatedAt:          timestamppb.New(createdAt),
	})
}

func (c *Client) RegisterStart(spaceID, creatorID, screenName, title string, startedAt, createdAt time.Time) error {
	return c.register(&Space{
		Id:                 spaceID,
		CreatorId:          creatorID,
		ScreenName:         screenName,
		Title:              title,
		NotificationStatus: SpaceNotificationStatus_START,
		ScheduledStart:     nil,
		StartedAt:          timestamppb.New(startedAt),
		CreatedAt:          timestamppb.New(createdAt),
	})
}

func (c *Client) register(record *Space) error {
	data, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	key := record.Id
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketSpace))
		return b.Put([]byte(key), data)
	})
}

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
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type SpaceRecord struct {
	ID            string    `db:"id"`
	CreatorID     string    `db:"creator_id"`
	ScreenName    string    `db:"screen_name"`
	Title         string    `db:"title"`
	NotifyTweetID int64     `db:"notify_tweet_id"`
	StartedAt     time.Time `db:"started_at"`
	CreatedAt     time.Time `db:"created_at"`
}

func OpenDB(filename string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("create table if not exists space (id text primary key, creator_id text, screen_name text, title text, notify_tweet_id integer, started_at datetime, created_at datetime)")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CheckNotified(q sqlx.Queryer, spaceID string) bool {
	var count int64
	err := sqlx.Get(q, &count, "select count(id) from space where id = ?", spaceID)
	if err != nil {
		return false
	}
	return count == 1
}

func RegisterNotified(e sqlx.Execer, record SpaceRecord) error {
	_, err := e.Exec("insert into space values (?, ?, ?, ?, ?, ?, ?)", record.ID, record.CreatorID, record.ScreenName, record.Title, record.NotifyTweetID, record.StartedAt, record.CreatedAt)
	return err
}

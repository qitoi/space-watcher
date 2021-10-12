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

package twitter

import (
	"time"
)

type User struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Username        string  `json:"username"`
	Description     *string `json:"description,omitempty"`
	Location        *string `json:"location,omitempty"`
	PinnedTweetID   *string `json:"pinned_tweet_id,omitempty"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
	Protected       *string `json:"protected,omitempty"`
	PublicMetrics   *struct {
		FollowersCount int64 `json:"followers_count,omitempty"`
		FollowingCount int64 `json:"following_count,omitempty"`
		TweetCount     int64 `json:"tweet_count,omitempty"`
		ListedCount    int64 `json:"listed_count,omitempty"`
	} `json:"public_metrics,omitempty"`
	URL       *string    `json:"url,omitempty"`
	Verified  *bool      `json:"verified,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

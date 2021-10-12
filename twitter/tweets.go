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

type Tweet struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	Attachments *struct {
		PollIDs   []string `json:"poll_ids"`
		MediaKeys []string `json:"media_keys"`
	} `json:"attachments,omitempty"`
	AuthorID         *string    `json:"author_id,omitempty"`
	ConversationID   *string    `json:"conversation_id,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	InReplyToUserID  *string    `json:"in_reply_to_user_id,omitempty"`
	Lang             *string    `json:"lang,omitempty"`
	NonPublicMetrics *struct {
		ImpressionCount   int64 `json:"impression_count"`
		URLLinkClicks     int64 `json:"url_link_clicks"`
		UserProfileClicks int64 `json:"user_profile_clicks"`
	} `json:"non_public_metrics,omitempty"`
	OrganicMetrics *struct {
		ImpressionCount   int64 `json:"impression_count"`
		LikeCount         int64 `json:"like_count"`
		ReplyCount        int64 `json:"reply_count"`
		RetweetCount      int64 `json:"retweet_count"`
		URLLinkClicks     int64 `json:"url_link_clicks"`
		UserProfileClicks int64 `json:"user_profile_clicks"`
	} `json:"organic_metrics,omitempty"`
	PossiblySensitive *bool `json:"possibly_sensitive,omitempty"`
	PromotedMetrics   *struct {
		ImpressionCount   int64 `json:"impression_count"`
		LikeCount         int64 `json:"like_count"`
		ReplyCount        int64 `json:"reply_count"`
		RetweetCount      int64 `json:"retweet_count"`
		URLLinkClicks     int64 `json:"url_link_clicks"`
		UserProfileClicks int64 `json:"user_profile_clicks"`
	} `json:"promoted_metrics,omitempty"`
	PublicMetrics *struct {
		RetweetCount int64 `json:"retweet_count"`
		ReplyCount   int64 `json:"reply_count"`
		LikeCount    int64 `json:"like_count"`
		QuoteCount   int64 `json:"quote_count"`
	} `json:"public_metrics,omitempty"`
	ReferencedTweets *[]struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"referenced_tweets,omitempty"`
	ReplySettings *string `json:"reply_settings,omitempty"`
	Source        *string `json:"source,omitempty"`
	Withheld      *struct {
		Copyright    bool     `json:"copyright"`
		CountryCodes []string `json:"country_codes"`
	} `json:"withheld,omitempty"`
}

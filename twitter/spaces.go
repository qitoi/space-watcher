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
	"context"
	"errors"
	"fmt"
	"time"
)

type Space struct {
	ID               string     `json:"id"`
	CreatorID        string     `json:"creator_id"`
	Title            string     `json:"title"`
	Lang             *string    `json:"lang,omitempty"`
	State            *string    `json:"state,omitempty"`
	IsTicketed       *bool      `json:"is_ticketed,omitempty"`
	HostIds          *[]string  `json:"host_ids,omitempty"`
	InvitedUserIDs   *[]string  `json:"invited_user_ids,omitempty"`
	SpeakerIDs       *[]string  `json:"speaker_ids,omitempty"`
	ParticipantCount *int64     `json:"participant_count,omitempty"`
	ScheduledStart   *time.Time `json:"scheduled_start,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

type SpacesByCreatorIDsRequest struct {
	UserIDs     []string
	Expansions  []string
	SpaceFields []string
	UserFields  []string
}

type SpacesByCreatorIDsResponse struct {
	Data     []Space `json:"data"`
	Includes *struct {
		Users *[]User `json:"users,omitempty"`
	} `json:"includes,omitempty"`
}

func (c *Client) GetSpacesByCreatorIDs(ctx context.Context, req SpacesByCreatorIDsRequest) (*SpacesByCreatorIDsResponse, *RateLimit, error) {
	if len(req.UserIDs) == 0 {
		return nil, nil, errors.New("invalid parameter")
	}

	params := make(map[string]string)

	setRequestParam(params, "user_ids", req.UserIDs)
	setRequestParam(params, "expansions", req.Expansions)
	setRequestParam(params, "space.fields", req.SpaceFields)
	setRequestParam(params, "user.fields", req.UserFields)

	var r SpacesByCreatorIDsResponse
	rate, err := c.Get(ctx, "spaces/by/creator_ids", params, &r)

	if err != nil {
		return nil, rate, err
	}

	return &r, rate, nil
}

func GetSpaceURL(spaceID string) string {
	return fmt.Sprintf("https://twitter.com/i/spaces/%s", spaceID)
}

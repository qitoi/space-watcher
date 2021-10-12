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

package bot

import (
	"testing"

	twitter2 "github.com/qitoi/spaces-notify-bot/twitter"
)

func TestEscapeMessage(t *testing.T) {
	msg := "@abc @@def efg@hij @klm-@nop http://example.com/@qrs #@tuv"
	actual := EscapeMessage(msg)
	expected := "@.abc @@.def efg@hij @.klm-@.nop http://example.com/@qrs #@tuv"
	if actual != expected {
		t.Errorf("EscapeMessage(%s), actual: %s, expecte: %s", msg, actual, expected)
	}
}

func TestGetTweetMessage(t *testing.T) {

	message := "{{.User.Name | escape}} starts Spaces {{.URL}}"
	actual, err := GetTweetMessage(
		message,
		twitter2.Space{
			ID:    "spaceid",
			Title: "SPACE_TITLE",
		},
		twitter2.User{
			Name: "UserName@test",
		},
	)

	if err != nil {
		t.Error(err)
	}

	expected := "UserName@.test starts Spaces https://twitter.com/i/spaces/spaceid"
	if actual != expected {
		t.Errorf("GetTweetMessage, actual: %s, expected: %s", actual, expected)
	}
}

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
	"strings"
	"text/template"

	"github.com/kylemcc/twitter-text-go/extract"

	twitter2 "github.com/qitoi/space-watcher/twitter"
)

func EscapeMessage(message string) string {
	entries := extract.ExtractMentionedScreenNames(message)
	urlEntries := extract.ExtractUrls(message)

	var reps []string

loop:
	for _, entry := range entries {
		for _, urlEntry := range urlEntries {
			if urlEntry.Range.Start <= entry.Range.Start && entry.Range.Stop <= urlEntry.Range.Stop {
				continue loop
			}
		}
		screenname, _ := entry.ScreenName()
		reps = append(reps, "@"+screenname, "@."+screenname)
	}
	return strings.NewReplacer(reps...).Replace(message)
}

func RenderTemplate(message string, space *twitter2.Space, user *twitter2.User) (string, error) {
	url := twitter2.GetSpaceURL(space.ID)

	t, err := template.New("message").
		Funcs(map[string]interface{}{
			"escape": EscapeMessage,
		}).
		Parse(message)

	if err != nil {
		return "", err
	}

	sb := &strings.Builder{}
	err = t.Execute(sb, struct {
		User  twitter2.User
		Space twitter2.Space
		URL   string
	}{
		Space: *space,
		User:  *user,
		URL:   url,
	})
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

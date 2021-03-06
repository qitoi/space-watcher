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
	"bufio"
	"context"
	"fmt"
	"os"

	twitter11 "github.com/dghubble/go-twitter/twitter"

	"github.com/qitoi/space-watcher/oauth1"
	twitter2 "github.com/qitoi/space-watcher/twitter"
)

func InitializeToken(config *Config) error {
	auth := oauth1.NewAuth(config.Twitter.ConsumerKey, config.Twitter.ConsumerSecret)
	url, err := auth.GetAuthorizationURL("oob")

	if err != nil {
		return err
	}

	fmt.Printf("authorization url: %s\n", url.String())

	var verifier string
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("PIN: ")
		if !scanner.Scan() {
			break
		}

		verifier = scanner.Text()
		if verifier != "" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	accessToken, accessSecret, err := auth.GetAccessToken(verifier)
	if err != nil {
		return err
	}

	bearerToken, err := twitter2.GetBearerToken(config.Twitter.ConsumerKey, config.Twitter.ConsumerSecret)
	if err != nil {
		return err
	}

	httpClient := auth.GetHttpClient(context.Background(), accessToken, accessSecret)
	client := twitter11.NewClient(httpClient)
	user, _, err := client.Accounts.VerifyCredentials(nil)
	if err != nil {
		return err
	}

	newConfig := *config
	newConfig.Twitter.AccessToken = accessToken
	newConfig.Twitter.AccessSecret = accessSecret
	newConfig.Twitter.BearerToken = bearerToken
	newConfig.Twitter.UserID = user.ID

	return SaveConfig(newConfig)
}

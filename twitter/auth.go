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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	twitterAPIOAuth2Token = "https://api.twitter.com/oauth2/token"
)

type Oauth2TokenResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

func GetBearerToken(consumerKey, consumerSecret string) (string, error) {
	client := &http.Client{}

	rawCredential := consumerKey + ":" + consumerSecret
	credential := base64.StdEncoding.EncodeToString([]byte(rawCredential))

	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	req, err := http.NewRequest(
		"POST",
		twitterAPIOAuth2Token,
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic "+credential)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("oauth2/token error status: %v", resp.StatusCode)
	}

	var response Oauth2TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.AccessToken, nil
}

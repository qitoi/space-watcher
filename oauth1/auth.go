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

package oauth1

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
)

type Auth struct {
	config        *oauth1.Config
	requestToken  string
	requestSecret string
}

func NewAuth(consumerKey, consumerSecret string) *Auth {
	return &Auth{
		config: &oauth1.Config{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			Endpoint:       twitter.AuthorizeEndpoint,
		},
	}
}

func (a *Auth) GetAuthorizationURL(callbackURL string) (*url.URL, error) {
	a.config.CallbackURL = callbackURL
	requestToken, requestSecret, err := a.config.RequestToken()
	if err != nil {
		return nil, err
	}

	a.requestToken = requestToken
	a.requestSecret = requestSecret

	return a.config.AuthorizationURL(requestToken)
}

func (a *Auth) GetAccessToken(verifier string) (accessToken string, accessSecret string, err error) {
	return a.config.AccessToken(a.requestToken, a.requestSecret, verifier)
}

func (a *Auth) GetHttpClient(ctx context.Context, accessToken, accessSecret string) *http.Client {
	token := oauth1.NewToken(accessToken, accessSecret)
	return a.config.Client(ctx, token)
}

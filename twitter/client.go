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
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	twitterAPIv2 = "https://api.twitter.com/2/"
)

type Client struct {
	bearer string
}

type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

type Errors []struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type APIError struct {
	Errors     Errors
	StatusCode int
	Status     string
}

func (e *APIError) Error() string {
	return e.Status
}

func NewClient(bearer string) *Client {
	return &Client{
		bearer: bearer,
	}
}

func (c *Client) Get(ctx context.Context, api string, params map[string]string, out interface{}) (*RateLimit, error) {
	req, err := http.NewRequest("GET", twitterAPIv2+api, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}

	req.URL.RawQuery = q.Encode()

	return c.execRequest(req, out)
}

func (c *Client) Post(ctx context.Context, api string, params map[string]string, out interface{}) (*RateLimit, error) {
	req, err := http.NewRequest("POST", twitterAPIv2+api, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	q := url.Values{}
	for key, value := range params {
		q.Add(key, value)
	}

	req.URL.RawQuery = q.Encode()

	return c.execRequest(req, out)
}

func (c *Client) execRequest(req *http.Request, out interface{}) (*RateLimit, error) {
	req.Header.Set("Authorization", "Bearer "+c.bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rate := &RateLimit{}
	if ls := resp.Header.Get("X-Rate-Limit-Limit"); ls != "" {
		rate.Limit, err = strconv.Atoi(ls)
		if err != nil {
			return nil, err
		}
	}

	if rs := resp.Header.Get("X-Rate-Limit-Remaining"); rs != "" {
		rate.Remaining, err = strconv.Atoi(rs)
		if err != nil {
			return nil, err
		}
	}

	if rs := resp.Header.Get("X-Rate-Limit-Reset"); rs != "" {
		rn, err := strconv.Atoi(rs)
		if err != nil {
			return nil, err
		}
		rate.Reset = time.Unix(int64(rn), 0)
	}

	err = parseResponse(resp, out)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			if apiErr.StatusCode/100 == 5 {
				return nil, apiErr
			}
		}
	}

	return rate, err
}

func parseResponse(resp *http.Response, out interface{}) error {
	var m map[string]json.RawMessage

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return err
	}

	var errors Errors
	if raw, ok := m["errors"]; ok {
		if err := json.Unmarshal(raw, &errors); err != nil {
			return err
		}
		delete(m, "errors")
	}

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, out); err != nil {
		return err
	}

	status := resp.StatusCode / 100
	if len(errors) > 0 || status == 4 || status == 5 {
		return &APIError{
			Errors:     errors,
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	return nil
}

func setRequestParam(m map[string]string, key string, values []string) {
	if len(values) > 0 {
		m[key] = strings.Join(values, ",")
	}
}

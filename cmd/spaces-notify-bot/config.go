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
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type TwitterConfig struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
	AccessToken    string `yaml:"access_token"`
	AccessSecret   string `yaml:"access_secret"`
	BearerToken    string `yaml:"bearer_token"`
	UserID         int64  `yaml:"user_id"`
}

type BotConfig struct {
	SearchInterval int64  `yaml:"search_interval"`
	Message        string `yaml:"message"`
}

type HealthCheckConfig struct {
	Enabled *bool `yaml:"enabled"`
	Port    *int  `yaml:"port"`
}

type Config struct {
	Twitter     TwitterConfig      `yaml:"twitter"`
	Bot         BotConfig          `yaml:"bot"`
	HealthCheck *HealthCheckConfig `yaml:"healthcheck"`
}

func SaveConfig(config Config) error {
	file, err := os.OpenFile("./config.yaml", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	err = yaml.NewEncoder(file).Encode(config)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig() (*Config, error) {
	file, err := os.OpenFile("./config.yaml", os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func CheckMinimalValidConfig(config Config) error {
	if config.Twitter.ConsumerKey == "" {
		return errors.New("invalid config: twitter.consumer_key")
	}
	if config.Twitter.ConsumerSecret == "" {
		return errors.New("invalid config: twitter.consumer_secret")
	}
	return nil
}

func CheckValidConfig(config Config) error {
	if config.Twitter.ConsumerKey == "" {
		return errors.New("invalid config: twitter.consumer_key")
	}
	if config.Twitter.ConsumerSecret == "" {
		return errors.New("invalid config: twitter.consumer_secret")
	}
	if config.Twitter.AccessToken == "" {
		return errors.New("invalid config: twitter.access_token")
	}
	if config.Twitter.AccessSecret == "" {
		return errors.New("invalid config: twitter.access_secret")
	}
	if config.Twitter.BearerToken == "" {
		return errors.New("invalid config: twitter.bearer_token")
	}
	if config.Twitter.UserID == 0 {
		return errors.New("invalid config: twitter.user_id")
	}
	if config.Bot.SearchInterval == 0 {
		return errors.New("invalid config: bot.search_interval")
	}
	if config.Bot.Message == "" {
		return errors.New("invalid config: bot.message")
	}
	if config.HealthCheck != nil {
		if config.HealthCheck.Enabled == nil {
			return errors.New("config not found: healthcheck.enabled")
		}
		if *config.HealthCheck.Enabled {
			if config.HealthCheck.Port == nil {
				return errors.New("config not found: healthcheck.port")
			}
		}
		if config.HealthCheck.Port != nil {
			if *config.HealthCheck.Port <= 0 || *config.HealthCheck.Port > 65535 {
				return errors.New("invalid config: healthcheck.port")
			}
		}
	}
	return nil
}

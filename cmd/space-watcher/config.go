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

	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Twitter     TwitterConfig     `yaml:"twitter"`
	Event       EventConfig       `yaml:"event"`
	HealthCheck HealthCheckConfig `yaml:"healthcheck_server"`
	Logger      LoggerConfig      `yaml:"logger"`
}

type TwitterConfig struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
	AccessToken    string `yaml:"access_token"`
	AccessSecret   string `yaml:"access_secret"`
	BearerToken    string `yaml:"bearer_token"`
	UserID         int64  `yaml:"user_id"`
}

type EventConfig struct {
	WatchInterval  int64            `yaml:"watch_interval"`
	Schedule       *EventItemConfig `yaml:"schedule,omitempty"`
	ScheduleRemind *EventItemConfig `yaml:"schedule_remind,omitempty"`
	Start          *EventItemConfig `yaml:"start,omitempty"`
}

type EventItemConfig struct {
	Before       int64 `yaml:"before,omitempty"`
	Notification *struct {
		Message string `yaml:"message,omitempty"`
	} `yaml:"notification,omitempty"`
	Command *struct {
		Name             string   `yaml:"name"`
		Args             []string `yaml:"args"`
		WorkingDirectory string   `yaml:"working_directory"`
	} `yaml:"command,omitempty"`
}

type HealthCheckConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    *int `yaml:"port,omitempty"`
}

type LoggerConfig struct {
	Level *LogLevel `yaml:"level"`
	Info  *string   `yaml:"info"`
	Error *string   `yaml:"error"`
}

type LogLevel zapcore.Level

func (l *LogLevel) MarshalYAML() (interface{}, error) {
	return zapcore.Level(*l).MarshalText()
}

func (l *LogLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return err
	}
	*l = LogLevel(level)
	return nil
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

func CheckMinimalValidConfig(config *Config) error {
	if config.Twitter.ConsumerKey == "" {
		return errors.New("invalid config: twitter.consumer_key")
	}
	if config.Twitter.ConsumerSecret == "" {
		return errors.New("invalid config: twitter.consumer_secret")
	}
	return nil
}

func CheckValidConfig(config *Config) error {
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
	if config.Event.WatchInterval == 0 {
		return errors.New("invalid config: bot.watch_interval")
	}

	// Schedule
	if schedule := config.Event.Schedule; schedule != nil {
		if notif := schedule.Notification; notif != nil {
			if notif.Message == "" {
				return errors.New("invalid config: event.schedule.notification.message")
			}
		}
		if cmd := schedule.Command; cmd != nil {
			if cmd.Name == "" {
				return errors.New("invalid config: event.schedule.command.name")
			}
			if cmd.WorkingDirectory == "" {
				return errors.New("invalid config: event.schedule.command.working_directory")
			}
		}
	}

	// ScheduleRemind
	if scheduleRemind := config.Event.ScheduleRemind; scheduleRemind != nil {
		if scheduleRemind.Before <= 0 {
			return errors.New("invalid config: event.schedule_remind.notification.before")
		}
		if notif := scheduleRemind.Notification; notif != nil {
			if notif.Message == "" {
				return errors.New("invalid config: event.schedule_remind.notification.message")
			}
		}
		if cmd := scheduleRemind.Command; cmd != nil {
			if cmd.Name == "" {
				return errors.New("invalid config: event.schedule_remind.command.name")
			}
			if cmd.WorkingDirectory == "" {
				return errors.New("invalid config: event.schedule_remind.command.working_directory")
			}
		}
	}

	// Start
	if start := config.Event.Start; start != nil {
		if notif := start.Notification; notif != nil {
			if notif.Message == "" {
				return errors.New("invalid config: event.start.notification.message")
			}
		}
		if cmd := start.Command; cmd != nil {
			if cmd.Name == "" {
				return errors.New("invalid config: event.start.command.name")
			}
			if cmd.WorkingDirectory == "" {
				return errors.New("invalid config: event.start.command.working_directory")
			}
		}
	}

	// HealthCheck
	if config.HealthCheck.Enabled && config.HealthCheck.Port == nil {
		return errors.New("config not found: healthcheck.port")
	}
	if config.HealthCheck.Port != nil && (*config.HealthCheck.Port <= 0 || *config.HealthCheck.Port > 65535) {
		return errors.New("invalid config: healthcheck.port")
	}

	// Logger
	if config.Logger.Info != nil {
		if *config.Logger.Info == "" {
			return errors.New("invalid config: logger.info")
		}
	}
	if config.Logger.Error != nil {
		if *config.Logger.Error == "" {
			return errors.New("invalid config: logger.error")
		}
	}

	return nil
}

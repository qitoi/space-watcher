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
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func fail(logger *zap.SugaredLogger, err error) {
	logger.Errorw(err.Error(), "error", err)
	os.Exit(1)
}

func main() {
	var init bool
	var help bool

	pflag.BoolVarP(&init, "init", "", false, "initialize token")
	pflag.BoolVarP(&help, "help", "h", false, "help")

	pflag.Parse()

	if help {
		pflag.Usage()
		os.Exit(0)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	if err := CheckMinimalValidConfig(*config); err != nil {
		panic(err)
	}

	c := Command{
		Config: config,
		Logger: logger.Sugar(),
	}

	if init {
		if err := c.InitializeToken(); err != nil {
			fail(c.Logger, err)
		}
		os.Exit(0)
	}

	if err := c.Start(); err != nil {
		fail(c.Logger, err)
	}
	os.Exit(0)
}

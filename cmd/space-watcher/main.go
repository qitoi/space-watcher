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
	"log"
	"os"

	"github.com/spf13/pflag"
)

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

	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := CheckMinimalValidConfig(config); err != nil {
		log.Fatal(err)
	}

	if init {
		if err := InitializeToken(config); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if err := Start(config); err != nil {
		log.Fatal(err)
	}
}

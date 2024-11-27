// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"flag"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/server-common-lib/interrupts"
	"net/http"
	"os"
	"strconv"
)

const component = "robot-universal-access"

func main() {
	opt := new(robotOptions)
	cfg := opt.gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if opt.shutdown {
		return
	}

	bot := newRobot(cfg)
	interrupts.OnInterrupt(func() {
		bot.wait()
	})

	// Return 200 on / for health checks.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	// For /**-hook, handle a webhook normally.
	http.Handle("/"+opt.service.HandlePath, bot)
	httpServer := &http.Server{Addr: ":" + strconv.Itoa(opt.service.Port)}

	framework.StartupServer(httpServer, opt.service)
}

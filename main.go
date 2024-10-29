// Copyright 2024 Chao Feng
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
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/server-common-lib/interrupts"
	"github.com/opensourceways/server-common-lib/logrusutil"
	"net/http"
	"os"
	"strconv"
)

func main() {
	os.Args = append(os.Args,
		"--port=8888",
		"--config-file=D:\\Project\\github\\opensourceways\\robot-common\\robot-generic-access\\testdata\\config.yaml",
		"--component-name=robot-gitcode-access",
		"--handle-path=gitcode-hook",
	)
	opt := new(robotOptions)
	cfg := opt.gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if opt.shutdown {
		return
	}
	logrusutil.ComponentInit(opt.componentName)

	bot := newRobot(cfg)
	interrupts.OnInterrupt(func() {
		bot.Wait()
	})

	// Return 200 on / for health checks.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	// For /**-hook, handle a webhook normally.
	http.Handle("/"+opt.handlePath, bot)
	httpServer := &http.Server{Addr: ":" + strconv.Itoa(opt.service.Port)}

	framework.StartupServer(httpServer, opt.service, config.ServerAdditionOptions{})
}

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
	"github.com/opensourceways/server-common-lib/interrupts"
	"github.com/opensourceways/server-common-lib/logrusutil"
	"github.com/opensourceways/server-common-lib/options"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

type robotOptions struct {
	service       options.ServiceOptions
	handlePath    string
	componentName string
	shutdown      bool
}

func (o *robotOptions) gatherOptions(fs *flag.FlagSet, args ...string) *configuration {

	o.service.AddFlags(fs)
	fs.StringVar(
		&o.handlePath, "handle-path", "webhook",
		"http server handle interface path",
	)
	fs.StringVar(
		&o.handlePath, "component-name", "robot-generic-access",
		"logging field to flag project",
	)

	_ = fs.Parse(args)

	if err := o.service.Validate(); err != nil {
		logrus.Errorf("invalid service options, err:%s", err.Error())
		o.shutdown = true
	}
	configmap := config.NewConfigmapAgent(&configuration{})
	if err := configmap.Load(o.service.ConfigFile); err != nil {
		logrus.Errorf("load config, err:%s", err.Error())
		return nil
	}

	return configmap.GetConfigmap().(*configuration)
}

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
	startup(bot, opt)
}

func startup(d *robot, o *robotOptions) {
	defer interrupts.WaitForGracefulShutdown()

	// Return 200 on / for health checks.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	// For /**-hook, handle a webhook normally.
	http.Handle("/"+o.handlePath, d)

	httpServer := &http.Server{Addr: ":" + strconv.Itoa(o.service.Port)}

	interrupts.ListenAndServe(httpServer, o.service.GracePeriod)
}

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
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/server-common-lib/options"
	"github.com/sirupsen/logrus"
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
		"http server handle's restapi path",
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

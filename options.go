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
	"github.com/sirupsen/logrus"
)

type robotOptions struct {
	service   config.FrameworkOptions
	interrupt bool
}

func (o *robotOptions) gatherOptions(fs *flag.FlagSet, args ...string) *configuration {

	o.service.AddFlagsComposite(fs)

	_ = fs.Parse(args)

	if err := o.service.ValidateComposite(); err != nil {
		logrus.WithError(err).Error("invalid service startup arguments")
		o.interrupt = true
		return nil
	}
	configmap, err := config.NewConfigmapAgent(&configuration{}, o.service.ConfigFile)
	if err != nil {
		logrus.WithError(err).Error("invalid item exists in the configmap")
		return nil
	}

	return configmap.GetConfigmap().(*configuration)
}

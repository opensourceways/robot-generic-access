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
	"github.com/go-resty/resty/v2"
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/robot-framework-lib/utils"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync"
)

const (
	missingEventTypeErrorMessage = "400 Bad Request: Missing Event Type Header"
	noBodyErrorMessage           = "400 Bad Request: request body should be ono-nil"
	noOrgErrorMessage            = "400 Bad Request: request body not contain owner"
	noRepoErrorMessage           = "400 Bad Request: request body not contain repo"
)

func newRobot(c *configuration) *robot {
	logger := framework.NewLogger().WithField("component", component)
	return &robot{client: resty.New().RemoveProxy().SetRetryCount(3).SetLogger(logger), configmap: c, log: logger}
}

type robot struct {
	client    *resty.Client
	configmap *configuration
	event     *client.GenericEvent
	log       *logrus.Entry
	wg        sync.WaitGroup
}

func (bot *robot) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	evt := client.NewGenericEvent(w, r, bot.log)
	if utils.GetString(evt.EventType) == "" {
		bot.log.Warning(missingEventTypeErrorMessage)
		http.Error(w, missingEventTypeErrorMessage, http.StatusBadRequest)
		return
	}

	if evt.GetMetaPayload() == nil {
		bot.log.Warning(noBodyErrorMessage)
		http.Error(w, noBodyErrorMessage, http.StatusBadRequest)
		return
	}

	if utils.GetString(evt.Org) == "" {
		bot.log.Warning(noOrgErrorMessage)
		http.Error(w, noOrgErrorMessage, http.StatusBadRequest)
		return
	}

	if utils.GetString(evt.Repo) == "" {
		bot.log.Warning(noRepoErrorMessage)
		http.Error(w, noRepoErrorMessage, http.StatusBadRequest)
		return
	}

	bot.event = evt
	endpoints := bot.configmap.GetEndpoints(*evt.Org, *evt.Repo, *evt.EventType)
	if len(endpoints) == 0 {
		return
	}

	bot.wg.Add(1)
	r.Header.Set(client.HeaderRobotChain, client.HeaderRobotChainAuthed)
	go bot.dispatcher(&r.Header, endpoints)
}

func (bot *robot) wait() {
	bot.wg.Wait() // Handle remaining requests
}

func (bot *robot) dispatcher(h *http.Header, endpoints []string) {
	defer bot.wg.Done()
	logger := bot.log.WithFields(*bot.event.CollectLoggingFields())
	for _, uri := range endpoints {

		req := bot.client.R()
		req.Header = *h
		req.SetBody(bot.event)

		resp, err := req.Post(uri)
		if err != nil {
			logger.Errorf("failed to send to %s , reason: %v", uri, err)
			return
		}
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.RawBody())
		}
		logger.Infof("successful to send to %s ", uri)
	}

}

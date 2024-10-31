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
	return &robot{client: resty.New().RemoveProxy().SetRetryCount(3), configmap: c}
}

type robot struct {
	client    *resty.Client
	configmap *configuration
	event     *client.GenericEvent
	wg        sync.WaitGroup
}

func (bot *robot) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	evt := client.NewGenericEvent(w, r)
	if evt.EventType == nil || *evt.EventType == "" {
		http.Error(w, missingEventTypeErrorMessage, http.StatusBadRequest)
		return
	}

	if evt.MetaPayload == nil {
		http.Error(w, noBodyErrorMessage, http.StatusBadRequest)
		return
	}

	if evt.Org == nil || *evt.Org == "" {
		http.Error(w, noOrgErrorMessage, http.StatusBadRequest)
		return
	}

	if evt.Repo == nil || *evt.Repo == "" {
		http.Error(w, noRepoErrorMessage, http.StatusBadRequest)
		return
	}

	bot.event = evt
	endpoints := bot.configmap.GetEndpoints(*evt.Org, *evt.Repo, *evt.EventType)
	if len(endpoints) == 0 {
		return
	}
	bot.dispatcher(&r.Header, endpoints)
}

func (bot *robot) Wait() {
	bot.wg.Wait() // Handle remaining requests
}

func (bot *robot) dispatcher(h *http.Header, endpoints []string) {

	lgr := logrus.NewEntry(logrus.StandardLogger()).WithFields(*bot.event.CollectLoggingFields())
	for _, uri := range endpoints {

		req := bot.client.R()
		req.Header = *h
		req.Body = bot.event.MetaPayload.Bytes()

		bot.wg.Add(1)
		go func(urlStr string) {
			defer bot.wg.Done()
			resp, err := req.Post(urlStr)
			if err != nil {
				lgr.Errorf("failed to send to %s , reason: %v", urlStr, err)
				return
			}
			if resp != nil {
				_, _ = io.Copy(io.Discard, resp.RawBody())
			}
			lgr.Infof("successful to send to %s ", urlStr)
		}(uri)
	}

}

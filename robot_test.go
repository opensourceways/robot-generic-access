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
	"bytes"
	"encoding/json"
	"flag"
	"github.com/opensourceways/server-common-lib/interrupts"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	headerEventType            = "X-GitCode-Event"
	headerEventTypeValue       = "Note Hook"
	headerEventGUID            = "X-GitCode-Delivery"
	headerEventGUIDValue       = "607294e6-0e65-417f-a42b-3edd70de32ee"
	headerUserAgent            = "User-Agent"
	headerUserAgentValue       = "git-gitcode-hook"
	headerContentTypeName      = "Content-Type"
	headerContentTypeJsonValue = "application/json"
	headerRobotChain           = "Robot-Chain"
	headerRobotChainAuthed     = "Request-Authenticated"

	mockRequestBody = "********************"
)

func TestDispatcherSuccess(t *testing.T) {
	args := []string{
		"***",
		"--port=8511",
		"--config-file=" + findTestdata(t, "testdata"+string(os.PathSeparator)+"config7.yaml"),
		"--handle-path=gitcode-hook",
	}

	opt := new(robotOptions)
	cnf := opt.gatherOptions(flag.NewFlagSet(args[0], flag.ExitOnError), args[1:]...)
	bot := newRobot(cnf)

	exitChannel := make(chan int)
	http.HandleFunc("/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentTypeName, headerContentTypeJsonValue)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentTypeName, headerContentTypeJsonValue)
		w.WriteHeader(http.StatusBadGateway)
		exitChannel <- 1
	})

	server := &http.Server{}

	servChannel := make(chan int)
	go func() {
		defer interrupts.WaitForGracefulShutdown()
		server.Addr = "localhost:18818"
		servChannel <- 1
		interrupts.ListenAndServe(server, time.Second)
	}()
	<-servChannel
	data, _ := os.ReadFile(findTestdata(t, "testdata"+string(os.PathSeparator)+"pr_note.json"))
	buf := &bytes.Buffer{}
	buf.Write(data)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/case1", buf)
	req.Header.Set(headerUserAgent, headerUserAgentValue)
	req.Header.Set(headerContentTypeName, headerContentTypeJsonValue)
	req.Header.Set(headerEventType, headerEventTypeValue)
	req.Header.Set(headerEventGUID, headerEventGUIDValue)
	bot.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	<-exitChannel
	interrupts.OnInterrupt(func() {
		t.Log("stop server")
	})
}

func TestDispatcherFail(t *testing.T) {
	args := []string{
		"***",
		"--port=8511",
		"--config-file=" + findTestdata(t, "testdata"+string(os.PathSeparator)+"config8.yaml"),
		"--handle-path=gitcode-hook",
	}

	opt := new(robotOptions)
	cnf := opt.gatherOptions(flag.NewFlagSet(args[0], flag.ExitOnError), args[1:]...)
	bot := newRobot(cnf)

	data, _ := os.ReadFile(findTestdata(t, "testdata"+string(os.PathSeparator)+"pr_note.json"))
	buf := &bytes.Buffer{}
	buf.Write(data)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/case2", buf)
	req.Header.Set(headerUserAgent, headerUserAgentValue)
	req.Header.Set(headerContentTypeName, headerContentTypeJsonValue)
	req.Header.Set(headerEventType, headerEventTypeValue)
	req.Header.Set(headerEventGUID, headerEventGUIDValue)
	bot.ServeHTTP(w, req)

	buf1 := &bytes.Buffer{}
	err := json.NewEncoder(buf1).Encode(bot.event)
	assert.Equal(t, nil, err)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/case3", buf1)
	req1.Header.Set(headerContentTypeName, headerContentTypeJsonValue)
	req1.Header.Set(headerEventType, headerEventTypeValue)
	req1.Header.Set(headerEventGUID, headerEventGUIDValue)
	req1.Header.Set(headerRobotChain, headerRobotChainAuthed)
	bot.ServeHTTP(w1, req1)

	bot.event.Repo = nil
	err = json.NewEncoder(buf1).Encode(bot.event)
	assert.Equal(t, nil, err)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/case5", buf1)
	req3.Header.Set(headerContentTypeName, headerContentTypeJsonValue)
	req3.Header.Set(headerEventType, headerEventTypeValue)
	req3.Header.Set(headerEventGUID, headerEventGUIDValue)
	req3.Header.Set(headerRobotChain, headerRobotChainAuthed)
	bot.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Result().StatusCode)
	var str strings.Builder
	_, _ = io.Copy(&str, w3.Result().Body)
	assert.Equal(t, noRepoErrorMessage+"\n", str.String())

	bot.event.Org = nil
	err = json.NewEncoder(buf1).Encode(bot.event)
	assert.Equal(t, nil, err)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/case5", buf1)
	req4.Header.Set(headerContentTypeName, headerContentTypeJsonValue)
	req4.Header.Set(headerEventType, headerEventTypeValue)
	req4.Header.Set(headerEventGUID, headerEventGUIDValue)
	req4.Header.Set(headerRobotChain, headerRobotChainAuthed)
	bot.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusBadRequest, w4.Result().StatusCode)
	var str2 strings.Builder
	_, _ = io.Copy(&str2, w4.Result().Body)
	assert.Equal(t, noOrgErrorMessage+"\n", str2.String())

	bot.event.EventType = nil
	err = json.NewEncoder(buf1).Encode(bot.event)
	assert.Equal(t, nil, err)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/case4", buf1)
	req2.Header.Set(headerContentTypeName, headerContentTypeJsonValue)
	req2.Header.Set(headerEventType, headerEventTypeValue)
	req2.Header.Set(headerEventGUID, headerEventGUIDValue)
	req2.Header.Set("Robot-Chain", "Request-Authenticated")
	bot.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Result().StatusCode)
	var str1 strings.Builder
	_, _ = io.Copy(&str1, w2.Result().Body)
	assert.Equal(t, missingEventTypeErrorMessage+"\n", str1.String())
}

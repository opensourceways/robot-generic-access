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
	"errors"
	"fmt"
	"k8s.io/utils/set"
	"net/url"
	"slices"
)

type configuration struct {
	ConfigItems accessConfig `json:"access,omitempty"`
}

func (c *configuration) Validate() error {
	return c.ConfigItems.validate()
}

type accessConfig struct {
	// Plugins is a map of repositories (eg "k/k") to lists of plugin names.
	RepoPlugins map[string][]string `json:"repo_plugins,omitempty"`

	// Plugins is a list available plugins.
	Plugins []pluginConfig `json:"plugins,omitempty"`
}

type pluginConfig struct {
	// Name of the plugin.
	Name string `json:"name" required:"true"`

	// Endpoint is the location of the plugin.
	Endpoint string `json:"endpoint" required:"true"`

	// Events are the events that this plugin can handle and should be forward to it.
	// If no events are specified, everything is sent.
	Events []string `json:"events,omitempty"`
}

func (a *accessConfig) validate() error {
	var botSet = set.Set[string]{}
	for i := range a.Plugins {
		if err := a.Plugins[i].validate(); err != nil {
			return err
		}
		botSet.Insert(a.Plugins[i].Name)
	}

	var e []string
	for _, item := range a.RepoPlugins {
		for _, value := range item {
			if !botSet.Has(value) {
				e = append(e, value)
			}
		}
	}

	if len(e) > 0 {
		return fmt.Errorf("repo_plugins %v missing plugins in the configmap", e)
	}

	return nil
}

func (p *pluginConfig) validate() error {
	if p.Name == "" {
		return errors.New("plugin missing name")
	}

	if p.Endpoint == "" {
		return errors.New(p.Name + " plugin missing endpoint")
	}

	if _, err := url.ParseRequestURI(p.Endpoint); err != nil {
		return errors.New(p.Endpoint + " not a valid url")
	}

	return nil
}

func (c *configuration) GetEndpoints(org, repo, eventType string) []string {
	var ans []string

	if c.ConfigItems.RepoPlugins == nil {
		return ans
	}

	var servers []string
	endpoint, ok := c.ConfigItems.RepoPlugins[org]
	if ok {
		servers = append(servers, endpoint...)
	}

	endpoint, ok = c.ConfigItems.RepoPlugins[org+"/"+repo]
	if ok {
		servers = append(servers, endpoint...)
	}

	if len(c.ConfigItems.Plugins) != 0 && len(servers) != 0 {
		ans = matchEndpoint(&c.ConfigItems.Plugins, eventType, servers...)
	}

	return ans
}

func matchEndpoint(m *[]pluginConfig, event string, robotNames ...string) (ans []string) {
	for _, val := range robotNames {
		for _, value := range *m {
			if value.Name == val && slices.Contains(value.Events, event) {
				ans = append(ans, value.Endpoint)
			}
		}
	}

	return
}

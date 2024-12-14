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
	"github.com/opensourceways/server-common-lib/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {

	type args struct {
		cnf  *configuration
		path string
	}

	testCases := []struct {
		no  string
		in  args
		out []error
	}{
		{
			"case0",
			args{
				&configuration{},
				"",
			},
			[]error{nil, nil},
		},
		{
			"case1",
			args{
				&configuration{
					ConfigItems: accessConfig{
						RepoPlugins: make(map[string][]string),
					},
				},
				"",
			},
			[]error{nil, nil},
		},
		{
			"case2",
			args{
				&configuration{
					ConfigItems: accessConfig{
						Plugins: []pluginConfig{},
					},
				},
				"",
			},
			[]error{nil, nil},
		},
		{
			"case3",
			args{
				&configuration{},
				"config2.yaml",
			},
			[]error{nil, errors.New("repo_plugins [serv1] missing plugins in the configmap")},
		},
		{
			"case4",
			args{
				&configuration{},
				"config3.yaml",
			},
			[]error{nil, errors.New("41231 not a valid url")},
		},
		{
			"case5",
			args{
				&configuration{},
				"config4.yaml",
			},
			[]error{nil, errors.New("plugin missing name")},
		},
		{
			"case6",
			args{
				&configuration{},
				"config5.yaml",
			},
			[]error{nil, errors.New("serv2 plugin missing endpoint")},
		},
		{
			"case7",
			args{
				&configuration{},
				"config.yaml",
			},
			[]error{nil, nil},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].no, func(t *testing.T) {
			if testCases[i].in.path != "" {
				err := utils.LoadFromYaml(findTestdata(t, testCases[i].in.path), testCases[i].in.cnf)
				assert.Equal(t, testCases[i].out[0], err)
			}

			err1 := testCases[i].in.cnf.Validate()
			assert.Equal(t, testCases[i].out[1], err1)
		})
	}

}

func TestGetEndpoints(t *testing.T) {
	type args struct {
		cnf       *configuration
		path      string
		org       string
		repo      string
		eventType string
	}

	testCases := []struct {
		no  string
		in  args
		out []string
	}{
		{
			"case0",
			args{
				&configuration{},
				"",
				"1",
				"2",
				"3",
			},
			([]string)(nil),
		},
		{
			"case1",
			args{
				&configuration{},
				"config.yaml",
				"1",
				"2",
				"3",
			},
			([]string)(nil),
		},
		{
			"case2",
			args{
				&configuration{},
				"config.yaml",
				"ibforuorg",
				"2",
				"Note Hook",
			},
			([]string)(nil),
		},
		{
			"case3",
			args{
				&configuration{},
				"config6.yaml",
				"org1",
				"2",
				"Push Hook",
			},
			([]string)(nil),
		},
		{
			"case4",
			args{
				&configuration{},
				"config6.yaml",
				"org1",
				"test1",
				"Issue Hook",
			},
			[]string{"http://localhost:7000/gitcode-hook2"},
		},
		{
			"case5",
			args{
				&configuration{},
				"config.yaml",
				"ibforuorg",
				"test1",
				"Issue Hook",
			},
			[]string{"http://localhost:7000/gitcode-hook", "http://localhost:7000/gitcode-hook2"},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].no, func(t *testing.T) {
			if testCases[i].in.path != "" {
				_ = utils.LoadFromYaml(findTestdata(t, testCases[i].in.path), testCases[i].in.cnf)
			}

			endpoints := testCases[i].in.cnf.GetEndpoints(testCases[i].in.org, testCases[i].in.repo, testCases[i].in.eventType)
			assert.Equal(t, testCases[i].out, endpoints)
		})
	}
}

func findTestdata(t *testing.T, path string) string {
	path = "testdata" + string(os.PathSeparator) + path
	i := 0
retry:
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Error(path + " not found")
		return ""
	}
	if _, err = os.Stat(absPath); !os.IsNotExist(err) {
		return absPath
	} else {
		i++
		path = ".." + string(os.PathSeparator) + path
		if i <= 3 {
			goto retry
		}
	}

	t.Log(path + " not found")
	return ""
}

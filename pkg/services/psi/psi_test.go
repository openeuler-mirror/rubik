// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jingxiao Lu
// Date: 2023-06-12
// Description: This file is used for testing psi.go

package psi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/core/typedef"
)

// TestNewManagerObj tests NewObj() for Factory
func TestNewManagerObj(t *testing.T) {
	var fact = Factory{
		ObjName: "psi",
	}
	nm, err := fact.NewObj()
	if err != nil {
		t.Fatalf("New PSI Manager failed: %v", err)
		return
	}
	fmt.Printf("New PSI Manager %s is %#v", fact.Name(), nm)
}

// TestConfigValidate tests Config Validate
func TestConfigValidate(t *testing.T) {
	var tests = []struct {
		name    string
		conf    *Config
		wantErr bool
	}{
		{
			name:    "TC1 - Default Config",
			conf:    NewConfig(),
			wantErr: true,
		},
		{
			name: "TC2 - Wrong Interval value",
			conf: &Config{
				Interval: minInterval - 1,
			},
			wantErr: true,
		},
		{
			name: "TC3 - Wrong Threshold value",
			conf: &Config{
				Interval:       minInterval,
				Avg10Threshold: minThreshold - 1,
			},
			wantErr: true,
		},
		{
			name: "TC4 - No resource type specified",
			conf: &Config{
				Interval:       minInterval,
				Avg10Threshold: minThreshold,
			},
			wantErr: true,
		},
		{
			name: "TC5 - Wrong resource type cpuacct - cpuacct is for psi subsystem, not for resource type",
			conf: &Config{
				Interval:       minInterval,
				Avg10Threshold: minThreshold,
				Resource:       []string{"cpu", "memory", "io", "cpuacct"},
			},
			wantErr: true,
		},
		{
			name: "TC6 - Success case - trully end",
			conf: &Config{
				Interval:       minInterval,
				Avg10Threshold: minThreshold,
				Resource:       []string{"cpu", "memory", "io"},
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.conf.Validate(); (err != nil) != tc.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

type FakeManager struct{}

func (m *FakeManager) ListContainersWithOptions(options ...api.ListOption) map[string]*typedef.ContainerInfo {
	return make(map[string]*typedef.ContainerInfo)
}
func (m *FakeManager) ListPodsWithOptions(options ...api.ListOption) map[string]*typedef.PodInfo {
	return make(map[string]*typedef.PodInfo, 1)
}

// TestManagerRun creates a fake manager and runs it
func TestManagerRun(t *testing.T) {
	nm := NewManager("psi")
	nm.conf.Interval = 1
	nm.PreStart(&FakeManager{})
	nm.SetConfig(func(configName string, d interface{}) error { return nil })
	if !nm.IsRunner() {
		t.Fatalf("FakeManager is not a runner!")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go nm.Run(ctx)
	time.Sleep(time.Second)
	cancel()
}

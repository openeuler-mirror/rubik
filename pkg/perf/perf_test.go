// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2022-02-09
// Description: cgroup perf stats testcase

// Package perf provide perf functions
package perf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
)

// TestCgroupStat testcase
func TestCgroupStat(t *testing.T) {
	if !hwSupport {
		t.Skipf("%s only run on physical machine", t.Name())
	}
	testCGRoot := filepath.Join(config.CgroupRoot, "perf_event", t.Name())
	assert.NoError(t, os.MkdirAll(testCGRoot, constant.DefaultDirMode))
	ioutil.WriteFile(filepath.Join(testCGRoot, "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
	defer func() {
		ioutil.WriteFile(filepath.Join(config.CgroupRoot, "perf_event", "tasks"), []byte(fmt.Sprint(os.Getpid())), constant.DefaultFileMode)
		assert.NoError(t, os.Remove(testCGRoot))
	}()
	type args struct {
		cgpath string
		dur    time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *PerfStat
		wantErr bool
	}{
		{
			name: "TC-normal",
			args: args{cgpath: testCGRoot, dur: time.Second},
		},
		{
			name:    "TC-empty cgroup path",
			args:    args{cgpath: "", dur: time.Second},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CgroupStat(tt.args.cgpath, tt.args.dur)
			if (err != nil) != tt.wantErr {
				t.Errorf("CgroupStat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

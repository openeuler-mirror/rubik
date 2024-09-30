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
	"os"
	"path/filepath"
	"testing"
	"time"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/tests/try"
)

var testConf []int = []int{INSTRUCTIONS, CYCLES, CACHEREFERENCES, CACHEMISS, LLCMISS, LLCACCESS}

// TestCgroupStat testcase
func TestCgroupStat(t *testing.T) {
	if !Support() {
		t.Skipf("%s only run on physical machine", t.Name())
	}
	testCGRoot := cgroup.AbsoluteCgroupPath("perf_event", t.Name(), "")
	try.MkdirAll(testCGRoot, constant.DefaultDirMode)
	try.WriteFile(filepath.Join(testCGRoot, "tasks"), fmt.Sprint(os.Getpid()))
	defer func() {
		try.WriteFile(cgroup.AbsoluteCgroupPath("perf_event", "tasks", ""), fmt.Sprint(os.Getpid()))
		try.RemoveAll(testCGRoot)
	}()
	type args struct {
		cgpath string
		dur    time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *Stat
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
			_, err := CgroupStat(tt.args.cgpath, tt.args.dur, testConf)
			if (err != nil) != tt.wantErr {
				t.Errorf("CgroupStat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

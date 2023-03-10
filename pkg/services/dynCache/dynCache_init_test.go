// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2023-02-21
// Description: This file will init cache limit directories before services running

// Package cachelimit is the service used for cache limit setting
package dynCache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/test/try"
)

func setMaskFile(t *testing.T, resctrlDir string, data string) {
	maskDir := filepath.Join(resctrlDir, "info", "L3")
	maskFile := filepath.Join(maskDir, "cbm_mask")
	try.MkdirAll(maskDir, constant.DefaultDirMode).OrDie()
	try.WriteFile(maskFile, data).OrDie()
	try.WriteFile(filepath.Join(resctrlDir, schemataFileName), "L3:0=7fff;1=7fff;2=7fff;3=7fff\nMB:0=100;1=100;2=100;3=100").OrDie()
}

func genNumaNodes(path string, num int) {
	for i := 0; i < num; i++ {
		try.MkdirAll(filepath.Join(path, fmt.Sprintf("node%d", i)), constant.DefaultDirMode).OrDie()
	}
}

func TestCacheLimit_InitCacheLimitDir(t *testing.T) {
	if !perf.Support() {
		t.Skipf("%s only run on physical machine", t.Name())
	}
	type fields struct {
		Config *Config
		Attr   *Attr
		Name   string
	}
	// defaultConfig := genDefaultConfig()
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		preHook  func(t *testing.T, c *DynCache)
		postHook func(t *testing.T, c *DynCache)
	}{
		{
			name: "TC1-normal cache limit dir setting",
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				c.Config.DefaultResctrlDir = try.GenTestDir().String()
				c.Config.DefaultLimitMode = modeStatic
				setMaskFile(t, c.Config.DefaultResctrlDir, "7fff")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 4)
			},
			postHook: func(t *testing.T, c *DynCache) {
				resctrlLevelMap := map[string]string{
					"rubik_max":     "L3:0=7fff;1=7fff;2=7fff;3=7fff\nMB:0=100;1=100;2=100;3=100\n",
					"rubik_high":    "L3:0=7f;1=7f;2=7f;3=7f\nMB:0=50;1=50;2=50;3=50\n",
					"rubik_middle":  "L3:0=f;1=f;2=f;3=f\nMB:0=30;1=30;2=30;3=30\n",
					"rubik_low":     "L3:0=7;1=7;2=7;3=7\nMB:0=10;1=10;2=10;3=10\n",
					"rubik_dynamic": "L3:0=7;1=7;2=7;3=7\nMB:0=10;1=10;2=10;3=10\n",
				}
				for level, expect := range resctrlLevelMap {
					schemataFile := filepath.Join(c.DefaultResctrlDir, level, schemataFileName)
					content := try.ReadFile(schemataFile).String()
					assert.Equal(t, expect, content)
				}
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
		{
			name:    "TC2-not share with host pid namespace",
			wantErr: true,
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				pidNameSpaceDir := try.GenTestDir().String()
				pidNameSpaceFileOri := try.WriteFile(filepath.Join(pidNameSpaceDir, "pid:[4026531836x]"), "")
				pidNameSpace := filepath.Join(pidNameSpaceDir, "pid")

				os.Symlink(pidNameSpaceFileOri.String(), pidNameSpace)
				c.Config.DefaultPidNameSpace = pidNameSpace
			},
			postHook: func(t *testing.T, c *DynCache) {
				try.RemoveAll(filepath.Dir(c.DefaultPidNameSpace))
			},
		},
		{
			name:    "TC3-pid namespace file is not link file",
			wantErr: true,
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				pidNameSpaceDir := try.GenTestDir().String()
				pidNameSpaceFileOri := try.WriteFile(filepath.Join(pidNameSpaceDir, "pid:[4026531836x]"), "")
				pidNameSpace := filepath.Join(pidNameSpaceDir, "pid")

				os.Link(pidNameSpaceFileOri.String(), pidNameSpace)
				c.Config.DefaultPidNameSpace = pidNameSpace
			},
			postHook: func(t *testing.T, c *DynCache) {
				try.RemoveAll(filepath.Dir(c.DefaultPidNameSpace))
			},
		},
		{
			name:    "TC4-resctrl path not exist",
			wantErr: true,
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				c.Config.DefaultResctrlDir = "/resctrl/path/is/not/exist"
			},
		},
		{
			name:    "TC5-resctrl schemata file not exist",
			wantErr: true,
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				c.Config.DefaultResctrlDir = try.GenTestDir().String()
			},
			postHook: func(t *testing.T, c *DynCache) {
				try.RemoveAll(c.DefaultResctrlDir)
			},
		},
		{
			name:    "TC6-no numa path",
			wantErr: true,
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				c.Attr.NumaNodeDir = "/numa/node/path/is/not/exist"
			},
		},
		{
			name:    "TC7-empty cbm mask file",
			wantErr: true,
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				c.Config.DefaultResctrlDir = try.GenTestDir().String()
				c.Config.DefaultLimitMode = modeStatic
				setMaskFile(t, c.Config.DefaultResctrlDir, "")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 0)
			},
			postHook: func(t *testing.T, c *DynCache) {
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
		{
			name: "TC8-low cmb mask value",
			fields: fields{
				Config: genDefaultConfig(),
				Attr:   &Attr{},
			},
			preHook: func(t *testing.T, c *DynCache) {
				c.Config.DefaultResctrlDir = try.GenTestDir().String()
				c.Config.DefaultLimitMode = modeStatic
				setMaskFile(t, c.Config.DefaultResctrlDir, "1")
				numaNodeDir := try.GenTestDir().String()
				c.Attr.NumaNodeDir = numaNodeDir
				genNumaNodes(c.Attr.NumaNodeDir, 0)
			},
			postHook: func(t *testing.T, c *DynCache) {
				try.RemoveAll(c.Config.DefaultResctrlDir)
				try.RemoveAll(c.Attr.NumaNodeDir)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DynCache{
				Config: tt.fields.Config,
				Attr:   tt.fields.Attr,
				Name:   tt.fields.Name,
			}
			if tt.preHook != nil {
				tt.preHook(t, c)
			}
			if err := c.InitCacheLimitDir(); (err != nil) != tt.wantErr {
				t.Errorf("CacheLimit.InitCacheLimitDir() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook(t, c)
			}
		})
	}
}

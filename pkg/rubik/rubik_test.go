// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2021-05-20
// Description: This file is used for rubik package test

package rubik

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/httpserver"
	"isula.org/rubik/pkg/util"
	"isula.org/rubik/pkg/workerpool"
)

// TestNewRubik is NewRubik function test
func TestNewRubik(t *testing.T) {
	os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	defer os.RemoveAll(constant.TmpTestDir)
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	os.Remove(tmpConfigFile)

	os.MkdirAll(tmpConfigFile, constant.DefaultDirMode)
	_, err := NewRubik(tmpConfigFile)
	assert.Equal(t, true, err != nil)
	err = os.RemoveAll(tmpConfigFile)
	assert.NoError(t, err)

	fd, err := os.Create(tmpConfigFile)
	assert.NoError(t, err)
	_, err = fd.WriteString(`{
						"autoCheck": false,
						"logDriver": "file",
						"logDir": "/tmp/rubik-test",
						"logSize": 2048,
						"logLevel": "debug",
						"cgroupRoot": "/tmp/rubik-test/cgroup"
}`)
	assert.NoError(t, err)
	_, err = NewRubik(tmpConfigFile)
	assert.NoError(t, err)

	os.Remove(tmpConfigFile)
	fd, err = os.Create(tmpConfigFile)
	assert.NoError(t, err)
	_, err = fd.WriteString(`{
						"logLevel": "debugabc"
}`)
	assert.NoError(t, err)
	_, err = NewRubik(tmpConfigFile)
	assert.Equal(t, true, err != nil)

	fd.Close()
}

// TestMonitor is Monitor function test
func TestMonitor(t *testing.T) {
	server, _ := httpserver.NewServer()
	rubik := &Rubik{
		server: server,
		pool: &workerpool.WorkerPool{
			WorkerNum:  1,
			WorkerBusy: 1,
		},
	}

	go rubik.Monitor()
	close(config.ShutdownChan)
}

// TestShutdown is Shutdown function test
func TestShutdown(t *testing.T) {
	server, _ := httpserver.NewServer()
	rubik := &Rubik{
		server: server,
		pool: &workerpool.WorkerPool{
			WorkerNum:  1,
			WorkerBusy: 0,
		},
	}

	rubik.Shutdown()
}

// TestSync is Sync function test
func TestSync(t *testing.T) {
	rubik := &Rubik{
		config: &config.Config{
			AutoCheck: true,
		},
	}

	err := rubik.Sync()
	assert.Equal(t, true, err != nil)
}

// TestServe is Serve function test
func TestServe(t *testing.T) {
	sock, err := httpserver.NewSock()
	assert.NoError(t, err)
	server, _ := httpserver.NewServer()
	rubik := &Rubik{
		server: server,
		sock:   sock,
		config: &config.Config{},
	}

	var errC chan error
	go func() {
		errC <- rubik.Serve()
	}()

	select {
	case err = <-errC:
	case <-time.After(time.Second):
		err = nil
	}
	assert.NoError(t, err)
}

var cfgA = `
{
	"autoCheck": true,
	"logDriver": "file",
	"logDir": "/tmp/rubik-test",
	"logSize": 2048,
	"logLevel": "debug",
	"cgroupRoot": "/tmp/rubik-test/cgroup"
}`

// TestRunAbnormality test run server abnormality
func TestRunAbnormality(t *testing.T) {
	old := os.Args
	defer func() {
		os.Args = old
	}()
	configFile := "config.json"
	fcfg := filepath.Join(constant.TmpTestDir, configFile)
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	if err != nil {
		assert.NoError(t, err)
	}
	err = ioutil.WriteFile(fcfg, []byte(cfgA), constant.DefaultFileMode)
	if err != nil {
		assert.NoError(t, err)
	}
	// case: argument not valid
	os.Args = []string{"invalid", "failed"}
	ret := Run(fcfg)
	assert.Equal(t, constant.ErrCodeFailed, ret)
	os.Args = []string{"rubik"}
	// case: file is locked
	lock, err := util.CreateLockFile(constant.LockFile)
	ret = Run(fcfg) // set rubik lock failed: ...
	assert.Equal(t, constant.ErrCodeFailed, ret)
	util.RemoveLockFile(lock, constant.LockFile)
	// case: invalid config.json
	err = ioutil.WriteFile(fcfg, []byte("invalid"), constant.DefaultFileMode)
	if err != nil {
		assert.NoError(t, err)
	}
	ret = Run(fcfg)
	assert.Equal(t, constant.ErrCodeFailed, ret)
}

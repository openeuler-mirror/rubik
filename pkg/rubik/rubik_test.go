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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"isula.org/rubik/pkg/checkpoint"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/util"
)

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
	_, err = os.Create(tmpConfigFile)
	assert.NoError(t, err)

	err = reCreateConfigFile(tmpConfigFile, `{
						"autoCheck": true,
						"logDriver": "file",
						"logDir": "/tmp/rubik-test",
						"logSize": 2048,
						"logLevel": "debug",
						"cgroupRoot": "/tmp/rubik-test/cgroup"
}`)
	assert.NoError(t, err)

	err = reCreateConfigFile(tmpConfigFile, `{
						"logLevel": "debugabc"
}`)
	assert.Equal(t, true, err != nil)

	err = reCreateConfigFile(tmpConfigFile, `{
						"autoConfig": true
}`)
	assert.Equal(t, true, strings.Contains(err.Error(), "must be defined"))

	err = reCreateConfigFile(tmpConfigFile, `{
						"autoConfig": true
}`)
	assert.Equal(t, true, strings.Contains(err.Error(), "must be defined"))

}

func reCreateConfigFile(path, content string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.WriteString(content)
	if err != nil {
		return err
	}
	_, err = NewRubik(path)
	if err != nil {
		return err
	}

	return nil
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

// TestRun test run server
func TestRun(t *testing.T) {
	if os.Getenv("BE_TESTRUN") == "1" {
		// case: config.json missing, use default config.json
		ret := Run("/dev/should/not/exist")
		fmt.Println(ret)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestRun")
	cmd.Env = append(os.Environ(), "BE_TESTRUN=1")
	err := cmd.Start()
	assert.NoError(t, err)
	sleepTime := 3
	time.Sleep(time.Duration(sleepTime) * time.Second)
	err = cmd.Process.Signal(syscall.SIGINT)
	assert.NoError(t, err)
}

// TestCacheLimit is CacheLimit function test
func TestCacheLimit(t *testing.T) {
	rubik := &Rubik{
		config: &config.Config{
			CacheCfg: config.CacheConfig{
				Enable:            true,
				DefaultLimitMode:  "invalid",
				DefaultResctrlDir: constant.TmpTestDir + "invalid",
			},
		},
	}

	err := rubik.CacheLimit()
	assert.Equal(t, true, err != nil)
	rubik.config.CacheCfg.Enable = false
	err = rubik.CacheLimit()
	assert.NoError(t, err)
}

// TestSmallRun test run function
func TestSmallRun(t *testing.T) {
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	os.Remove(tmpConfigFile)
	err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)
	assert.NoError(t, err)
	fd, err := os.Create(tmpConfigFile)
	assert.NoError(t, err)
	fd.WriteString(`{
						"cacheConfig": {
                            "enable": true,
                            "defaultLimitMode": "invalid"
						}
}`)
	assert.NoError(t, err)
	err = fd.Close()
	assert.NoError(t, err)

	res := run(tmpConfigFile)
	assert.Equal(t, constant.ErrCodeFailed, res)
}

// TestInitKubeClient test initKubeClient
func TestInitKubeClient(t *testing.T) {
	r := &Rubik{config: &config.Config{AutoConfig: true}}
	err := r.initKubeClient()
	assert.Equal(t, true, strings.Contains(err.Error(), "must be defined"))
}

// TestInitEventHandler test initEventHandler
func TestInitEventHandler(t *testing.T) {
	r := &Rubik{config: &config.Config{AutoConfig: true}}
	err := r.initEventHandler()
	assert.Equal(t, true, strings.Contains(err.Error(), "kube-client is not initialized"))

	r.kubeClient = &kubernetes.Clientset{}
	err = r.initEventHandler()
	assert.Equal(t, true, strings.Contains(err.Error(), "must be defined"))
}

// TestInitCheckpoint test initCheckpoint
func TestInitCheckpoint(t *testing.T) {
	r := &Rubik{config: &config.Config{AutoConfig: true}}
	err := r.initCheckpoint()
	assert.Equal(t, true, strings.Contains(err.Error(), "kube-client not initialized"))

	r.kubeClient = &kubernetes.Clientset{}
	err = r.initCheckpoint()
	assert.Equal(t, true, strings.Contains(err.Error(), "missing"))
}

// TestAddUpdateDelEvent test Event
func TestAddUpdateDelEvent(t *testing.T) {
	r, err := NewRubik("")
	assert.NoError(t, err)
	cpm := checkpoint.NewManager()
	r.cpm = cpm
	oldPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:  "aaa",
			Name: "podaaa",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}
	r.AddEvent(oldPod)
	oldPod.Status.Phase = corev1.PodRunning
	r.AddEvent(oldPod)
	assert.Equal(t, "podaaa", r.cpm.Checkpoint.Pods["aaa"].Name)

	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:  "aaa",
			Name: "podbbb",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}
	r.UpdateEvent(oldPod, newPod)
	_, ok := r.cpm.Checkpoint.Pods["aaa"]
	assert.Equal(t, false, ok)

	r.AddEvent(oldPod)
	newPod.Status.Phase = corev1.PodRunning
	r.UpdateEvent(oldPod, newPod)
	assert.Equal(t, "podbbb", r.cpm.Checkpoint.Pods["aaa"].Name)

	r.DeleteEvent(newPod)
	_, ok = r.cpm.Checkpoint.Pods["aaa"]
	assert.Equal(t, false, ok)
	r.UpdateEvent(oldPod, newPod)
	assert.Equal(t, "podbbb", r.cpm.Checkpoint.Pods["aaa"].Name)
}

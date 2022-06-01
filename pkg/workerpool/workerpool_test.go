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
// Create: 2021-05-07
// Description: workerpool test

package workerpool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/cachelimit"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
)

func testTaskDoPre() (*config.Config, error) {
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	os.Remove(tmpConfigFile)
	if err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode); err != nil {
		return nil, err
	}
	fd, err := os.Create(tmpConfigFile)
	if err != nil {
		return nil, err
	}
	if _, err = fd.WriteString(`{
						"cgroupRoot": "/tmp/rubik-test"
}`); err != nil {
		return nil, err
	}
	if err = fd.Close(); err != nil {
		return nil, err
	}
	cfg, err := config.NewConfig(tmpConfigFile)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// TestTask_Do is task do function test
func TestTask_Do(t *testing.T) {
	defer os.RemoveAll(constant.TmpTestDir)
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	_, err := testTaskDoPre()
	assert.NoError(t, err)

	data := `{"Pods": {"podabc": {"CgroupPath": "kubepods/abc/abc/abc", "QoSLevel": 1}}}`
	task := &QosTask{req: []byte(data), err: make(chan error, 1), ctx: context.Background()}
	err = task.do()
	assert.Contains(t, err.Error(), "qos level number")
	data = `{"Pods": {"podabc": {"CgroupPath": "kubepods/abc/abc/abc", "QoSLevel": 0},"podabc2": {"CgroupP` +
		`ath": "kubepods/abc/abc/abc2", "QoSLevel": 0}}}`
	task = &QosTask{req: []byte(data), err: make(chan error, 1), ctx: context.Background()}
	for _, podID := range []string{"abc", "abc2"} {
		err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", "kubepods/abc/abc/"+podID))
		assert.NoError(t, err)
		err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "memory", "kubepods/abc/abc/"+podID))
		assert.NoError(t, err)
	}
	if err := task.do(); err != nil {
		assert.Contains(t, err.Error(), "set qos level error")
	}
	for _, podID := range []string{"abc", "abc2"} {
		err = os.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", "kubepods/abc/abc/"+podID),
			constant.DefaultDirMode)
		assert.NoError(t, err)
		_, err = os.Create(filepath.Join(constant.TmpTestDir, "cpu", "kubepods/abc/abc/"+podID, "cpu.qos_level"))
		assert.NoError(t, err)
		err = os.MkdirAll(filepath.Join(constant.TmpTestDir, "memory", "kubepods/abc/abc/"+podID),
			constant.DefaultDirMode)
		assert.NoError(t, err)
		_, err = os.Create(filepath.Join(constant.TmpTestDir, "memory", "kubepods/abc/abc/"+podID,
			"memory.qos_level"))
		assert.NoError(t, err)
	}
	defer func() {
		for _, podID := range []string{"abc", "abc2"} {
			err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", "kubepods/abc/abc/"+podID))
			assert.NoError(t, err)
			err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "memory", "kubepods/abc/abc/"+podID))
			assert.NoError(t, err)
		}
	}()
	err = task.do()
	assert.NoError(t, err)
	err = os.Remove(tmpConfigFile)
	assert.NoError(t, err)
}

// TestNewQosTask is NewQosTask function test
func TestNewQosTask(t *testing.T) {
	data := `{"Pods": {"podabc": {"CgroupPath": "kubepods/abc/abc/abc", "QoSLevel": 1}}}`
	newTask := NewQosTask(context.Background(), []byte(data))

	taskErr := errors.New("task error")
	newTask.error() <- taskErr
	err := <-newTask.error()
	assert.Equal(t, err, taskErr)
}

// test_rubik_pods_number_exceed_max_0001
func TestTask_Do_ExceedMaxPods(t *testing.T) {
	data := `{"Pods": {"podabc": {"CgroupPath": "kubepods/abc/abc", "QoSLevel": -1, "CacheLimitLevel": "low"}`
	for i := 0; i < (constant.MaxPodsPerRequest + 1); i++ {
		data += fmt.Sprintf(`, "podabc%v": {"CgroupPath": "kubepods/a", "QoSLevel": 0, "CacheLimitLevel": "low"}`, i)
	}
	data += `}}`
	task := &QosTask{req: []byte(data), err: make(chan error, 1), ctx: context.Background()}
	err := task.do()
	assert.Contains(t, err.Error(), "could not exceed")
}

type mockTask struct {
	err chan error
	ctx context.Context
}

func (mt *mockTask) do() error {
	sleepTime := 5
	time.Sleep(time.Duration(sleepTime) * time.Second)
	return nil
}

func (mt *mockTask) error() chan error {
	return mt.err
}

func (mt *mockTask) context() context.Context {
	return mt.ctx
}

// test_rubik_wait_requests_exceed_max_0001
func TestExceedMaxWaitTask(t *testing.T) {
	pool := NewWorkerPool(constant.WorkerNum, constant.TaskChanCapacity)
	pool.Start()

	for i := 0; i < (constant.WorkerNum + constant.TaskChanCapacity); i++ {
		mt := &mockTask{err: make(chan error, 1), ctx: context.Background()}
		go func(mt *mockTask) {
			pool.PushTask(mt)
		}(mt)
	}

	i, testTimeout := 1, 5
	for ; i <= testTimeout; i++ {
		// wait until task channel is full
		if len(pool.TasksChan) < constant.TaskChanCapacity {
			time.Sleep(time.Second)
			continue
		}
		mt := &mockTask{err: make(chan error, 1)}
		err := pool.PushTask(mt)
		assert.Contains(t, err.Error(), "exceed")
		break
	}

	if i == (testTimeout + 1) {
		t.Error("test timeout")
	}
}

// TestWaitTaskTimeout tests task do timeout
func TestWaitTaskTimeout(t *testing.T) {
	pool := NewWorkerPool(constant.WorkerNum, constant.TaskChanCapacity)
	pool.Start()

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	mt := &mockTask{err: make(chan error, 1), ctx: ctx}
	err := pool.PushTask(mt)
	assert.Contains(t, err.Error(), "timeout")
}

// TestTaskDoCacheLimit test do cache limit
func TestTaskDoCacheLimit(t *testing.T) {
	defer os.RemoveAll(constant.TmpTestDir)
	cfg, err := testTaskDoPre()
	assert.NoError(t, err)

	data := `{"Pods": {"podabc": {"CgroupPath": "kubepods/abc/abc", "QoSLevel": -1, "CacheLimitLevel": "low"},` +
		`"podabc2": {"CgroupPath": "kubepods/abc/abc2", "QoSLevel": 0, "CacheLimitLevel": "low"}}}`
	task := &QosTask{req: []byte(data), err: make(chan error, 1), ctx: context.Background()}
	for _, podID := range []string{"abc", "abc2"} {
		err = os.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", "kubepods/abc/abc"+podID),
			constant.DefaultDirMode)
		assert.NoError(t, err)
		_, err = os.Create(filepath.Join(constant.TmpTestDir, "cpu", "kubepods/abc/abc"+podID, "cpu.qos_level"))
		assert.NoError(t, err)
		err = os.MkdirAll(filepath.Join(constant.TmpTestDir, "memory", "kubepods/abc/abc"+podID),
			constant.DefaultDirMode)
		assert.NoError(t, err)
		_, err = os.Create(filepath.Join(constant.TmpTestDir, "memory", "kubepods/abc/abc"+podID, "memory.qos_level"))
		assert.NoError(t, err)
	}
	defer func() {
		for _, podID := range []string{"abc", "abc2"} {
			err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", "kubepods/abc/abc"+podID))
			assert.NoError(t, err)
			err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "memory", "kubepods/abc/abc"+podID))
			assert.NoError(t, err)
		}
	}()

	cachelimit.Init(&cfg.CacheCfg)
	err = task.do()
	assert.Equal(t, true, err != nil)
	assert.Equal(t, task.ctx, task.context())
	data = `{"Pods": {"podabc": {"CgroupPath": "kubepods/abc/abc", "QoSLevel": -1, "CacheLimitLevel": "invalid"},` +
		`"podabc2": {"CgroupPath": "kubepods/abc/abc2", "QoSLevel": 0, "CacheLimitLevel": "low"}}}`
	task = &QosTask{req: []byte(data), err: make(chan error, 1), ctx: context.Background()}
	err = task.do()
	assert.Equal(t, true, err != nil)
}

// TestTaskDoDecodeFail test decode fail
func TestTaskDoDecodeFail(t *testing.T) {
	data := `{"Pods": {"podabc": {"CgroupPath": "kubepods/abc, "QoSLevel": 1}}}`
	task := &QosTask{req: []byte(data), err: make(chan error, 1), ctx: context.Background()}
	err := task.do()
	assert.Contains(t, err.Error(), "unmarshal request body failed")
}

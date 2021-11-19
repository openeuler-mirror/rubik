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
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
)

func testTaskDoPre() error {
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	os.Remove(tmpConfigFile)
	if err := os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode); err != nil {
		return err
	}
	fd, err := os.Create(tmpConfigFile)
	if err != nil {
		return err
	}
	if _, err = fd.WriteString(`{
						"cgroupRoot": "/tmp/rubik-test"
}`); err != nil {
		return err
	}
	if err = fd.Close(); err != nil {
		return err
	}
	if _, err = config.NewConfig(tmpConfigFile); err != nil {
		return err
	}
	return nil
}

// TestTask_Do is task do function test
func TestTask_Do(t *testing.T) {
	defer os.RemoveAll(constant.TmpTestDir)
	tmpConfigFile := filepath.Join(constant.TmpTestDir, "config.json")
	err := testTaskDoPre()
	assert.NoError(t, err)
	pods := make(map[string]api.PodQoS, 1)
	pods["abc"] = api.PodQoS{CgroupPath: "kubepods/abc/abc/abc", QosLevel: 1}
	task := &QosTask{req: &api.SetQosRequest{Pods: pods}, err: make(chan error, 1), ctx: context.Background()}
	err = task.do()
	assert.Contains(t, err.Error(), "qos level number")
	podsNum := 2
	pods = make(map[string]api.PodQoS, podsNum)
	pods["abc"] = api.PodQoS{CgroupPath: "kubepods/abc/abc/abc", QosLevel: 0}
	pods["abc2"] = api.PodQoS{CgroupPath: "kubepods/abc2/abc2/abc2", QosLevel: -1}
	task = &QosTask{req: &api.SetQosRequest{Pods: pods}, err: make(chan error, 1), ctx: context.Background()}
	for _, podID := range []string{"abc", "abc2"} {
		err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", pods[podID].CgroupPath))
		assert.NoError(t, err)
		err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "memory", pods[podID].CgroupPath))
		assert.NoError(t, err)
	}
	err = task.do()
	if err := task.do(); err != nil {
		assert.Contains(t, err.Error(), "set qos level error")
	}
	for _, podID := range []string{"abc", "abc2"} {
		err = os.MkdirAll(filepath.Join(constant.TmpTestDir, "cpu", pods[podID].CgroupPath), constant.DefaultDirMode)
		assert.NoError(t, err)
		_, err = os.Create(filepath.Join(constant.TmpTestDir, "cpu", pods[podID].CgroupPath, "cpu.qos_level"))
		assert.NoError(t, err)
		err = os.MkdirAll(filepath.Join(constant.TmpTestDir, "memory", pods[podID].CgroupPath),
			constant.DefaultDirMode)
		assert.NoError(t, err)
		_, err = os.Create(filepath.Join(constant.TmpTestDir, "memory", pods[podID].CgroupPath, "memory.qos_level"))
		assert.NoError(t, err)
	}
	defer func() {
		for _, podID := range []string{"abc", "abc2"} {
			err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "cpu", pods[podID].CgroupPath))
			assert.NoError(t, err)
			err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "memory", pods[podID].CgroupPath))
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
	podQos := api.PodQoS{
		CgroupPath: "/tmp/abc",
		QosLevel:   0,
	}
	pods := make(map[string]api.PodQoS, 1)
	pods["abc"] = podQos
	reqs := api.SetQosRequest{
		Pods: pods,
	}
	newTask := NewQosTask(context.Background(), reqs)
	assert.Equal(t, newTask.req.Pods["abc"].CgroupPath, "/tmp/abc")

	taskErr := errors.New("task error")
	newTask.error() <- taskErr
	err := <-newTask.error()
	assert.Equal(t, err, taskErr)
}

// test_rubik_pods_number_exceed_max_0001
func TestTask_Do_ExceedMaxPods(t *testing.T) {
	pods := make(map[string]api.PodQoS, constant.MaxPodsPerRequest+1)
	for i := 0; i < (constant.MaxPodsPerRequest + 1); i++ {
		pods[strconv.Itoa(i)] = api.PodQoS{
			CgroupPath: "kubepods/abc/abc/abc",
			QosLevel:   0,
		}
	}

	task := &QosTask{
		req: &api.SetQosRequest{
			Pods: pods,
		},
		err: make(chan error, 1),
		ctx: context.Background(),
	}

	err := task.do()
	assert.Equal(t, true, err != nil)
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

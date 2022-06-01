// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: rubik team
// Create: 2021-04-28
// Description: This file is used for unix socket server

package workerpool

import (
	"context"
	"encoding/json"
	"sync/atomic"

	"github.com/pkg/errors"

	"isula.org/rubik/api"
	"isula.org/rubik/pkg/cachelimit"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/qos"
	"isula.org/rubik/pkg/tinylog"
)

type task interface {
	do() error
	error() chan error
	context() context.Context
}

// QosTask defines qos task
type QosTask struct {
	req []byte
	err chan error
	ctx context.Context
}

// WorkerPool defines worker pool struct
type WorkerPool struct {
	WorkerNum  int
	WorkerBusy int32
	TasksChan  chan task
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerNum, chanSize int) *WorkerPool {
	tasksChan := make(chan task, chanSize)
	return &WorkerPool{WorkerNum: workerNum, TasksChan: tasksChan}
}

// Start starts a worker pool
func (pool *WorkerPool) Start() {
	worker := func() {
		for task := range pool.TasksChan {
			atomic.AddInt32(&pool.WorkerBusy, 1)
			select {
			case _, ok := <-task.context().Done():
				if !ok { // for CodeDEX
				}
				task.error() <- errors.Errorf("handle request timeout or context canceled, ctx closed=%v", !ok)
			default:
				task.error() <- task.do()
			}
			atomic.AddInt32(&pool.WorkerBusy, -1)
		}
	}

	for i := 0; i < pool.WorkerNum; i++ {
		go worker()
	}
}

// PushTask pushes a task to worker pool
func (pool *WorkerPool) PushTask(newTask task) error {
	var (
		err error
		ok  bool
	)
	select {
	case pool.TasksChan <- newTask:
		select {
		case err, ok = <-newTask.error():
			if !ok {
				err = errors.New("task channel closed")
			}
		case _, ok = <-newTask.context().Done():
			if !ok { // for CodeDEX
			}
			err = errors.Errorf("handle request timeout or context canceled, ctx closed=%v", !ok)
		}
	default:
		err = errors.New("exceed task channel capacity")
	}

	return err
}

// NewQosTask creates a new QosTask
func NewQosTask(ctx context.Context, reqs []byte) *QosTask {
	return &QosTask{req: reqs, err: make(chan error, 1), ctx: ctx}
}

func (task *QosTask) error() chan error {
	return task.err
}

func (task *QosTask) context() context.Context {
	return task.ctx
}

func (task *QosTask) do() error {
	var (
		errFlag    bool
		sErr, vErr error
		reqs       api.SetQosRequest
	)

	err := json.Unmarshal(task.req, &reqs)
	if err != nil {
		return errors.Errorf("unmarshal request body failed: %v", err)
	}

	if len(reqs.Pods) > constant.MaxPodsPerRequest {
		return errors.Errorf("pods number could not exceed %d per http request", constant.MaxPodsPerRequest)
	}

	for podID, req := range reqs.Pods {
		pod, nErr := qos.NewPodInfo(task.ctx, podID, config.CgroupRoot, req)
		if nErr != nil {
			return nErr
		}
		if sErr = pod.SetQos(); sErr != nil {
			tinylog.WithCtx(task.ctx).Errorf("Set pod %v qos level error: %v", podID, sErr)
			errFlag = true
			continue
		}
		if vErr = pod.ValidateQos(); vErr != nil {
			tinylog.WithCtx(task.ctx).Errorf("Validate pod %v qos level error: %v", podID, vErr)
			errFlag = true
		}
		if !cachelimit.ClEnabled() {
			continue
		}
		if req.QosLevel == constant.MaxLevel.Int() {
			cachelimit.AddOnlinePod(podID, req.CgroupPath)
			continue
		}
		podCacheInfo, err := cachelimit.NewCacheLimitPodInfo(task.ctx, podID, req)
		if err != nil {
			return err
		}
		if err = podCacheInfo.SetCacheLimit(); err != nil {
			tinylog.WithCtx(task.ctx).Errorf("Set pod %v cache limit error: %v", podID, err)
			errFlag = true
		}
	}

	if errFlag {
		return errors.New("set qos level error")
	}

	return nil
}

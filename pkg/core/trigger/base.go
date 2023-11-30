// Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-05-16
// Description: This file is used for diverse triggers

// Package trigger defines diverse triggers
package trigger

import (
	"fmt"
	"sync"

	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
)

// Typ is the type of trigger
type Typ int8

const (
	// ExpulsionAnno is the key of expulsion trigger
	ExpulsionAnno = "expulsion"
	// ResourceAnalysisAnno is the key of resource analysis trigger
	ResourceAnalysisAnno = "resourceAnalysis"
)
const (
	// EXPULSION is the key of expulsion trigger
	EXPULSION Typ = iota
	// RESOURCEANALYSIS is the key of resource analysis trigger
	RESOURCEANALYSIS
)

type triggerCreator func() Trigger

var triggerCreatorMap = map[Typ]triggerCreator{
	EXPULSION:        expulsionCreator,
	RESOURCEANALYSIS: analyzerCreator,
}

// Descriptor defines methods for describing triggers
type Descriptor interface {
	Name() string
}

// Executor is the trigger execution function interface
type Executor interface {
	Execute(Factor) (Factor, error)
	Stop() error
}

// Trigger interface defines the trigger methods
type Trigger interface {
	Descriptor
	Execute(Factor) error
	SetNext(...Trigger) Trigger
}

// TreeTrigger organizes Triggers in a tree format and executes sub-triggers in a chain of responsibility mode
type TreeTrigger struct {
	name        string
	exec        Executor
	subTriggers []Trigger
}

// withTreeTrigger returns a BaseMetric object
func withTreeTrigger(name string, exec Executor) *TreeTrigger {
	return &TreeTrigger{
		name:        name,
		exec:        exec,
		subTriggers: make([]Trigger, 0)}
}

// SetNext sets the trigger that needs to be checked next
func (t *TreeTrigger) SetNext(triggers ...Trigger) Trigger {
	t.subTriggers = append(t.subTriggers, triggers...)
	return t
}

// Name returns the name of trigger
func (t *TreeTrigger) Name() string {
	return t.name
}

// Execute executes the sub-triggers of the current trigger
func (t *TreeTrigger) Execute(f Factor) error {
	if t.exec == nil {
		return fmt.Errorf("trigger can not execute: %v", t.name)
	}
	var errs error
	res, err := t.exec.Execute(f)
	if err != nil {
		return fmt.Errorf("failed to execute %v: %v", t.name, err)
	}

	for _, next := range t.subTriggers {
		if err := next.Execute(res); err != nil {
			errs = util.AppendErr(errs, util.AddErrorPrfix(err, t.name))
		}
	}
	return errs
}

// NewTrigger returns a trigger singleton according to type
func NewTrigger(t Typ) Trigger {
	if _, ok := triggerCreatorMap[t]; !ok {
		log.Warnf("undefine trigger: %v", t)
		return nil
	}
	return triggerCreatorMap[t]()
}

var (
	runningExecutors = make(map[string]Executor, 0)
	execLock         sync.Mutex
)

func appendUsedExecutors(name string, exec Executor) {
	if exec == nil {
		log.Errorf("invalid executor: %v", name)
		return
	}
	execLock.Lock()
	defer execLock.Unlock()
	if _, existed := runningExecutors[name]; existed {
		log.Errorf("the executor already exist: %v", name)
		return
	}
	log.Infof("append executor successfully: %v", name)
	runningExecutors[name] = exec
}

// StopUsedExecutors stops running executors
func StopUsedExecutors() error {
	execLock.Lock()
	defer execLock.Unlock()
	var errs error
	// stop the executors one by one
	for name, exec := range runningExecutors {
		log.Infof("stopping executor %v", name)
		if exec == nil {
			log.Infof("executor has already stopped: %v", name)
			continue
		}
		if err := exec.Stop(); err != nil {
			errs = util.AppendErr(errs, util.AddErrorPrfix(err, name))
		}
	}
	// clear executors
	for k := range runningExecutors {
		delete(runningExecutors, k)
	}
	return errs
}

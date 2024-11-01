// Copyright (c) Huawei Technologies Co., Ltd. 2021-2024. All rights reserved.
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
// Description: This file is used for common triggers

// Package common defines trigger interface and tree trigger
package common

import (
	"context"
	"fmt"

	"isula.org/rubik/pkg/common/util"
)

// TreeTrigger organizes Triggers in a tree format and executes sub-triggers in a chain of responsibility mode
type TreeTrigger struct {
	name        string
	exec        Executor
	subTriggers []Trigger
}

// NewTreeTrigger returns a BaseMetric object
func NewTreeTrigger(name string) *TreeTrigger {
	return &TreeTrigger{
		name:        name,
		subTriggers: make([]Trigger, 0)}
}

// Name returns the name of trigger
func (t *TreeTrigger) Name() string {
	return t.name
}

// Activate executes the sub-triggers of the current trigger
func (t *TreeTrigger) Activate(ctx context.Context) error {
	if t.exec == nil {
		return fmt.Errorf("trigger can not execute: %v", t.name)
	}
	var errs error
	res, err := t.exec.Execute(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute %v: %v", t.name, err)
	}

	for _, next := range t.subTriggers {
		if err := next.Activate(res); err != nil {
			errs = util.AppendErr(errs, util.AddErrorPrfix(err, t.name))
		}
	}
	return errs
}

// Halt stops the sub-triggers of the current trigger and current trigger
func (t *TreeTrigger) Halt() error {
	if t.exec == nil {
		return nil
	}
	var errs error

	for _, next := range t.subTriggers {
		errs = util.AppendErr(errs, next.Halt())
	}
	errs = util.AppendErr(errs, t.exec.Stop())
	if errs != nil {
		return fmt.Errorf("failed to stop trigger %v: %v", t.name, errs)
	}
	return errs
}

// SetName sets the trigger name
func (t *TreeTrigger) SetName(name string) {
	t.name = name
}

// SetExecutor sets the executor of the trigger
func (t *TreeTrigger) SetExecutor(exec Executor) {
	t.exec = exec
}

// SetNext sets the triggers that need to be checked next
func (t *TreeTrigger) SetNext(triggers ...Trigger) Trigger {
	t.subTriggers = append(t.subTriggers, triggers...)
	return t
}

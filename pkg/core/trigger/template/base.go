// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2024-10-30
// Description: This file is used for It provides trigger templates
// Further, just provide the name and the operation to be performed

package template

import (
	"context"
	"fmt"

	"isula.org/rubik/pkg/core/trigger/common"
)

// Transformation returns the pods after convertion
type Transformation func(context.Context) (context.Context, error)

// Action acts on Pods
type Action func(context.Context) error

// ResourceAnalyzer is the resource analysis trigger
type BaseTemplate struct {
	common.TreeTrigger
	transfomer Transformation
	actor      Action
}

// Execute filters the corresponding Pod according to the operation type and triggers it on demand
func (t *BaseTemplate) Execute(ctx context.Context) (context.Context, error) {
	if t.transfomer != nil {
		return transform(ctx, t.transfomer)
	}

	if t.actor != nil {
		return ctx, act(ctx, t.actor)
	}
	return ctx, nil
}

// Stop stops the executor
func (t *BaseTemplate) Stop() error {
	return nil
}

func transform(ctx context.Context, f Transformation) (context.Context, error) {
	if f == nil {
		return nil, fmt.Errorf("podFilter method is not implemented")
	}
	// pods, ok := ctx.Value(common.TARGETPODS).(map[string]*typedef.PodInfo)
	// if !ok {
	// 	return ctx, fmt.Errorf("failed to get target pods")
	// }
	ctx, err := f(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to transform pod: %v", err)
	}
	return ctx, nil
}

func act(ctx context.Context, f Action) error {
	if f == nil {
		return fmt.Errorf("podAction method is not implemented")
	}
	// pods, ok := ctx.Value(common.TARGETPODS).(map[string]*typedef.PodInfo)
	// if !ok {
	// 	return nil
	// }
	return f(ctx)
}

type opt func(t *BaseTemplate)

func WithName(name string) opt {
	return func(t *BaseTemplate) {
		t.SetName(name)
	}
}

func WithPodTransformation(f Transformation) opt {
	return func(t *BaseTemplate) {
		t.transfomer = f
	}
}

func WithPodAction(f Action) opt {
	return func(t *BaseTemplate) {
		t.actor = f
	}
}

func FromBaseTemplate(opts ...opt) common.Trigger {
	t := &BaseTemplate{
		TreeTrigger: *common.NewTreeTrigger("base template"),
	}
	for _, opt := range opts {
		opt(t)
	}
	t.SetExecutor(t)
	return t
}

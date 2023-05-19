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
// Description: This file is used for *

package trigger

import "isula.org/rubik/pkg/common/log"

// expulsionExec is the singleton of eviction triggers
var expulsionExec = &Expulsion{}
var expulsionCreator = func() Trigger {
	return &TreeTrigger{name: ExpulsionAnno, exec: expulsionExec}
}

// Expulsion is the trigger to evict pods
type Expulsion struct{}

// Execute evicts pods based on the id of the given pod
func (e *Expulsion) Execute(f Factor) (Factor, error) {
	log.Infof("need to evict pod %v, real operation is TO BE CONTINUE", f.Message())
	return nil, nil
}

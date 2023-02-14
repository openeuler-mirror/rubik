// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: jingrui
// Create: 2022-04-17
// Description: try provide some helper functions for unit-test.
//
// Package try provide some helper function for unit-test, if you want
// to use try outside unit-test, please add notes.

package try

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_OrDie test try some-func or die.
func Test_OrDie(t *testing.T) {
	ret := GenTestDir()
	ret.OrDie()
	dname := ret.String()
	WriteFile(SecureJoin(dname, "die.txt").String(), "ok").OrDie()
	RemoveAll(dname).OrDie()
}

// Test_ErrMessage test try some-func or check the error.
func Test_ErrMessage(t *testing.T) {
	ret := GenTestDir()
	assert.Equal(t, ret.ErrMessage(), "")
	dname := ret.String()
	WriteFile(SecureJoin(dname, "log.txt").String(), "ok").ErrMessage()
	assert.Equal(t, RemoveAll(dname).ErrMessage(), "")
}

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
//
// 1. Try some function quiet|log|die on error. because some test
// function does not care the error returned, but the code checker
// always generate unuseful warnings. This method can suppress the
// noisy warnings.
//
// 2. Provide testdir helper to generate tmpdir for unitest.
//
package try

import (
	"fmt"
	"io/ioutil"
	"os"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/google/uuid"

	"isula.org/rubik/pkg/constant"
)

// Ret provide some action for error.
type Ret struct {
	val interface{}
	err error
}

func newRet(err error) Ret {
	return Ret{
		val: nil,
		err: err,
	}
}

// OrDie try func or die with fail.
func (r Ret) OrDie() {
	if r.err == nil {
		return
	}
	fmt.Printf("try failed, die with error %v", r.err)
	os.Exit(-1)
}

// ErrMessage get ret error string format
func (r Ret) ErrMessage() string {
	if r.err == nil {
		return ""
	}
	return fmt.Sprintf("%v", r.err)
}

// String get ret val and convert to string
func (r Ret) String() string {
	val, ok := r.val.(string)
	if ok {
		return val
	}
	return ""
}

// SecureJoin wrap error to Ret.
func SecureJoin(root, unsafe string) Ret {
	name, err := securejoin.SecureJoin(root, unsafe)
	ret := newRet(err)
	if err == nil {
		ret.val = name
	}
	return ret
}

// MkdirAll wrap error to Ret.
func MkdirAll(path string, perm os.FileMode) Ret {
	if err := os.MkdirAll(path, perm); err != nil {
		return newRet(err)
	}
	return newRet(nil)
}

// RemoveAll wrap error to Ret.
func RemoveAll(path string) Ret {
	if err := os.RemoveAll(path); err != nil {
		return newRet(err)
	}
	return newRet(nil)
}

// WriteFile wrap error to Ret.
func WriteFile(filename string, data []byte, perm os.FileMode) Ret {
	ret := newRet(nil)
	ret.val = filename
	if err := ioutil.WriteFile(filename, data, perm); err != nil {
		ret.err = err
	}
	return ret
}

const (
	testdir = "/tmp/rubik-test"
)

// GenTestDir gen testdir
func GenTestDir() Ret {
	name := fmt.Sprintf("%s/%s", testdir, uuid.New().String())
	ret := MkdirAll(name, constant.DefaultDirMode)
	ret.val = name
	return ret
}

// DelTestDir del testdir, this function only need call once.
func DelTestDir() Ret {
	return RemoveAll(testdir)
}

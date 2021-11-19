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
// Create: 2021-05-12
// Description: filelock related tests

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/constant"
)

// TestCreateLockFile is CreateLockFile function test
func TestCreateLockFile(t *testing.T) {
	lockFile := filepath.Join(constant.TmpTestDir, "rubik.lock")
	err := os.RemoveAll(lockFile)
	assert.NoError(t, err)

	lock, err := CreateLockFile(lockFile)
	assert.NoError(t, err)
	RemoveLockFile(lock, lockFile)
}

// TestLockFail is CreateLockFile fail test
func TestLockFail(t *testing.T) {
	lockFile := filepath.Join(constant.TmpTestDir, "rubik.lock")
	err := os.RemoveAll(constant.TmpTestDir)
	assert.NoError(t, err)
	os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)

	_, err = os.Create(filepath.Join(constant.TmpTestDir, "rubik-lock"))
	assert.NoError(t, err)
	_, err = CreateLockFile(filepath.Join(constant.TmpTestDir, "rubik-lock", "rubik.lock"))
	assert.Equal(t, true, err != nil)
	err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "rubik-lock"))
	assert.NoError(t, err)

	err = os.MkdirAll(lockFile, constant.DefaultDirMode)
	assert.NoError(t, err)
	_, err = CreateLockFile(lockFile)
	assert.Equal(t, true, err != nil)
	err = os.RemoveAll(lockFile)
	assert.NoError(t, err)

	_, err = CreateLockFile(lockFile)
	assert.NoError(t, err)
	_, err = CreateLockFile(lockFile)
	assert.Equal(t, true, err != nil)
	err = os.RemoveAll(lockFile)
	assert.NoError(t, err)
}

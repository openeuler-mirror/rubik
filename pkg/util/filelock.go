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
// Create: 2021-10-18
// Description: filelock related common functions

package util

import (
	"os"
	"path/filepath"
	"syscall"

	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/tinylog"
)

// CreateLockFile creates a lock file
func CreateLockFile(p string) (*os.File, error) {
	path := filepath.Clean(p)
	if err := os.MkdirAll(filepath.Dir(path), constant.DefaultDirMode); err != nil {
		return nil, err
	}

	lock, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, constant.DefaultFileMode)
	if err != nil {
		return nil, err
	}

	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		tinylog.DropError(lock.Close())
		return nil, err
	}

	return lock, nil
}

// RemoveLockFile removes lock file - this function used cleanup resource,
// errors will ignored to make sure more source is cleaned.
func RemoveLockFile(lock *os.File, path string) {
	tinylog.DropError(syscall.Flock(int(lock.Fd()), syscall.LOCK_UN))
	tinylog.DropError(lock.Close())
	tinylog.DropError(os.Remove(path))
}

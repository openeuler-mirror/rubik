// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: hanchao
// Create: 2023-03-11
// Description: This file is used to implement blkio system call

// Package iocost provide iocost isolate feature.
package iocost

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const (
	devNoMax = 256
)

func getBlkDeviceNo(devName string) (string, error) {
	devPath := filepath.Join("/dev", devName)
	fi, err := os.Stat(devPath)
	if err != nil {
		return "", fmt.Errorf("stat %s failed with error: %v", devName, err)
	}

	if fi.Mode()&os.ModeDevice == 0 {
		return "", fmt.Errorf("%s is not a device", devName)
	}

	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("failed to get Sys(), %v has type %v", devName, st)
	}

	devno := st.Rdev
	major, minor := devno/devNoMax, devno%devNoMax
	return fmt.Sprintf("%v:%v", major, minor), nil
}

func getDirInode(file string) (uint64, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get Sys(), %v has type %v", file, st)
	}
	return st.Ino, nil
}

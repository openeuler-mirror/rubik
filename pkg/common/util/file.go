// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2021-04-17
// Description: filepath related common functions

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"isula.org/rubik/pkg/common/constant"
)

const (
	fileMaxSize = 10 * 1024 * 1024 // 10MB
)

// CreateFile create full path including dir and file.
func CreateFile(path string) (*os.File, error) {
	path = filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(path), constant.DefaultDirMode); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, constant.DefaultFileMode)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// IsDir returns true if the file exists and it is a dir
func IsDir(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// ReadSmallFile read small file less than 10MB
func ReadSmallFile(path string) ([]byte, error) {
	size, err := FileSize(path)
	if err != nil {
		return nil, err
	}

	if size > fileMaxSize {
		return nil, fmt.Errorf("file too big")
	}
	return ReadFile(path)
}

// FileSize returns the size of file
func FileSize(path string) (int64, error) {
	if !PathExist(path) {
		return 0, fmt.Errorf("%v: No such file or directory", path)
	}
	st, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// PathExist returns true if the path exists
func PathExist(path string) bool {
	if _, err := os.Lstat(path); err != nil {
		return false
	}

	return true
}

// LockFile locks a file, creating a file if it does not exist
func LockFile(path string) (*os.File, error) {
	lock, err := CreateFile(path)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return lock, err
	}
	return lock, nil
}

// UnlockFile unlock file - this function used cleanup resource,
func UnlockFile(lock *os.File) error {
	// errors will ignored to make sure more source is cleaned.
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_UN); err != nil {
		return err
	}
	return nil
}

// ReadFile reads a file
func ReadFile(path string) ([]byte, error) {
	path = filepath.Clean(path)
	if IsDir(path) {
		return nil, fmt.Errorf("%v is not a file", path)
	}
	if !PathExist(path) {
		return nil, fmt.Errorf("%v: No such file or directory", path)
	}
	return ioutil.ReadFile(path)
}

// WriteFile writes a file, if the file does not exist, create the file (including the upper directory)
func WriteFile(path, content string) error {
	if IsDir(path) {
		return fmt.Errorf("%v is not a file", path)
	}
	// try to create parent directory
	dirPath := filepath.Dir(path)
	if !PathExist(dirPath) {
		if err := os.MkdirAll(dirPath, constant.DefaultDirMode); err != nil {
			return fmt.Errorf("error create dir %v: %v", dirPath, err)
		}
	}
	return ioutil.WriteFile(path, []byte(content), constant.DefaultFileMode)
}

// AppendFile appends content to the file
func AppendFile(path, content string) error {
	if IsDir(path) {
		return fmt.Errorf("%v is not a file", path)
	}
	if !PathExist(path) {
		return fmt.Errorf("%v: No such file or directory", path)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constant.DefaultFileMode)
	defer func() {
		if err != f.Close() {
			return
		}
	}()
	if err != nil {
		return fmt.Errorf("error open file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("error write file: %v", err)
	}
	return nil
}

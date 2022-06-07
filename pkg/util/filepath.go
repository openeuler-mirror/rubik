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
	"io/ioutil"
	"os"
	"path/filepath"

	"isula.org/rubik/pkg/constant"
)

const (
	fileMaxSize = 10 * 1024 * 1024 // 10MB
)

// CreateFile create full path including dir and file.
func CreateFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), constant.DefaultDirMode); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return f.Close()
}

// IsDirectory returns true if the file exists and it is a dir
func IsDirectory(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// ReadSmallFile read small file less than 10MB
func ReadSmallFile(path string) ([]byte, error) {
	st, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if st.Size() > fileMaxSize {
		return nil, constant.ErrFileTooBig
	}
	return ioutil.ReadFile(path) // nolint: gosec
}

// PathExist returns true if the path exists
func PathExist(path string) bool {
	if _, err := os.Lstat(path); err != nil {
		return false
	}

	return true
}

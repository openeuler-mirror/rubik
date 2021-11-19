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
// Description: filepath related common functions testing

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/constant"
)

// TestIsDirectory is IsDirectory function test
func TestIsDirectory(t *testing.T) {
	directory, err := ioutil.TempDir(constant.TmpTestDir, t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(directory)

	filePath, err := ioutil.TempFile(directory, t.Name())
	assert.NoError(t, err)

	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TC1-directory is exist",
			args: args{path: directory},
			want: true,
		},
		{
			name: "TC2-directory is not exist",
			args: args{path: "/directory/is/not/exist"},
			want: false,
		},
		{
			name: "TC3-test file path",
			args: args{path: filePath.Name()},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDirectory(tt.args.path); got != tt.want {
				t.Errorf("IsDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
	err = filePath.Close()
	assert.NoError(t, err)
}

// TestPathIsExist is PathExist function test
func TestPathIsExist(t *testing.T) {
	filePath, err := ioutil.TempDir(constant.TmpTestDir, "file_exist")
	assert.NoError(t, err)
	defer os.RemoveAll(filePath)

	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TC1-path is exist",
			args: args{path: filePath},
			want: true,
		},
		{
			name: "TC2-path is not exist",
			args: args{path: "/path/is/not/exist"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathExist(tt.args.path); got != tt.want {
				t.Errorf("PathExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestReadSmallFile is test for read file
func TestReadSmallFile(t *testing.T) {
	filePath, err := ioutil.TempDir(constant.TmpTestDir, "read_file")
	assert.NoError(t, err)
	defer os.RemoveAll(filePath)

	// case1: ok
	err = ioutil.WriteFile(filepath.Join(filePath, "ok"), []byte{}, constant.DefaultFileMode)
	assert.NoError(t, err)
	_, err = ReadSmallFile(filepath.Join(filePath, "ok"))
	assert.NoError(t, err)

	// case2: too big
	size := 20000000
	big := make([]byte, size, size)
	err = ioutil.WriteFile(filepath.Join(filePath, "big"), big, constant.DefaultFileMode)
	assert.NoError(t, err)
	_, err = ReadSmallFile(filepath.Join(filePath, "big"))
	assert.Error(t, err)

	// case3: file not exist
	_, err = ReadSmallFile(filepath.Join(filePath, "missing"))
	assert.Error(t, err)
}

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

// Package util is common utilitization
package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/common/constant"
)

// TestIsDir is IsDir function test
func TestIsDir(t *testing.T) {
	os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode)
	defer os.RemoveAll(constant.TmpTestDir)
	directory, err := ioutil.TempDir(constant.TmpTestDir, t.Name())
	assert.NoError(t, err)

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
			if got := IsDir(tt.args.path); got != tt.want {
				t.Errorf("IsDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
	err = filePath.Close()
	assert.NoError(t, err)
}

// TestPathIsExist is PathExist function test
func TestPathIsExist(t *testing.T) {
	os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode)
	defer os.RemoveAll(constant.TmpTestDir)
	filePath, err := ioutil.TempDir(constant.TmpTestDir, "file_exist")
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
	os.RemoveAll(constant.TmpTestDir)
	os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode)
	defer os.RemoveAll(constant.TmpTestDir)
	filePath, err := ioutil.TempDir(constant.TmpTestDir, "read_file")
	assert.NoError(t, err)

	// case1: ok
	err = ioutil.WriteFile(filepath.Join(filePath, "ok"), []byte{}, constant.DefaultFileMode)
	assert.NoError(t, err)
	_, err = ReadSmallFile(filepath.Join(filePath, "ok"))
	assert.NoError(t, err)

	// case2: too big
	size := 20000000
	big := make([]byte, size)
	err = ioutil.WriteFile(filepath.Join(filePath, "big"), big, constant.DefaultFileMode)
	assert.NoError(t, err)
	_, err = ReadSmallFile(filepath.Join(filePath, "big"))
	assert.Error(t, err)

	// case3: file not exist
	_, err = ReadSmallFile(filepath.Join(filePath, "missing"))
	assert.Error(t, err)
}

// TestCreateLockFile is CreateLockFile function test
func TestCreateLockFile(t *testing.T) {
	lockFile := filepath.Join(constant.TmpTestDir, "rubik.lock")
	err := os.RemoveAll(lockFile)
	assert.NoError(t, err)

	lock, err := LockFile(lockFile)
	assert.NoError(t, err)
	UnlockFile(lock)
	assert.NoError(t, lock.Close())
	assert.NoError(t, os.Remove(lockFile))
}

// TestLockFail is CreateLockFile fail test
func TestLockFail(t *testing.T) {
	lockFile := filepath.Join(constant.TmpTestDir, "rubik.lock")
	err := os.RemoveAll(constant.TmpTestDir)
	assert.NoError(t, err)
	os.MkdirAll(constant.TmpTestDir, constant.DefaultDirMode)

	_, err = os.Create(filepath.Join(constant.TmpTestDir, "rubik-lock"))
	assert.NoError(t, err)
	_, err = LockFile(filepath.Join(constant.TmpTestDir, "rubik-lock", "rubik.lock"))
	assert.Equal(t, true, err != nil)
	err = os.RemoveAll(filepath.Join(constant.TmpTestDir, "rubik-lock"))
	assert.NoError(t, err)

	err = os.MkdirAll(lockFile, constant.DefaultDirMode)
	assert.NoError(t, err)
	_, err = LockFile(lockFile)
	assert.Equal(t, true, err != nil)
	err = os.RemoveAll(lockFile)
	assert.NoError(t, err)

	_, err = LockFile(lockFile)
	assert.NoError(t, err)
	_, err = LockFile(lockFile)
	assert.Equal(t, true, err != nil)
	err = os.RemoveAll(lockFile)
	assert.NoError(t, err)
}

// TestReadFile tests ReadFile
func TestReadFile(t *testing.T) {
	os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode)
	defer os.RemoveAll(constant.TmpTestDir)
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		pre     func(t *testing.T)
		post    func(t *testing.T)
		want    []byte
		wantErr bool
	}{
		{
			name: "TC1-path is dir",
			args: args{
				path: constant.TmpTestDir,
			},
			pre: func(t *testing.T) {
				_, err := ioutil.TempDir(constant.TmpTestDir, "TC1")
				assert.NoError(t, err)
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre(t)
			}
			got, err := ReadFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFile() = %v, want %v", got, tt.want)
			}
			if tt.post != nil {
				tt.post(t)
			}

		})
	}
}

// TestWriteFile tests WriteFile
func TestWriteFile(t *testing.T) {
	os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode)
	defer os.RemoveAll(constant.TmpTestDir)
	var filePath = filepath.Join(constant.TmpTestDir, "cpu", "kubepods", "PodXXX")
	type args struct {
		path    string
		content string
	}
	tests := []struct {
		name    string
		args    args
		pre     func(t *testing.T)
		post    func(t *testing.T)
		wantErr bool
	}{
		{
			name: "TC1-path is dir",
			args: args{
				path: constant.TmpTestDir,
			},
			pre: func(t *testing.T) {
				_, err := ioutil.TempDir(constant.TmpTestDir, "TC1")
				assert.NoError(t, err)
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			wantErr: true,
		},
		{
			name: "TC2-create dir & write file",
			args: args{
				path:    filePath,
				content: "1",
			},
			pre: func(t *testing.T) {
				assert.NoError(t, os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode))
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre(t)
			}
			if err := WriteFile(tt.args.path, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.post != nil {
				tt.post(t)
			}
		})
	}
}

func TestAppendFile(t *testing.T) {
	os.Mkdir(constant.TmpTestDir, constant.DefaultDirMode)
	defer os.RemoveAll(constant.TmpTestDir)
	var (
		dirPath  = filepath.Join(constant.TmpTestDir, "cpu", "kubepods", "PodXXX")
		filePath = filepath.Join(dirPath, "cpu.cfs_quota_us")
	)
	type args struct {
		path    string
		content string
	}
	tests := []struct {
		name    string
		args    args
		pre     func(t *testing.T)
		post    func(t *testing.T)
		wantErr bool
	}{
		{
			name: "TC1-path is dir",
			args: args{
				path: constant.TmpTestDir,
			},
			pre: func(t *testing.T) {
				_, err := ioutil.TempDir(constant.TmpTestDir, "TC1")
				assert.NoError(t, err)
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(filepath.Join(constant.TmpTestDir, "TC1")))
			},
			wantErr: true,
		},
		{
			name: "TC2-empty path",
			args: args{
				path: dirPath,
			},
			pre: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			wantErr: true,
		},
		{
			name: "TC3-write file success",
			args: args{
				path:    filePath,
				content: "1",
			},
			pre: func(t *testing.T) {
				assert.NoError(t, os.MkdirAll(dirPath, constant.DefaultDirMode))
				assert.NoError(t, ioutil.WriteFile(filePath, []byte(""), constant.DefaultFileMode))
			},
			post: func(t *testing.T) {
				assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre(t)
			}
			err := AppendFile(tt.args.path, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppendFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
			if tt.post != nil {
				tt.post(t)
			}
		})
	}
}

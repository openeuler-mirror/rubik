// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Create: 2023-01-05
// Description: This file defines cgroupAttr and CgroupKey

// Package cgroup uses map to manage cgroup parameters and provides a friendly and simple cgroup usage method
package cgroup

import (
	"fmt"
	"path/filepath"
	"strconv"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
)

var rootDir = constant.DefaultCgroupRoot

// ReadCgroupFile reads data from cgroup files
func ReadCgroupFile(subsys, cgroupParent, cgroupFileName string) ([]byte, error) {
	cgfile := filepath.Join(rootDir, subsys, cgroupParent, cgroupFileName)
	if !util.PathExist(cgfile) {
		return nil, fmt.Errorf("%v: no such file or diretory", cgfile)
	}
	return util.ReadFile(cgfile)
}

// WriteCgroupFile writes data to cgroup file
func WriteCgroupFile(subsys, cgroupParent, cgroupFileName string, content string) error {
	cgfile := filepath.Join(rootDir, subsys, cgroupParent, cgroupFileName)
	if !util.PathExist(cgfile) {
		return fmt.Errorf("%v: no such file or diretory", cgfile)
	}
	return util.WriteFile(cgfile, content)
}

// InitMountDir sets the mount directory of the cgroup file system
func InitMountDir(arg string) {
	rootDir = arg
}

type (
	// Key uniquely determines the cgroup value of the container or pod
	Key struct {
		// SubSys refers to the subsystem of the cgroup
		SubSys string
		// FileName represents the cgroup file name
		FileName string
	}
	// Attr represents a single cgroup attribute, and Err represents whether the Value is available
	Attr struct {
		Value string
		Err   error
	}
	// SetterAndGetter is used for set and get value to/from cgroup file
	SetterAndGetter interface {
		SetCgroupAttr(*Key, string) error
		GetCgroupAttr(*Key) *Attr
	}
)

// Expect judges whether Attr is consistent with the input
func (attr *Attr) Expect(arg interface{}) error {
	if attr.Err != nil {
		return attr.Err
	}

	switch arg := arg.(type) {
	case int:
		value, err := attr.Int()
		if err != nil {
			return fmt.Errorf("fail to convert: %v", err)
		}
		if value != arg {
			return fmt.Errorf("%v is not equal to %v", value, arg)
		}
	case string:
		if attr.Value != arg {
			return fmt.Errorf("%v is not equal to %v", attr.Value, arg)
		}
	case int64:
		value, err := attr.Int64()
		if err != nil {
			return fmt.Errorf("fail to convert: %v", err)
		}
		if value != arg {
			return fmt.Errorf("%v is not equal to %v", value, arg)
		}
	default:
		return fmt.Errorf("invalid expect type: %T", arg)
	}
	return nil
}

// Int64 parses CgroupAttr as int64 type
func (attr *Attr) Int64() (int64, error) {
	if attr.Err != nil {
		return 0, attr.Err
	}
	return util.ParseInt64(attr.Value)
}

// Int parses CgroupAttr as int type
func (attr *Attr) Int() (int, error) {
	if attr.Err != nil {
		return 0, attr.Err
	}
	return strconv.Atoi(attr.Value)
}

// Int64Map parses CgroupAttr64 as map[string]int64 type
func (attr *Attr) Int64Map() (map[string]int64, error) {
	if attr.Err != nil {
		return nil, attr.Err
	}
	return util.ParseInt64Map(attr.Value)
}

// CPUStat parses CgroupAttr64 as CPUStat type
func (attr *Attr) CPUStat() (*CPUStat, error) {
	if attr.Err != nil {
		return nil, attr.Err
	}
	return NewCPUStat(attr.Value)
}

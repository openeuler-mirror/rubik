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
	"strings"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
)

// Config is the configuration of cgroup
type Config struct {
	RootDir      string
	CgroupDriver Driver
}

var conf = &Config{
	RootDir:      constant.DefaultCgroupRoot,
	CgroupDriver: defaultDriver(),
}

type option func(c *Config) error

func WithRoot(cgroupRoot string) option {
	return func(c *Config) error {
		c.RootDir = cgroupRoot
		return nil
	}
}

func WithDriver(driverType string) option {
	return func(c *Config) error {
		d, err := newCgroupDriver(driverType)
		if err != nil {
			return err
		}
		c.CgroupDriver = d
		return nil
	}
}

// Init sets the mount directory of the cgroup file system & driver
func Init(opts ...option) error {
	for _, opt := range opts {
		if err := opt(conf); err != nil {
			return fmt.Errorf("failed to init cgroup: %v", err)
		}
	}
	return nil
}

// AbsoluteCgroupPath returns the absolute path of the cgroup
func AbsoluteCgroupPath(elem ...string) string {
	elem = append([]string{conf.RootDir}, elem...)
	return filepath.Join(elem...)
}

// ReadCgroupFile reads data from cgroup files
func ReadCgroupFile(elem ...string) ([]byte, error) {
	return readCgroupFile(AbsoluteCgroupPath(elem...))
}

// WriteCgroupFile writes data to cgroup file
func WriteCgroupFile(content string, elem ...string) error {
	return writeCgroupFile(AbsoluteCgroupPath(elem...), content)
}

func readCgroupFile(cgPath string) ([]byte, error) {
	if !util.PathExist(cgPath) {
		return nil, fmt.Errorf("%v: no such file or directory", cgPath)
	}
	return util.ReadFile(cgPath)
}

func writeCgroupFile(cgPath, content string) error {
	if !util.PathExist(cgPath) {
		return fmt.Errorf("%v: no such file or directory", cgPath)
	}
	return util.WriteFile(cgPath, content)
}

// GetMountDir returns the mount point path of the cgroup
func GetMountDir() string {
	return conf.RootDir
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
			return fmt.Errorf("failed to convert: %v", err)
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
			return fmt.Errorf("failed to convert: %v", err)
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

// Int64Map parses CgroupAttr as map[string]int64 type
func (attr *Attr) Int64Map() (map[string]int64, error) {
	if attr.Err != nil {
		return nil, attr.Err
	}
	return util.ParseInt64Map(attr.Value)
}

// CPUStat parses CgroupAttr as CPUStat type
func (attr *Attr) CPUStat() (*CPUStat, error) {
	if attr.Err != nil {
		return nil, attr.Err
	}
	return NewCPUStat(attr.Value)
}

// PSI parses CgroupAttr as psi Pressure type
func (attr *Attr) PSI() (*Pressure, error) {
	if attr.Err != nil {
		return nil, attr.Err
	}
	return NewPSIData(attr.Value)
}

// Hierarchy is used to represent a cgroup path
type Hierarchy struct {
	MountPoint string `json:"mountPoint,omitempty"`
	Path       string `json:"cgroupPath"`
}

// NewHierarchy creates a Hierarchy instance
func NewHierarchy(mountPoint, path string) *Hierarchy {
	return &Hierarchy{
		MountPoint: mountPoint,
		Path:       path,
	}
}

// SetCgroupAttr sets value to the cgroup file
func (h *Hierarchy) SetCgroupAttr(key *Key, value string) error {
	if err := validateCgroupKey(key); err != nil {
		return err
	}
	var mountPoint = conf.RootDir
	if len(h.MountPoint) > 0 {
		mountPoint = h.MountPoint
	}
	return writeCgroupFile(filepath.Join(mountPoint, key.SubSys, h.Path, key.FileName), value)
}

// GetCgroupAttr gets cgroup file content
func (h *Hierarchy) GetCgroupAttr(key *Key) *Attr {
	if err := validateCgroupKey(key); err != nil {
		return &Attr{Err: err}
	}
	var mountPoint = conf.RootDir
	if len(h.MountPoint) > 0 {
		mountPoint = h.MountPoint
	}
	data, err := readCgroupFile(filepath.Join(mountPoint, key.SubSys, h.Path, key.FileName))
	if err != nil {
		return &Attr{Err: err}
	}
	return &Attr{Value: strings.TrimSpace(string(data)), Err: nil}
}

// validateCgroupKey is used to verify the validity of the cgroup key
func validateCgroupKey(key *Key) error {
	if key == nil {
		return fmt.Errorf("key cannot be empty")
	}
	if len(key.SubSys) == 0 || len(key.FileName) == 0 {
		return fmt.Errorf("invalid key")
	}
	return nil
}

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
// Date: 2023-03-09
// Description: This file is used for quota turbo client

// Package quotaturbo is for Quota Turbo feature
package quotaturbo

import (
	"fmt"
)

// ConfigViewer is used to view the parameters of quotaTurbo
type ConfigViewer interface {
	AlarmWaterMark() int
	HighWaterMark() int
	CgroupRoot() string
	ElevateLimit() float64
	SlowFallbackRatio() float64
	CPUFloatingLimit() float64
}

// ClientAPI is the client interface
type ClientAPI interface {
	// Quota detection and adjustment
	AdjustQuota() error
	// Cgroup lifecycle management
	AddCgroup(string, float64) error
	RemoveCgroup(string) error
	AllCgroups() []string
	// Parameter management
	WithOptions(...Option) error
	ConfigViewer
}

// Client is quotaTurbo client
type Client struct {
	*StatusStore
	Driver
}

// NewClient returns a quotaTurbo client instance
func NewClient() ClientAPI {
	return &Client{
		StatusStore: NewStatusStore(),
		Driver:      &EventDriver{},
	}
}

// WithOptions configures parameters for the client
func (c *Client) WithOptions(opts ...Option) error {
	if len(opts) == 0 {
		return nil
	}
	var (
		conf = *c.StatusStore.Config
		errs error
	)
	for _, opt := range opts {
		if err := opt(&conf); err != nil {
			errs = appendErr(errs, err)
		}
	}
	if errs != nil {
		return errs
	}
	c.StatusStore.Config = &conf
	return nil
}

//  AdjustQuota is used to update status and adjust cgroup quota value
func (c *Client) AdjustQuota() error {
	if err := c.updateCPUUtils(); err != nil {
		return fmt.Errorf("fail to get current cpu utilization: %v", err)
	}
	if len(c.cpuQuotas) == 0 {
		return nil
	}
	var errs error
	if err := c.updateCPUQuotas(); err != nil {
		errs = appendErr(errs, err)
	}
	c.adjustQuota(c.StatusStore)
	if err := c.writeQuota(); err != nil {
		errs = appendErr(errs, err)
	}
	return errs
}

// AlarmWaterMark returns AlarmWaterMark of QuotaTurbo
func (c *Client) AlarmWaterMark() int {
	return c.StatusStore.Config.AlarmWaterMark
}

// HighWaterMark returns HighWaterMark of QuotaTurbo
func (c *Client) HighWaterMark() int {
	return c.StatusStore.Config.HighWaterMark
}

// CgroupRoot returns CgroupRoot of QuotaTurbo
func (c *Client) CgroupRoot() string {
	return c.StatusStore.Config.CgroupRoot
}

// ElevateLimit returns ElevateLimit of QuotaTurbo
func (c *Client) ElevateLimit() float64 {
	return c.StatusStore.Config.ElevateLimit
}

// SlowFallbackRatio returns SlowFallbackRatio of QuotaTurbo
func (c *Client) SlowFallbackRatio() float64 {
	return c.StatusStore.Config.SlowFallbackRatio
}

// CPUFloatingLimit returns CPUFloatingLimit of QuotaTurbo
func (c *Client) CPUFloatingLimit() float64 {
	return c.StatusStore.Config.CPUFloatingLimit
}

// WithWaterMark sets HighWaterMark and AlarmWaterMark of QuotaTurbo
func WithWaterMark(highVal, alarmVal int) Option {
	return func(c *Config) error {
		alarmTmp := c.AlarmWaterMark
		highTmp := c.HighWaterMark
		c.AlarmWaterMark = alarmVal
		c.HighWaterMark = highVal
		if err := c.validateWaterMark(); err != nil {
			c.AlarmWaterMark = alarmTmp
			c.HighWaterMark = highTmp
			return err
		}
		return nil
	}
}

// WithCgroupRoot sets CgroupRoot of QuotaTurbo
func WithCgroupRoot(path string) Option {
	return func(c *Config) error {
		c.CgroupRoot = path
		return nil
	}
}

// WithElevateLimit sets ElevateLimit of QuotaTurbo
func WithElevateLimit(val float64) Option {
	return func(c *Config) error {
		if val < minimumUtilization || val > maximumUtilization {
			return fmt.Errorf("the size range of SingleTotalIncreaseLimit is [0,100]")
		}
		c.ElevateLimit = val
		return nil
	}

}

// WithSlowFallbackRatio sets SlowFallbackRatio of QuotaTurbo
func WithSlowFallbackRatio(val float64) Option {
	return func(c *Config) error {
		c.SlowFallbackRatio = val
		return nil
	}
}

// WithCPUFloatingLimit sets CPUFloatingLimit of QuotaTurbo
func WithCPUFloatingLimit(val float64) Option {
	return func(c *Config) error {
		if val < minimumUtilization || val > maximumUtilization {
			return fmt.Errorf("the size range of SingleTotalIncreaseLimit is [0,100]")
		}
		c.CPUFloatingLimit = val
		return nil
	}
}

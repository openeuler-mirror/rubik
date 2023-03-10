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

	"github.com/hashicorp/go-multierror"
)

// Client is quotaTurbo client
type Client struct {
	*StatusStore
	Driver
}

// NewClient returns a quotaTurbo client instance
func NewClient() *Client {
	return &Client{
		StatusStore: NewStatusStore(),
		Driver:      &EventDriver{},
	}
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
		errs = multierror.Append(errs, err)
	}
	c.adjustQuota(c.StatusStore)
	if err := c.writeQuota(); err != nil {
		errs = multierror.Append(errs, err)
	}
	return errs
}

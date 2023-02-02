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
// Create: 2023-02-01
// Description: This file contains commom cgroup operation

// Package util is common utilitization
package util

import (
	"path/filepath"

	"isula.org/rubik/pkg/common/constant"
)

var (
	// CgroupRoot is the unique cgroup mount point globally
	CgroupRoot = constant.DefaultCgroupRoot
)

// AbsoluteCgroupPath returns absolute cgroup path of specified subsystem of a relative path
func AbsoluteCgroupPath(subsys string, relativePath string) string {
	if subsys == "" || relativePath == "" {
		return ""
	}
	return filepath.Join(CgroupRoot, subsys, relativePath)
}

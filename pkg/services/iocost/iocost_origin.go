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
// Description: This file is used to implement system iocost interface

// Package iocost
package iocost

import (
	"fmt"
	"strconv"

	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	// iocost model file
	iocostModelFile = "blkio.cost.model"
	// iocost weight file
	iocostWeightFile = "blkio.cost.weight"
	// iocost weight qos file
	iocostQosFile = "blkio.cost.qos"
	// cgroup writeback file
	wbBlkioinoFile = "memory.wb_blkio_ino"
)

// configIOCostQoS for config iocost qos.
func configIOCostQoS(devno string, enable bool) error {
	t := 0
	if enable {
		t = 1
	}
	qosParam := fmt.Sprintf("%v enable=%v ctrl=user min=100.00 max=100.00", devno, t)
	return cgroup.WriteCgroupFile(qosParam, blkcgRootDir, iocostQosFile)
}

// configIOCostModel for config iocost model
func configIOCostModel(devno string, p interface{}) error {
	var paramStr string
	switch param := p.(type) {
	case LinearParam:
		if param.Rbps <= 0 || param.Rseqiops <= 0 || param.Rrandiops <= 0 ||
			param.Wbps <= 0 || param.Wseqiops <= 0 || param.Wrandiops <= 0 {
			return fmt.Errorf("invalid params, linear params must be greater than 0")
		}

		paramStr = fmt.Sprintf("%v rbps=%v rseqiops=%v rrandiops=%v wbps=%v wseqiops=%v wrandiops=%v",
			devno,
			param.Rbps, param.Rseqiops, param.Rrandiops,
			param.Wbps, param.Wseqiops, param.Wrandiops,
		)
	default:
		return fmt.Errorf("invalid model param")
	}
	return cgroup.WriteCgroupFile(paramStr, blkcgRootDir, iocostModelFile)
}

// configPodIOCostWeight for config iocost weight
// cgroup v1 iocost cannot be inherited. Therefore, only the container level can be configured.
func configPodIOCostWeight(relativePath string, weight uint64) error {
	if err := cgroup.WriteCgroupFile(strconv.FormatUint(weight, scale), blkcgRootDir,
		relativePath, iocostWeightFile); err != nil {
		return err
	}
	if err := bindMemcgBlkcg(relativePath); err != nil {
		return err
	}
	return nil
}

// bindMemcgBlkcg for bind memcg and blkcg
func bindMemcgBlkcg(containerRelativePath string) error {
	blkcgPath := cgroup.AbsoluteCgroupPath(blkcgRootDir, containerRelativePath)
	ino, err := getDirInode(blkcgPath)
	if err != nil {
		return err
	}

	if err := cgroup.WriteCgroupFile(strconv.FormatUint(ino, scale),
		memcgRootDir, containerRelativePath, wbBlkioinoFile); err != nil {
		return err
	}
	return nil
}

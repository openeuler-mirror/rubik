// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Yanting Song
// Create: 2022-07-19
// Description: quota burst setting for pods

// Package quota is for quota settings
package quota

import (
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
)


// SetPodsQuotaBurst sync pod's burst quota when autoconfig is set
func SetPodsQuotaBurst(podInfos map[string]*typedef.PodInfo) {
	for _, pi := range podInfos {
		setPodQuotaBurst(pi)
	}
}

// UpdatePodQuotaBurst update pod's burst quota
func UpdatePodQuotaBurst(opi, npi *typedef.PodInfo) {
	// cpm.GetPod returns nil if pod.UID not exist
	if opi == nil || npi == nil {
		log.Errorf("quota-burst got invalid nil podInfo")
		return
	}
	if opi.QuotaBurst == npi.QuotaBurst {
		return
	}
	setPodQuotaBurst(npi)
}

// SetPodQuotaBurst set each container's cpu.cfs_burst_ns
func SetPodQuotaBurst(podInfo *typedef.PodInfo) {
	// cpm.GetPod returns nil if pod.UID not exist
	if podInfo == nil {
		log.Errorf("quota-burst got invalid nil podInfo")
		return
	}
	setPodQuotaBurst(podInfo)
}

func setPodQuotaBurst(podInfo *typedef.PodInfo) {
	var invalidBurst int64 = -1
	if podInfo.QuotaBurst == invalidBurst {
		return
	}
	burst := big.NewInt(podInfo.QuotaBurst).String()
	for _, c := range podInfo.Containers {
		err := setCtrQuotaBurst([]byte(burst), c)
		if err != nil {
			log.Errorf("set container quota burst failed: %v", err)
		}
	}
}

func setCtrQuotaBurst(burst []byte, c *typedef.ContainerInfo) error {
	const (
		fname = "cpu.cfs_burst_us"
		subsys = "cpu"
	)
	cgpath := filepath.Join(c.CgroupRoot, subsys, c.CgroupAddr, fname)

	if _, err := os.Stat(cgpath); err != nil && os.IsNotExist(err) {
		return errors.Errorf("quota-burst path=%v missing", cgpath)
	}

	if err := ioutil.WriteFile(cgpath, burst, constant.DefaultFileMode); err != nil {
		return errors.Errorf("quota-burst path=%v setting failed: %v", cgpath, err)
	}
	log.Infof("quota-burst path=%v setting success", cgpath)
	return nil
}

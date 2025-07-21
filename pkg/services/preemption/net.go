// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Niu Qianqian
// Create: 2025-07-01
// Description: Integrating oncn-bwm features

package preemption

import (
	"fmt"
	"os"
	"strconv"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
)

const (
	minWaterline = 20
	maxWaterline = 9999 * 1024
	minBandwidth = 1
	maxBandwidth = 9999 * 1024
)

type NetConfig struct {
	Waterline     int `json:"waterline,omitempty"`
	BandwidthLow  int `json:"bandwidthLow,omitempty"`
	BandwidthHigh int `json:"bandwidthHigh,omitempty"`
}

func getNetLevelStr(qosLevel int) string {
	if qosLevel == constant.Offline {
		return "4294967295" // uint32(-1)
	}
	return "0"
}

func validateNetResConf(conf *PreemptionConfig) error {
	if !isSupportNetqos() {
		return fmt.Errorf("this machine does not support net preemption, please install oncn-bwm first.")
	}

	if conf.Net.Waterline < minWaterline || conf.Net.Waterline > maxWaterline {
		return fmt.Errorf("net waterline %d out of range [%d,%d]", conf.Net.Waterline, minWaterline, maxWaterline)
	}

	for _, per := range []int{
		conf.Net.BandwidthLow, conf.Net.BandwidthHigh} {
		if per < minBandwidth || per > maxBandwidth {
			return fmt.Errorf("net bandwidth %d out of range [%d,%d]", per, minBandwidth, maxBandwidth)
		}
	}

	if conf.Net.BandwidthLow >= conf.Net.BandwidthHigh {
		return fmt.Errorf("net bandwidthLow is larger than bandwidthHigh")
	}

	return nil
}

func enableNetRes(pod *typedef.PodInfo) error {
	var err error
	var pid string

	if pid, err = getPodProcID(pod); err != nil {
		return fmt.Errorf("failed to get Pod procID %s: %v", pid, err)
	}

	if err = enablePodNetqos(pid); err != nil {
		disablePodNetqos(pid)
		return err
	}

	return nil
}

func initNetRes(conf *PreemptionConfig) error {
	var err error
	pid := strconv.Itoa(os.Getpid())
	defer func() {
		if err != nil {
			disablePodNetqos(pid)
		}
	}()
	if err = enablePodNetqos(pid); err != nil {
		return err
	}

	// The bandwidth or waterline can be set only after netqos has been enabled at least once.
	if err = setPodNetqosWaterline(conf.Net.Waterline); err != nil {
		return err
	}

	if err = setPodNetqosBandwidth(conf.Net.BandwidthLow, conf.Net.BandwidthHigh); err != nil {
		return err
	}

	return err
}

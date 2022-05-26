// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Song Yanting
// Create: 2022-6-7
// Description: blkio setting for pods

// Package blkio now only support byte unit.
// For example, limit read operation maximum 10 MBps, set value 10485760
// More units will be supported like MB, KB...
package blkio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/util"
)

const (
	deviceReadBpsFile   = "blkio.throttle.read_bps_device"
	deviceWriteBpsFile  = "blkio.throttle.write_bps_device"
	deviceReadIopsFile  = "blkio.throttle.read_iops_device"
	deviceWriteIopsFile = "blkio.throttle.write_iops_device"
)

// DeviceConfig defines blkio device configurations
type DeviceConfig struct {
	DeviceName  string `json:"device,omitempty"`
	DeviceValue string `json:"value,omitempty"`
}

// BlkConfig defines blkio device configurations
type BlkConfig struct {
	DeviceReadBps   []DeviceConfig `json:"device_read_bps,omitempty"`
	DeviceWriteBps  []DeviceConfig `json:"device_write_bps,omitempty"`
	DeviceReadIops  []DeviceConfig `json:"device_read_iops,omitempty"`
	DeviceWriteIops []DeviceConfig `json:"device_write_iops,omitempty"`
}

// SetBlkio set blkio limtis according to annotation
func SetBlkio(pod *corev1.Pod) {
	cfg := decodeBlkioCfg(pod.Annotations[constant.BlkioKey])
	if cfg == nil {
		return
	}
	blkioLimit(pod, cfg, false)
}

// WriteBlkio updates blkio limtis according to annotation
func WriteBlkio(old *corev1.Pod, new *corev1.Pod) {
	if new.Status.Phase != corev1.PodRunning {
		return
	}

	if old.Annotations[constant.BlkioKey] == new.Annotations[constant.BlkioKey] {
		return
	}

	// empty old blkio limits
	if oldCfg := decodeBlkioCfg(old.Annotations[constant.BlkioKey]); oldCfg != nil {
		blkioLimit(old, oldCfg, true)
	}

	// set new blkio limits
	if newCfg := decodeBlkioCfg(new.Annotations[constant.BlkioKey]); newCfg != nil {
		blkioLimit(new, newCfg, false)
	}
}

func blkioLimit(pod *corev1.Pod, cfg *BlkConfig, empty bool) {
	if len(cfg.DeviceReadBps) > 0 {
		tryWriteBlkioLimit(pod, cfg.DeviceReadBps, deviceReadBpsFile, empty)
	}
	if len(cfg.DeviceWriteBps) > 0 {
		tryWriteBlkioLimit(pod, cfg.DeviceWriteBps, deviceWriteBpsFile, empty)
	}
	if len(cfg.DeviceReadIops) > 0 {
		tryWriteBlkioLimit(pod, cfg.DeviceReadIops, deviceReadIopsFile, empty)
	}
	if len(cfg.DeviceWriteIops) > 0 {
		tryWriteBlkioLimit(pod, cfg.DeviceWriteIops, deviceWriteIopsFile, empty)
	}
}

func decodeBlkioCfg(blkioCfg string) *BlkConfig {
	if len(blkioCfg) == 0 {
		return nil
	}
	log.Infof("blkioCfg is %v", blkioCfg)
	cfg := &BlkConfig{
		DeviceReadBps:   []DeviceConfig{},
		DeviceWriteBps:  []DeviceConfig{},
		DeviceReadIops:  []DeviceConfig{},
		DeviceWriteIops: []DeviceConfig{},
	}
	reader := bytes.NewReader([]byte(blkioCfg))
	if err := json.NewDecoder(reader).Decode(cfg); err != nil {
		log.Errorf("decode blkioCfg failed with error: %v", err)
		return nil
	}
	return cfg
}

func tryWriteBlkioLimit(pod *corev1.Pod, devCfgs []DeviceConfig, deviceFilePath string, empty bool) {
	for _, devCfg := range devCfgs {
		devName, devLimit := devCfg.DeviceName, devCfg.DeviceValue

		fi, err := os.Stat(devName)
		if err != nil {
			log.Errorf("stat %s failed with error %v", devName, err)
			continue
		}
		if fi.Mode()&os.ModeDevice == 0 {
			log.Errorf("%s is not a device", devName)
			continue
		}

		if st, ok := fi.Sys().(*syscall.Stat_t); ok {
			devno := st.Rdev
			major, minor := devno/256, devno%256
			var limit string
			if empty == true {
				limit = fmt.Sprintf("%v:%v 0", major, minor)
			} else {
				limit = fmt.Sprintf("%v:%v %s", major, minor, devLimit)
			}
			writeBlkioLimit(pod, limit, deviceFilePath)
		} else {
			log.Errorf("failed to get Sys(), %v has type %v", devName, st)
		}
	}
}

func writeBlkioLimit(pod *corev1.Pod, limit, deviceFilePath string) {
	const (
		dockerPrefix = "docker://"
		containerdPrefix = "containerd://"
		blkioPath = "blkio"
	)
	podCgroupPath := util.GetPodCgroupPath(pod)
	for _, container := range pod.Status.ContainerStatuses {
		containerID := strings.TrimPrefix(container.ContainerID, dockerPrefix)
		containerID = strings.TrimPrefix(containerID, containerdPrefix)
		containerPath := filepath.Join(podCgroupPath, containerID)
		containerBlkFilePath := filepath.Join(config.CgroupRoot, blkioPath, containerPath, deviceFilePath)

		err := ioutil.WriteFile(containerBlkFilePath, []byte(limit), constant.DefaultFileMode)
		if err != nil {
			log.Errorf("writeBlkioLimit write %v to %v failed with error: %v", limit, containerBlkFilePath, err)
			continue
		}
		log.Infof("writeBlkioLimit write %s to %v success", limit, containerBlkFilePath)
	}
}

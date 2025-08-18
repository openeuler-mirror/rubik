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
// Description: This file is used to implement iolimit

// Package iolimit provides io-limit feature for container cgroup management.
package iolimit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services/helper"
)

// convertToMajorMinorFunc is a function variable that can be replaced in tests
var convertToMajorMinorFunc = convertToMajorMinorImpl

const (
	blkcgRootDir = "blkio"
)

const (
	deviceReadBpsFile   = "blkio.throttle.read_bps_device"
	deviceWriteBpsFile  = "blkio.throttle.write_bps_device"
	deviceReadIopsFile  = "blkio.throttle.read_iops_device"
	deviceWriteIopsFile = "blkio.throttle.write_iops_device"
)

// DeviceConfig defines blkio device configurations.
type DeviceConfig struct {
	DeviceName  string `json:"device,omitempty"`
	DeviceValue string `json:"value,omitempty"`
}

// IOLimitFactory is the factory for creating IOLimit instances.
type IOLimitFactory struct {
	ObjName string
}

// IOLimit is the service implementation for container IO limiting.
type IOLimit struct {
	helper.ServiceBase
}

// BlkConfig defines blkio device configurations for all throttle types.
type BlkConfig struct {
	DeviceReadBps   []DeviceConfig `json:"device_read_bps,omitempty"`
	DeviceWriteBps  []DeviceConfig `json:"device_write_bps,omitempty"`
	DeviceReadIops  []DeviceConfig `json:"device_read_iops,omitempty"`
	DeviceWriteIops []DeviceConfig `json:"device_write_iops,omitempty"`
}

// Name returns the IOLimit factory name.
func (i IOLimitFactory) Name() string {
	return "IOLimitFactory"
}

// NewObj creates a new object of IOLimit.
func (i IOLimitFactory) NewObj() (interface{}, error) {
	return &IOLimit{
		ServiceBase: *helper.NewServiceBase(i.ObjName),
	}, nil
}

// SetConfig sets the configuration of IOLimit.
// Currently no configuration is needed, so it just returns nil.
func (i *IOLimit) SetConfig(f helper.ConfigHandler) error {
	return nil
}

// PreStart performs pre-start work for IOLimit.
// It configures IO limits for all existing pods.
func (i *IOLimit) PreStart(viewer api.Viewer) error {
	if viewer == nil {
		return fmt.Errorf("invalid pods viewer")
	}
	pods := viewer.ListPodsWithOptions()
	for _, pod := range pods {
		if err := i.configIOLimit(pod); err != nil {
			return fmt.Errorf("failed to config io limit for pod %s: %v", pod.Name, err)
		}
	}
	return nil
}

// Terminate performs cleanup work for IOLimit.
// Currently no cleanup is needed, so it just returns nil.
func (i *IOLimit) Terminate(_ api.Viewer) error {
	// nothing to do here, just return nil.
	return nil
}

// AddPod adds a pod to IOLimit and configures its IO limits.
func (i *IOLimit) AddPod(podInfo *typedef.PodInfo) error {
	if podInfo == nil {
		return fmt.Errorf("invalid pod info")
	}
	return i.configIOLimit(podInfo)
}

// UpdatePod updates a pod in IOLimit and reconfigures its IO limits.
func (i *IOLimit) UpdatePod(old, new *typedef.PodInfo) error {
	if new == nil {
		return fmt.Errorf("invalid pod info")
	}
	return i.configIOLimit(new)
}

// DeletePod removes a pod from IOLimit.
// Currently no cleanup is needed for deletion, so it just returns nil.
func (i *IOLimit) DeletePod(podInfo *typedef.PodInfo) error {
	// nothing to do here, just return nil.
	return nil
}

// configIOLimit configures IO limits for a specific pod.
// It first clears existing throttle configurations, then applies new ones if specified.
func (i *IOLimit) configIOLimit(podInfo *typedef.PodInfo) error {
	cfgString := podInfo.Annotations[constant.BlkioKey]
	if len(cfgString) == 0 {
		log.Infof("pod %s does not have blkio config, skip", podInfo.Name)
		return nil
	}

	// firstly clear all config
	// blkio cgroup hierarchical is not enabled default, only set container cgroups
	for _, container := range podInfo.IDContainersMap {
		if err := clearAllBlkioThrottleFiles(container.Path); err != nil {
			return fmt.Errorf("failed to clear blkio throttle files for container %s of pod %s: %v", container.Name, podInfo.Name, err)
		}
	}

	// secondly parse the config
	cfg, err := parseIOLimitConfig(cfgString)
	if err != nil {
		return fmt.Errorf("parse blkio config for pod %s failed: %v", podInfo.Name, err)
	}

	// thirdly apply the config to cgroup files
	for _, container := range podInfo.IDContainersMap {
		if err := applyIOLimitConfig(container.Path, cfg); err != nil {
			return fmt.Errorf("failed to apply blkio config for container %s of pod %s: %v", container.Name, podInfo.Name, err)
		}
	}

	return nil
}

// parseIOLimitConfig parses the blkio configuration string into a BlkConfig struct.
// The input string should be in JSON format representing the blkio configuration.
// It returns a BlkConfig struct or an error if parsing fails.
func parseIOLimitConfig(blkioCfg string) (*BlkConfig, error) {
	if len(blkioCfg) == 0 {
		return nil, fmt.Errorf("blkio config is empty")
	}

	cfg := &BlkConfig{
		DeviceReadBps:   []DeviceConfig{},
		DeviceWriteBps:  []DeviceConfig{},
		DeviceReadIops:  []DeviceConfig{},
		DeviceWriteIops: []DeviceConfig{},
	}
	reader := bytes.NewReader([]byte(blkioCfg))
	if err := json.NewDecoder(reader).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode blkio config failed: %v", err)
	}
	return cfg, nil
}

// clearAllBlkioThrottleFiles clears all 4 blkio throttle files for a given cgroup path.
// This resets all device throttle configurations to default values.
func clearAllBlkioThrottleFiles(cgroupPath string) error {
	files := []string{
		deviceReadBpsFile,
		deviceWriteBpsFile,
		deviceReadIopsFile,
		deviceWriteIopsFile,
	}

	for _, file := range files {
		if err := clearConfig(cgroupPath, file); err != nil {
			return fmt.Errorf("failed to clear %s: %v", file, err)
		}
	}
	log.Infof("successfully cleared all blkio throttle files for cgroup %s", cgroupPath)
	return nil
}

// clearConfig clears a specific blkio throttle file by resetting all device values to 0.
func clearConfig(cgroupPath, file string) error {
	params, err := cgroup.ReadCgroupFile(blkcgRootDir, cgroupPath, file)
	if err != nil {
		return fmt.Errorf("read cgroup file %s failed: %v", file, err)
	}
	if len(params) == 0 {
		log.Infof("cgroup file %s is empty, skip", file)
		return nil
	}

	// Parse params and reset values to 0
	// Format is typically "major:minor value", we need to set value to 0
	resetContent := parseAndResetParams(string(params))

	// Write the reset content back to the file
	if err := cgroup.WriteCgroupFile(resetContent, blkcgRootDir, cgroupPath, file); err != nil {
		return fmt.Errorf("reset cgroup file %s failed: %v", file, err)
	}
	log.Infof("successfully reset cgroup file %s", file)
	return nil
}

// parseAndResetParams parses the cgroup params and resets all values to 0.
// Input format: "major:minor value" (e.g., "8:0 1048576")
// Output format: "major:minor 0" (e.g., "8:0 0")
func parseAndResetParams(params string) string {
	if len(params) == 0 {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(params), "\n")
	var resetLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Split by space to separate device and value
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// Keep device part (major:minor) and set value to 0
			device := parts[0]
			resetLine := fmt.Sprintf("%s 0", device)
			resetLines = append(resetLines, resetLine)
		}
	}

	return strings.Join(resetLines, "\n")
}

// applyIOLimitConfig applies the parsed BlkConfig to all corresponding cgroup files.
func applyIOLimitConfig(cgroupPath string, cfg *BlkConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Define the device config mappings
	deviceConfigs := []struct {
		fileName    string
		devices     []DeviceConfig
		description string
	}{
		{deviceReadBpsFile, cfg.DeviceReadBps, "device read bps"},
		{deviceWriteBpsFile, cfg.DeviceWriteBps, "device write bps"},
		{deviceReadIopsFile, cfg.DeviceReadIops, "device read iops"},
		{deviceWriteIopsFile, cfg.DeviceWriteIops, "device write iops"},
	}

	// Apply all device configs in a loop
	for _, config := range deviceConfigs {
		if err := applyDeviceConfig(cgroupPath, config.fileName, config.devices); err != nil {
			return fmt.Errorf("failed to apply %s config: %v", config.description, err)
		}
	}
	return nil
}

// applyDeviceConfig applies device configurations to a specific cgroup throttle file.
func applyDeviceConfig(cgroupPath, fileName string, devices []DeviceConfig) error {
	if len(devices) == 0 {
		log.Infof("no device config for file %s, skip", fileName)
		return nil
	}

	var configLines []string
	for _, device := range devices {
		if device.DeviceName == "" || device.DeviceValue == "" {
			log.Warnf("invalid device config: device=%s, value=%s", device.DeviceName, device.DeviceValue)
			continue
		}

		// Convert device name to major:minor format if needed
		majorMinor, err := convertToMajorMinor(device.DeviceName)
		if err != nil {
			log.Warnf("failed to convert device %s to major:minor format: %v", device.DeviceName, err)
			continue
		}

		configLine := fmt.Sprintf("%s %s", majorMinor, device.DeviceValue)
		configLines = append(configLines, configLine)
	}

	if len(configLines) == 0 {
		log.Infof("no valid device config for file %s, skip", fileName)
		return nil
	}

	configContent := strings.Join(configLines, "\n")
	if err := cgroup.WriteCgroupFile(configContent, blkcgRootDir, cgroupPath, fileName); err != nil {
		return fmt.Errorf("failed to write config to file %s: %v", fileName, err)
	}

	return nil
}

// convertToMajorMinor converts device name to major:minor format.
// It supports both device paths (e.g., /dev/sda) and existing major:minor format (e.g., 8:0).
func convertToMajorMinor(deviceName string) (string, error) {
	return convertToMajorMinorFunc(deviceName)
}

// convertToMajorMinorImpl is the actual implementation of convertToMajorMinor.
// This can be replaced with a mock in tests.
func convertToMajorMinorImpl(deviceName string) (string, error) {
	// Try to parse as numeric major:minor if it contains only digits and ':'
	if strings.Count(deviceName, ":") == 1 {
		parts := strings.Split(deviceName, ":")
		if len(parts) == 2 {
			if _, err := strconv.Atoi(parts[0]); err == nil {
				if _, err := strconv.Atoi(parts[1]); err == nil {
					return deviceName, nil
				}
			}
		}
	}

	// If it's a device path like /dev/sda, get its major:minor
	if strings.HasPrefix(deviceName, "/dev/") {
		stat, err := os.Stat(deviceName)
		if err != nil {
			return "", fmt.Errorf("failed to stat device %s: %v", deviceName, err)
		}

		// Get the device numbers from file info
		if stat.Mode()&os.ModeDevice != 0 {
			sys := stat.Sys()
			if sysstat, ok := sys.(*syscall.Stat_t); ok {
				major := unix.Major(sysstat.Rdev)
				minor := unix.Minor(sysstat.Rdev)
				return fmt.Sprintf("%d:%d", major, minor), nil
			}
		}
		return "", fmt.Errorf("device %s is not a block device", deviceName)
	}

	return "", fmt.Errorf("unsupported device name format: %s", deviceName)
}

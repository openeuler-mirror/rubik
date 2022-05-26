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
// Description: blkio test
package blkio

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/typedef"
	"isula.org/rubik/pkg/util"
)

const (
	blkioAnnotation = `{"device_read_bps":[{"device":"/dev/sda1","value":"52428800"}, {"device":"/dev/sda","value":"105857600"}],
"device_write_bps":[{"device":"/dev/sda1","value":"105857600"}],
"device_read_iops":[{"device":"/dev/sda1","value":"200"}],
"device_write_iops":[{"device":"/dev/sda1","value":"300"}]}`
	invBlkioAnnotation = `{"device_read_bps":[{"device":"/dev/sda1","value":"52428800"}, {"device":"/dev/sda","value":"105857600"}}`
)

var (
	devicePaths = map[string]string{
		"device_read_bps":   deviceReadBpsFile,
		"device_write_bps":  deviceWriteBpsFile,
		"device_read_iops":  deviceReadIopsFile,
		"device_write_iops": deviceWriteIopsFile,
	}
	status = corev1.PodStatus{
		ContainerStatuses: []corev1.ContainerStatus{
			{ContainerID: "docker://aaa"},
		},
		QOSClass: corev1.PodQOSBurstable,
		Phase:    corev1.PodRunning,
	}
	containerDir = filepath.Join(constant.TmpTestDir, "blkio/kubepods/burstable/podaaa/aaa")
)

func getMajor(devName string) (major int, err error) {
	cmd := fmt.Sprintf("ls -l %v | awk -F ' ' '{print $5}'", devName)
	out, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		return -1, err
	}
	o := strings.TrimSuffix(strings.TrimSpace(string(out)), ",")
	return strconv.Atoi(o)
}

func getMinor(devName string) (minor int, err error) {
	cmd := fmt.Sprintf("ls -l %v | awk -F ' ' '{print $6}'", devName)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return -1, err
	}
	o := strings.TrimSuffix(strings.TrimSpace(string(out)), "\n")
	return strconv.Atoi(o)
}

func TestBlkioAnnotation1(t *testing.T) {
	// valid blkiocfg format
	cfg := decodeBlkioCfg(blkioAnnotation)
	assert.True(t, len(cfg.DeviceReadBps) > 0)
	assert.True(t, len(cfg.DeviceWriteBps) > 0)
	assert.True(t, len(cfg.DeviceReadIops) > 0)
	assert.True(t, len(cfg.DeviceWriteIops) > 0)

	assert.Equal(t, "/dev/sda1", cfg.DeviceReadBps[0].DeviceName)
	assert.Equal(t, "52428800", cfg.DeviceReadBps[0].DeviceValue)
	assert.Equal(t, "/dev/sda", cfg.DeviceReadBps[1].DeviceName)
	assert.Equal(t, "105857600", cfg.DeviceReadBps[1].DeviceValue)

	assert.Equal(t, "/dev/sda1", cfg.DeviceWriteBps[0].DeviceName)
	assert.Equal(t, "105857600", cfg.DeviceWriteBps[0].DeviceValue)

	assert.Equal(t, "/dev/sda1", cfg.DeviceReadIops[0].DeviceName)
	assert.Equal(t, "200", cfg.DeviceReadIops[0].DeviceValue)

	assert.Equal(t, "/dev/sda1", cfg.DeviceWriteIops[0].DeviceName)
	assert.Equal(t, "300", cfg.DeviceWriteIops[0].DeviceValue)

	// invalid blkiocfg format
	cfg = decodeBlkioCfg(invBlkioAnnotation)
	assert.Equal(t, (*BlkConfig)(nil), cfg)
}

func TestBlkioAnnotation2(t *testing.T) {
	// valid blkiocfg format, valid + invalid device name
	s1 := `{"device_read_bps":[{"device":"/dev/sda1","value":"10485760"}, {"device":"/dev/sda","value":"10485760"}],
			"device_read_iops":[{"device":"/dev/sda1","value":"200"}, {"device":"/dev/123","value":"123"}]}`
	cfg := decodeBlkioCfg(s1)
	assert.True(t, len(cfg.DeviceReadBps) == 2)
	assert.True(t, len(cfg.DeviceWriteBps) == 0)
	assert.True(t, len(cfg.DeviceReadIops) == 2)
	assert.True(t, len(cfg.DeviceWriteIops) == 0)

	assert.Equal(t, "/dev/sda1", cfg.DeviceReadBps[0].DeviceName)
	assert.Equal(t, "10485760", cfg.DeviceReadBps[0].DeviceValue)
	assert.Equal(t, "/dev/sda", cfg.DeviceReadBps[1].DeviceName)
	assert.Equal(t, "10485760", cfg.DeviceReadBps[1].DeviceValue)

	assert.Equal(t, "/dev/sda1", cfg.DeviceReadIops[0].DeviceName)
	assert.Equal(t, "200", cfg.DeviceReadIops[0].DeviceValue)
	assert.Equal(t, "/dev/123", cfg.DeviceReadIops[1].DeviceName)
	assert.Equal(t, "123", cfg.DeviceReadIops[1].DeviceValue)
}

func listDevices() []string {
	dir, _ := ioutil.ReadDir("/sys/block")
	devices := make([]string, 0, len(dir))
	for _, f := range dir {
		devices = append(devices, f.Name())
	}
	return devices
}

func genarateRandDev(n int, devices []string) string {
	const bytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
	b := make([]byte, n)
	valid := true
	for valid {
		valid = false
		for i := range b {
			b[i] = bytes[typedef.RandInt(len(bytes))]
		}
		for _, device := range devices {
			if device == string(b) {
				valid = true
			}
		}
	}
	return string(b)
}

func mkdirHelper(t *testing.T) {
	assert.NoError(t, os.RemoveAll(constant.TmpTestDir))
	for _, fname := range []string{deviceReadBpsFile, deviceWriteBpsFile, deviceReadIopsFile, deviceWriteIopsFile} {
		assert.NoError(t, util.CreateFile(filepath.Join(containerDir, fname)))
	}
}

func TestSetBlkio(t *testing.T) {
	testFunc := func(s1, deviceType, devicePath string, devices []string) {
		mkdirHelper(t)
		defer os.RemoveAll(constant.TmpTestDir)

		config.CgroupRoot = constant.TmpTestDir
		pod := &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				UID: "aaa",
				Annotations: map[string]string{
					constant.BlkioKey: s1,
				},
			},
			Status: status,
		}

		SetBlkio(pod)

		expected := ""
		for _, device := range devices {
			major, err := getMajor("/dev/" + device)
			if major == 0 || err != nil {
				continue
			}
			minor, err := getMinor("/dev/" + device)
			if err != nil {
				continue
			}
			expected += fmt.Sprintf("%v:%v %v", major, minor, 10485760)
		}
		actual, _ := ioutil.ReadFile(filepath.Join(containerDir, devicePath))

		assert.Equal(t, expected, strings.TrimSuffix(string(actual), "\n"))
	}

	// test valid devices names from /sys/block
	devices := listDevices()
	for deviceType, devicePath := range devicePaths {
		for _, device := range devices {
			cfg := `{"` + deviceType + `":[{"device":"/dev/` + device + `","value":"10485760"}]}`
			testFunc(cfg, deviceType, devicePath, []string{device})
		}
	}

	// test invalid device names from random generated characters
	for deviceType, devicePath := range devicePaths {
		invalidDeviceName := genarateRandDev(3, devices)
		cfg := `{"` + deviceType + `":[{"device":"/dev/` + invalidDeviceName + `","value":"10485760"}]}`
		testFunc(cfg, deviceType, devicePath, []string{invalidDeviceName})

	}

	// test valid devices names + invalid devices names
	for deviceType, devicePath := range devicePaths {
		for _, device := range devices {
			invalidDeviceName := genarateRandDev(3, devices)
			cfg := `{"` + deviceType + `":[{"device":"/dev/` + device + `","value":"10485760"}, {"device":"/dev/` + invalidDeviceName + `","value":"10485760"}]}`
			testFunc(cfg, deviceType, devicePath, []string{device, invalidDeviceName})
		}
	}
}

func TestWriteBlkio(t *testing.T) {
	mkdirHelper(t)
	defer os.RemoveAll(constant.TmpTestDir)

	config.CgroupRoot = constant.TmpTestDir
	old := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			UID: "aaa",
			Annotations: map[string]string{
				constant.BlkioKey: "",
			},
		},
		Status: status,
	}
	SetBlkio(&old)

	testFunc := func(newCfg, deviceType, devicePath string, devices []string) {
		new := corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				UID: "aaa",
				Annotations: map[string]string{
					constant.BlkioKey: newCfg,
				},
			},
			Status: status,
		}
		WriteBlkio(&old, &new)

		old = new
		expected := ""
		for _, device := range devices {
			major, _ := getMajor("/dev/" + device)
			if major == 0 {
				continue
			}
			minor, _ := getMinor("/dev/" + device)
			expected += fmt.Sprintf("%v:%v %v", major, minor, 10485760)
		}
		file := filepath.Join(containerDir, devicePath)
		actual, _ := ioutil.ReadFile(file)

		assert.Equal(t, expected, strings.TrimSuffix(string(actual), "\n"))
	}

	// test valid devices names from /sys/block
	devices := listDevices()
	for deviceType, devicePath := range devicePaths {
		for _, device := range devices {
			newCfg := `{"` + deviceType + `":[{"device":"/dev/` + device + `","value":"10485760"}]}`
			testFunc(newCfg, deviceType, devicePath, []string{device})
		}
	}
}

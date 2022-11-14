// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: hanchao
// Create: 2022-10-28
// Description: iocost test

// Package iocost is for iocost.
package iocost

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/try"
	"isula.org/rubik/pkg/typedef"
)

const paramsLen = 2

func TestIOcostFeatureSwitch(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	SetIOcostEnable(false)
	err := ConfigIOcost(nil)
	assert.Equal(t, err.Error(), "iocost feature is disable")
	err = SetPodWeight(nil)
	assert.Equal(t, err.Error(), "iocost feature is disable")
	err = ShutDown()
	assert.Equal(t, err.Error(), "iocost feature is disable")
}

// TestIocostConfig is testing for IocostConfig interface.
func TestIocostConfig(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	SetIOcostEnable(true)
	devs, err := getAllBlockDevice()
	assert.NoError(t, err)
	var devname, devno string
	for k, v := range devs {
		devname = k
		devno = v
		break
	}

	tests := []struct {
		name       string
		config     config.IOcostConfig
		qosCheck   bool
		modelCheck bool
		qosParam   string
		modelParam string
	}{
		{
			name: "Test iocost enable",
			config: config.IOcostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: config.Param{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: true,
			qosParam:   devno + " enable=1",
			modelParam: devno + " ctrl=user model=linear " +
				"rbps=600 rseqiops=600 rrandiops=600 " +
				"wbps=600 wseqiops=600 wrandiops=600",
		},
		{
			name: "Test iocost disable",
			config: config.IOcostConfig{
				Dev:    devname,
				Enable: false,
				Model:  "linear",
				Param: config.Param{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
		{
			name: "Test iocost enable",
			config: config.IOcostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: config.Param{
					Rbps: 500, Rseqiops: 500, Rrandiops: 500,
					Wbps: 500, Wseqiops: 500, Wrandiops: 500,
				},
			},
			qosCheck:   true,
			modelCheck: true,
			qosParam:   devno + " enable=1",
			modelParam: devno + " ctrl=user model=linear " +
				"rbps=500 rseqiops=500 rrandiops=500 " +
				"wbps=500 wseqiops=500 wrandiops=500",
		},
		{
			name: "Test iocost no dev error",
			config: config.IOcostConfig{
				Dev:    "xxx",
				Enable: true,
				Model:  "linear",
				Param: config.Param{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
		{
			name: "Test iocost enable",
			config: config.IOcostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: config.Param{
					Rbps: 500, Rseqiops: 500, Rrandiops: 500,
					Wbps: 500, Wseqiops: 500, Wrandiops: 500,
				},
			},
			qosCheck:   true,
			modelCheck: true,
			qosParam:   devno + " enable=1",
			modelParam: devno + " ctrl=user model=linear " +
				"rbps=500 rseqiops=500 rrandiops=500 " +
				"wbps=500 wseqiops=500 wrandiops=500",
		},
		{
			name: "Test iocost non-linear error",
			config: config.IOcostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linearx",
				Param: config.Param{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
		{
			name: "Test iocost enable",
			config: config.IOcostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: config.Param{
					Rbps: 500, Rseqiops: 500, Rrandiops: 500,
					Wbps: 500, Wseqiops: 500, Wrandiops: 500,
				},
			},
			qosCheck:   true,
			modelCheck: true,
			qosParam:   devno + " enable=1",
			modelParam: devno + " ctrl=user model=linear " +
				"rbps=500 rseqiops=500 rrandiops=500 " +
				"wbps=500 wseqiops=500 wrandiops=500",
		},
		{
			name: "Test iocost param error",
			config: config.IOcostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: config.Param{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 0,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := []config.IOcostConfig{
				tt.config,
			}
			err := ConfigIOcost(params)
			assert.NoError(t, err)
			if tt.qosCheck {
				filePath := filepath.Join(config.CgroupRoot, blkSubName, iocostQosFile)
				qosParamByte, err := ioutil.ReadFile(filePath)
				assert.NoError(t, err)
				qosParams := strings.Split(string(qosParamByte), "\n")
				for _, qosParam := range qosParams {
					paramList := strings.FieldsFunc(qosParam, unicode.IsSpace)
					if len(paramList) >= paramsLen && strings.Compare(paramList[0], devno) == 0 {
						assert.Equal(t, tt.qosParam, qosParam[:len(tt.qosParam)])
						break
					}
				}
			}
			if tt.modelCheck {
				filePath := filepath.Join(config.CgroupRoot, blkSubName, iocostModelFile)
				modelParamByte, err := ioutil.ReadFile(filePath)
				assert.NoError(t, err)
				modelParams := strings.Split(string(modelParamByte), "\n")
				for _, modelParam := range modelParams {
					paramList := strings.FieldsFunc(modelParam, unicode.IsSpace)
					if len(paramList) >= paramsLen && strings.Compare(paramList[0], devno) == 0 {
						assert.Equal(t, tt.modelParam, modelParam[:len(tt.modelParam)])
						break
					}
				}
			}
		})
	}
}

// TestSetPodWeight is testing for SetPodWeight interface.
func TestSetPodWeight(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	// deploy enviroment
	const testCgroupPath = "/rubik-test"
	rubikBlkioTestPath := filepath.Join(config.CgroupRoot, blkSubName, testCgroupPath)
	rubikMemTestPath := filepath.Join(config.CgroupRoot, memSubName, testCgroupPath)
	try.MkdirAll(rubikBlkioTestPath, constant.DefaultDirMode)
	try.MkdirAll(rubikMemTestPath, constant.DefaultDirMode)
	defer try.RemoveAll(rubikBlkioTestPath)
	defer try.RemoveAll(rubikMemTestPath)
	SetIOcostEnable(true)

	tests := []struct {
		name    string
		pod     *typedef.PodInfo
		wantErr bool
		want    string
	}{
		{
			name: "Test online qos level",
			pod: &typedef.PodInfo{
				CgroupRoot: config.CgroupRoot,
				CgroupPath: testCgroupPath,
				Offline:    false,
			},
			wantErr: false,
			want:    "default 1000\n",
		},
		{
			name: "Test offline qos level",
			pod: &typedef.PodInfo{
				CgroupRoot: config.CgroupRoot,
				CgroupPath: testCgroupPath,
				Offline:    true,
			},
			wantErr: false,
			want:    "default 10\n",
		},
		{
			name: "Test error cgroup path",
			pod: &typedef.PodInfo{
				CgroupRoot: config.CgroupRoot,
				CgroupPath: "var/log/rubik/rubik-test",
				Offline:    true,
			},
			wantErr: true,
			want:    "default 10\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetPodWeight(tt.pod)
			if tt.wantErr {
				assert.Equal(t, err != nil, true)
				return
			}
			assert.NoError(t, err)
			weightFile := filepath.Join(rubikBlkioTestPath, "blkio.cost.weight")
			weightOnline, err := ioutil.ReadFile(weightFile)
			assert.NoError(t, err)
			assert.Equal(t, string(weightOnline), tt.want)
		})
	}
}

// TestBindMemcgBlkio is testing for bindMemcgBlkio
func TestBindMemcgBlkio(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	// deploy enviroment
	const testCgroupPath = "rubik-test"
	rubikBlkioTestPath := filepath.Join(config.CgroupRoot, blkSubName, testCgroupPath)
	rubikMemTestPath := filepath.Join(config.CgroupRoot, memSubName, testCgroupPath)
	try.MkdirAll(rubikBlkioTestPath, constant.DefaultDirMode)
	try.MkdirAll(rubikMemTestPath, constant.DefaultDirMode)
	defer try.RemoveAll(rubikBlkioTestPath)
	defer try.RemoveAll(rubikMemTestPath)
	SetIOcostEnable(true)

	containers := make(map[string]*typedef.ContainerInfo, 5)
	for i := 0; i < 5; i++ {
		dirName := "container" + strconv.Itoa(i)
		blkContainer := filepath.Join(rubikBlkioTestPath, dirName)
		memContainer := filepath.Join(rubikMemTestPath, dirName)
		try.MkdirAll(blkContainer, constant.DefaultDirMode)
		try.MkdirAll(memContainer, constant.DefaultDirMode)
		containers[dirName] = &typedef.ContainerInfo{
			Name:       dirName,
			CgroupRoot: config.CgroupRoot,
			CgroupAddr: filepath.Join(testCgroupPath, dirName),
		}
	}
	err := bindMemcgBlkio(containers)
	assert.NoError(t, err)

	for key, container := range containers {
		memcgPath := container.CgroupPath(memSubName)
		blkcgPath := container.CgroupPath(blkSubName)
		ino, err := getDirInode(blkcgPath)
		assert.NoError(t, err)
		wbBlkioInfo := filepath.Join(memcgPath, wbBlkioinoFile)
		blkioInoStr, err := ioutil.ReadFile(wbBlkioInfo)
		assert.NoError(t, err)
		result := fmt.Sprintf("wb_blkio_path:/%v/%v\nwb_blkio_ino:%v\n", testCgroupPath, key, ino)
		assert.Equal(t, result, string(blkioInoStr))
	}
}

// TestClearIOcost is testing for ClearIOcost interface.
func TestClearIOcost(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	devs, err := getAllBlockDevice()
	assert.NoError(t, err)
	filePath := filepath.Join(config.CgroupRoot, blkSubName, iocostQosFile)
	for _, devno := range devs {
		qosStr := fmt.Sprintf("%v enable=1", devno)
		err := writeIOcost(filePath, qosStr)
		assert.NoError(t, err)
	}
	err = ShutDown()
	assert.NoError(t, err)
	qosParamByte, err := ioutil.ReadFile(filePath)
	assert.NoError(t, err)
	qosParams := strings.Split(string(qosParamByte), "\n")
	for _, qosParam := range qosParams {
		paramList := strings.FieldsFunc(qosParam, unicode.IsSpace)
		if len(paramList) >= paramsLen {
			assert.Equal(t, paramList[1], "enable=0")
		}
	}
}

// TestGetBlkDevice is testing for get block device interface.
func TestGetBlkDevice(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	devs, err := getAllBlockDevice()
	assert.NoError(t, err)
	for index, dev := range devs {
		devno, err := getBlkDeviceNo(index)
		assert.NoError(t, err)
		assert.Equal(t, dev, devno)
	}
	_, err = getBlkDeviceNo("")
	assert.Equal(t, err != nil, true)
	_, err = getBlkDeviceNo("xxx")
	assert.Equal(t, err != nil, true)
}

// TestConfigQos is testing for ConfigQos interface.
func TestConfigQos(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	SetIOcostEnable(true)
	devs, err1 := getAllBlockDevice()
	assert.NoError(t, err1)
	var devno string
	for _, v := range devs {
		devno = v
		break
	}

	tests := []struct {
		name   string
		enable bool
		want   string
	}{
		{
			name:   "Test qos disable",
			enable: false,
			want:   devno + " enable=0",
		},
		{
			name:   "Test qos enable",
			enable: true,
			want:   devno + " enable=1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configQos(tt.enable, devno)
			assert.NoError(t, err)
			filePath := filepath.Join(config.CgroupRoot, blkSubName, iocostQosFile)
			qosParamByte, err := ioutil.ReadFile(filePath)
			assert.NoError(t, err)
			qosParams := strings.Split(string(qosParamByte), "\n")
			for _, qosParam := range qosParams {
				paramList := strings.FieldsFunc(qosParam, unicode.IsSpace)
				if len(paramList) >= paramsLen && strings.Compare(paramList[0], devno) == 0 {
					assert.Equal(t, tt.want, qosParam[:len(tt.want)])
					break
				}
			}
		})
	}
}

// TestConfigLinearModel is testing for ConfigLinearModel interface.
func TestConfigLinearModel(t *testing.T) {
	if !HwSupport() {
		t.Skipf("%s only run on support iocost machine", t.Name())
	}

	SetIOcostEnable(true)
	devs, err := getAllBlockDevice()
	assert.NoError(t, err)
	var devno string
	for _, v := range devs {
		devno = v
		break
	}

	tests := []struct {
		name        string
		linearModel config.Param
		wantErr     bool
		modelParam  string
	}{
		{
			name: "Test linear model",
			linearModel: config.Param{
				Rbps: 500, Rseqiops: 500, Rrandiops: 500,
				Wbps: 500, Wseqiops: 500, Wrandiops: 500,
			},
			wantErr: false,
			modelParam: devno + " ctrl=user model=linear " +
				"rbps=500 rseqiops=500 rrandiops=500 " +
				"wbps=500 wseqiops=500 wrandiops=500",
		},
		{
			name: "Test linear model",
			linearModel: config.Param{
				Rbps: 600, Rseqiops: 600, Rrandiops: 600,
				Wbps: 600, Wseqiops: 600, Wrandiops: 600,
			},
			wantErr: false,
			modelParam: devno + " ctrl=user model=linear " +
				"rbps=600 rseqiops=600 rrandiops=600 " +
				"wbps=600 wseqiops=600 wrandiops=600",
		},
		{
			name: "Test missing parameter",
			linearModel: config.Param{
				Rseqiops: 600, Rrandiops: 600,
				Wbps: 600, Wseqiops: 600, Wrandiops: 600,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configLinearModel(tt.linearModel, devno)
			if tt.wantErr {
				assert.Equal(t, err != nil, true)
				return
			}
			assert.NoError(t, err)
			filePath := filepath.Join(config.CgroupRoot, blkSubName, iocostModelFile)
			modelParamByte, err := ioutil.ReadFile(filePath)
			assert.NoError(t, err)

			modelParams := strings.Split(string(modelParamByte), "\n")
			for _, modelParam := range modelParams {
				paramList := strings.FieldsFunc(modelParam, unicode.IsSpace)
				if len(paramList) >= paramsLen && strings.Compare(paramList[0], devno) == 0 {
					assert.Equal(t, tt.modelParam, modelParam[:len(tt.modelParam)])
					break
				}
			}
		})
	}
}

func TestPartFunction(t *testing.T) {
	const testCgroup = "/var/rubikcgroup/"
	qosParam := strings.Repeat("a", paramMaxLen+1)
	err := writeIOcost(testCgroup, qosParam)
	assert.Equal(t, err.Error(), "param size exceeds "+strconv.Itoa(paramMaxLen))

	_, err = getDirInode(testCgroup)
	assert.Equal(t, true, err != nil)
}

func getAllBlockDevice() (map[string]string, error) {
	files, err := ioutil.ReadDir(sysDevBlock)
	if err != nil {
		log.Infof("read dir %v failed, err:%v", sysDevBlock, err.Error())
		return nil, err
	}
	devs := make(map[string]string)
	for _, f := range files {
		if f.Mode()&os.ModeSymlink != 0 {
			fullName := filepath.Join(sysDevBlock, f.Name())
			realPath, err := os.Readlink(fullName)
			if err != nil {
				continue
			}
			vs := strings.Split(realPath, "/")
			const penultimate = 2
			if len(vs) > penultimate && strings.Compare(vs[len(vs)-penultimate], "block") == 0 {
				const excludeBlock = "dm-"
				dmLen := len(excludeBlock)
				if len(vs[len(vs)-1]) > dmLen && strings.Compare(vs[len(vs)-1][:dmLen], excludeBlock) == 0 {
					continue
				}
				devs[vs[len(vs)-1]] = f.Name()
			}
		}
	}
	return devs, nil
}

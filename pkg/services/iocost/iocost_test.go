package iocost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services/helper"
	"isula.org/rubik/test/try"
)

const (
	sysDevBlock = "/sys/dev/block"
	objName     = "ioCost"
	paramsLen   = 2
)

type ResultItem struct {
	testName   string
	devno      string
	qosCheck   bool
	modelCheck bool
	qosParam   string
	modelParam string
}

var (
	iocostConfigTestItems []IOCostConfig
	resultItmes           []ResultItem
)

func createTestIteam(dev, devno string, enable bool, val int64) (*IOCostConfig, *ResultItem) {
	qosStr := devno + " enable=0"
	name := "Test iocost disable"
	if enable {
		qosStr = devno + " enable=1"
		name = fmt.Sprintf("Test iocost enable: val=%v", val)
	}

	cfg := IOCostConfig{
		Dev:    dev,
		Enable: enable,
		Model:  "linear",
		Param: LinearParam{
			Rbps: val, Rseqiops: val, Rrandiops: val,
			Wbps: val, Wseqiops: val, Wrandiops: val,
		},
	}
	res := ResultItem{
		testName:   name,
		devno:      devno,
		qosCheck:   true,
		modelCheck: enable,
		qosParam:   qosStr,
		modelParam: fmt.Sprintf("%v ctrl=user model=linear rbps=%v rseqiops=%v rrandiops=%v wbps=%v wseqiops=%v wrandiops=%v",
			devno, val, val, val, val, val, val),
	}
	return &cfg, &res
}
func createIOCostConfigTestItems() {
	devs, err := getAllBlockDevice()
	if err != nil {
		panic("get blkck devices error")
	}
	for dev, devno := range devs {
		for _, val := range []int64{600, 700, 800} {
			for _, e := range []bool{true, false, true} {
				cfg, res := createTestIteam(dev, devno, e, val)
				iocostConfigTestItems = append(iocostConfigTestItems, *cfg)
				resultItmes = append(resultItmes, *res)
			}
		}
	}

	var dev, devno string
	for k, v := range devs {
		dev = k
		devno = v
		break
	}

	cfg, res := createTestIteam(dev, devno, true, 900)
	iocostConfigTestItems = append(iocostConfigTestItems, *cfg)
	resultItmes = append(resultItmes, *res)

	/*
		cfg, res = createTestIteam(dev, devno, true, 1000)
		res.testName = "Test iocost config no dev"
		cfg.Dev = "XXX"
	*/
}

func TestIOCostSupport(t *testing.T) {
	assert.Equal(t, ioCostSupport(), true)
	cgroup.InitMountDir("/var/tmp/rubik")
	assert.Equal(t, ioCostSupport(), false)
	cgroup.InitMountDir(constant.DefaultCgroupRoot)
}

func TestIOCostID(t *testing.T) {
	obj := IOCost{ServiceBase: helper.ServiceBase{Name: objName}}
	assert.Equal(t, obj.ID(), objName)
}

func TestIOCostSetConfig(t *testing.T) {
	obj := IOCost{ServiceBase: helper.ServiceBase{Name: objName}}
	err := obj.SetConfig(nil)
	assert.Error(t, err)

	err = obj.SetConfig(func(configName string, d interface{}) error {
		return fmt.Errorf("config handler error test")
	})
	assert.Error(t, err)
	assert.EqualError(t, err, "config handler error test")

	for i, item := range iocostConfigTestItems {
		nodeConfig := NodeConfig{
			NodeName:     "global",
			IOCostConfig: []IOCostConfig{item},
		}

		t.Run(resultItmes[i].testName, func(t *testing.T) {
			var nodeConfigs []NodeConfig
			nodeConfigs = append(nodeConfigs, nodeConfig)
			cfgStr, err := json.Marshal(nodeConfigs)
			assert.NoError(t, err)
			err = obj.SetConfig(func(configName string, d interface{}) error {
				assert.Equal(t, configName, objName)
				return json.Unmarshal(cfgStr, d)
			})
			assert.NoError(t, err)
			checkResult(t, &resultItmes[i])
		})
	}
}

func TestConfigIOCost(t *testing.T) {
	obj := IOCost{ServiceBase: helper.ServiceBase{Name: objName}}
	assert.Equal(t, obj.ID(), objName)

	var devname, devno string
	devs, err := getAllBlockDevice()
	assert.NoError(t, err)
	assert.NotEmpty(t, devs)

	for k, v := range devs {
		devname = k
		devno = v
		break
	}

	testItems := []struct {
		name       string
		config     IOCostConfig
		qosCheck   bool
		modelCheck bool
		qosParam   string
		modelParam string
	}{
		{
			name: "Test iocost enable",
			config: IOCostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: LinearParam{
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
			config: IOCostConfig{
				Dev:    devname,
				Enable: false,
				Model:  "linear",
				Param: LinearParam{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
		{
			name: "Test modifying iocost linear parameters",
			config: IOCostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: LinearParam{
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
			name: "Test iocost disable",
			config: IOCostConfig{
				Dev:    devname,
				Enable: false,
				Model:  "linear",
				Param: LinearParam{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
		{
			name: "Test iocost no dev error",
			config: IOCostConfig{
				Dev:    "xxx",
				Enable: true,
				Model:  "linear",
				Param: LinearParam{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
		{
			name: "Test iocost non-linear error",
			config: IOCostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linearx",
				Param: LinearParam{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 600,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
		{
			name: "Test iocost param error",
			config: IOCostConfig{
				Dev:    devname,
				Enable: true,
				Model:  "linear",
				Param: LinearParam{
					Rbps: 600, Rseqiops: 600, Rrandiops: 600,
					Wbps: 600, Wseqiops: 600, Wrandiops: 0,
				},
			},
			qosCheck:   true,
			modelCheck: false,
			qosParam:   devno + " enable=0",
		},
	}

	for _, tt := range testItems {
		t.Run(tt.name, func(t *testing.T) {
			params := []IOCostConfig{
				tt.config,
			}
			obj.configIOCost(params)
			if tt.qosCheck {
				qos, err := cgroup.ReadCgroupFile(blkcgRootDir, iocostQosFile)
				assert.NoError(t, err)
				qosParams := strings.Split(string(qos), "\n")
				for _, qosParam := range qosParams {
					paramList := strings.FieldsFunc(qosParam, unicode.IsSpace)
					if len(paramList) >= paramsLen && strings.Compare(paramList[0], devno) == 0 {
						assert.Equal(t, tt.qosParam, qosParam[:len(tt.qosParam)])
						break
					}
				}
			}
			if tt.modelCheck {
				modelParamByte, err := cgroup.ReadCgroupFile(blkcgRootDir, iocostModelFile)
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

func TestClearIOcost(t *testing.T) {
	obj := IOCost{ServiceBase: helper.ServiceBase{Name: objName}}
	assert.Equal(t, obj.ID(), objName)

	devs, err := getAllBlockDevice()
	assert.NoError(t, err)
	for _, devno := range devs {
		qosStr := fmt.Sprintf("%v enable=1", devno)
		err := cgroup.WriteCgroupFile(qosStr, blkcgRootDir, iocostQosFile)
		assert.NoError(t, err)
	}

	err = obj.Terminate(nil)
	assert.NoError(t, err)
	qosParamByte, err := cgroup.ReadCgroupFile(blkcgRootDir, iocostQosFile)
	assert.NoError(t, err)
	qosParams := strings.Split(string(qosParamByte), "\n")
	for _, qosParam := range qosParams {
		paramList := strings.FieldsFunc(qosParam, unicode.IsSpace)
		if len(paramList) >= paramsLen {
			assert.Equal(t, paramList[1], "enable=0")
		}
	}
}

func TestSetPodWeight(t *testing.T) {
	// deploy enviroment
	const podCgroupPath = "/rubik-podtest"
	rubikBlkioTestPath := cgroup.AbsoluteCgroupPath(blkcgRootDir, podCgroupPath)
	rubikMemTestPath := cgroup.AbsoluteCgroupPath(memcgRootDir, podCgroupPath)
	try.MkdirAll(rubikBlkioTestPath, constant.DefaultDirMode)
	try.MkdirAll(rubikMemTestPath, constant.DefaultDirMode)
	//defer try.RemoveAll(rubikBlkioTestPath)
	//defer try.RemoveAll(rubikMemTestPath)
	containerPath := podCgroupPath + "/container" + strconv.Itoa(0)
	try.MkdirAll(cgroup.AbsoluteCgroupPath(memcgRootDir, containerPath), constant.DefaultDirMode)
	try.MkdirAll(cgroup.AbsoluteCgroupPath(blkcgRootDir, containerPath), constant.DefaultDirMode)

	tests := []struct {
		name       string
		cgroupPath string
		weight     int
		wantErr    bool
		want       string
		errMsg     string
	}{
		{
			name:       "Test online qos level",
			cgroupPath: containerPath,
			weight:     onlineWeight,
			wantErr:    false,
			want:       "default 1000\n",
		},
		{
			name:       "Test offline qos level",
			cgroupPath: containerPath,
			weight:     offlineWeight,
			wantErr:    false,
			want:       "default 10\n",
		},
		{
			name:       "Test error cgroup path",
			cgroupPath: "/var/log/rubik/rubik-test",
			weight:     offlineWeight,
			wantErr:    true,
			errMsg:     "no such file or diretory",
		},
		{
			name:       "Test error value",
			cgroupPath: containerPath,
			weight:     100000,
			wantErr:    true,
			errMsg:     "invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConfigContainerIOCostWeight(tt.cgroupPath, uint64(tt.weight))
			if tt.wantErr {
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			assert.NoError(t, err)

			// check weight
			weight, err := cgroup.ReadCgroupFile(blkcgRootDir, tt.cgroupPath, iocostWeightFile)
			assert.NoError(t, err)
			assert.Equal(t, string(weight), tt.want)

			// check cgroup writeback
			ino, err := getDirInode(cgroup.AbsoluteCgroupPath(blkcgRootDir, tt.cgroupPath))
			assert.NoError(t, err)
			result := fmt.Sprintf("wb_blkio_path:%v\nwb_blkio_ino:%v\n", tt.cgroupPath, ino)
			blkioInoStr, err := cgroup.ReadCgroupFile(memcgRootDir, tt.cgroupPath, wbBlkioinoFile)
			assert.NoError(t, err)
			assert.Equal(t, result, string(blkioInoStr))
		})
	}
}

func getAllBlockDevice() (map[string]string, error) {
	files, err := os.ReadDir(sysDevBlock)
	if err != nil {
		return nil, err
	}

	devs := make(map[string]string)
	for _, f := range files {
		if f.Type()&os.ModeSymlink != 0 {
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

func checkResult(t *testing.T, result *ResultItem) {
	if result.qosCheck {
		qos, err := cgroup.ReadCgroupFile(blkcgRootDir, iocostQosFile)
		assert.NoError(t, err)
		qosParams := strings.Split(string(qos), "\n")
		for _, qosParam := range qosParams {
			paramList := strings.FieldsFunc(qosParam, unicode.IsSpace)
			if len(paramList) >= paramsLen && strings.Compare(paramList[0], result.devno) == 0 {
				assert.Equal(t, result.qosParam, qosParam[:len(result.qosParam)])
				break
			}
		}
	}
	if result.modelCheck {
		modelParamByte, err := cgroup.ReadCgroupFile(blkcgRootDir, iocostModelFile)
		assert.NoError(t, err)
		modelParams := strings.Split(string(modelParamByte), "\n")
		for _, modelParam := range modelParams {
			paramList := strings.FieldsFunc(modelParam, unicode.IsSpace)
			if len(paramList) >= paramsLen && strings.Compare(paramList[0], result.devno) == 0 {
				assert.Equal(t, result.modelParam, modelParam[:len(result.modelParam)])
				break
			}
		}
	}
}

func TestMain(m *testing.M) {
	if !ioCostSupport() {
		fmt.Println("this machine not support iocost")
		return
	}
	createIOCostConfigTestItems()
	m.Run()
}

// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2022-01-18
// Description: offline pod cache limit directory init function

package cachelimit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/checkpoint"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/perf"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
	"isula.org/rubik/pkg/util"
)

const (
	schemataFile = "schemata"
	numaNodeDir  = "/sys/devices/system/node"
	dirPrefix    = "rubik_"
	perfEvent    = "perf_event"

	lowLevel     = "low"
	middleLevel  = "middle"
	highLevel    = "high"
	maxLevel     = "max"
	dynamicLevel = "dynamic"

	staticMode  = "static"
	dynamicMode = "dynamic"

	defaultL3PercentMax = 100
	defaultMbPercentMax = 100
	minAdjustInterval   = 10
	maxAdjustInterval   = 10000
	// minimum perf duration, unit ms
	minPerfDur = 10
	// maximum perf duration, unit ms
	maxPerfDur = 10000
	minPercent = 10
	maxPercent = 100

	base2, base16, bitSize = 2, 16, 32
)

var (
	numaNum, l3PercentDynamic, mbPercentDynamic int
	defaultLimitMode                            string
	enable                                      bool
	cpm                                         *checkpoint.Manager
)

type cacheLimitSet struct {
	level     string
	clDir     string
	L3Percent int
	MbPercent int
}

// Init init and starts cache limit
func Init(m *checkpoint.Manager, cfg *config.CacheConfig) error {
	enable = true
	if !isHostPidns("/proc/self/ns/pid") {
		return errors.New("share pid namespace with host is needed for cache limit")
	}
	if !perf.HwSupport() {
		return errors.New("hardware event perf not supported")
	}
	if err := checkCacheCfg(cfg); err != nil {
		return err
	}
	if err := checkResctrlExist(cfg); err != nil {
		return err
	}
	if err := initCacheLimitDir(cfg); err != nil {
		return errors.Errorf("cache limit directory create failed: %v", err)
	}
	defaultLimitMode = cfg.DefaultLimitMode
	cpm = m

	go wait.Until(syncCacheLimit, time.Second, config.ShutdownChan)
	missMax, missMin := 20, 10
	dynamicFunc := func() { startDynamic(cfg, missMax, missMin) }
	go wait.Until(dynamicFunc, time.Duration(cfg.AdjustInterval)*time.Millisecond, config.ShutdownChan)
	return nil
}

func isHostPidns(path string) bool {
	ns, err := os.Readlink(path)
	if err != nil {
		log.Errorf("get pid namespace inode error: %v", err)
		return false
	}
	hostPidInode := "4026531836"
	return strings.Trim(ns, "pid:[]") == hostPidInode
}

func checkCacheCfg(cfg *config.CacheConfig) error {
	defaultLimitMode = cfg.DefaultLimitMode
	if defaultLimitMode != staticMode && defaultLimitMode != dynamicMode {
		return errors.Errorf("invalid cache limit mode: %s, should be %s or %s",
			cfg.DefaultLimitMode, staticMode, dynamicMode)
	}
	if cfg.AdjustInterval < minAdjustInterval || cfg.AdjustInterval > maxAdjustInterval {
		return errors.Errorf("adjust interval %d out of range [%d,%d]",
			cfg.AdjustInterval, minAdjustInterval, maxAdjustInterval)
	}
	if cfg.PerfDuration < minPerfDur || cfg.PerfDuration > maxPerfDur {
		return errors.Errorf("perf duration %d out of range [%d,%d]", cfg.PerfDuration, minPerfDur, maxPerfDur)
	}
	for _, per := range []int{cfg.L3Percent.Low, cfg.L3Percent.Mid, cfg.L3Percent.High, cfg.MemBandPercent.Low,
		cfg.MemBandPercent.Mid, cfg.MemBandPercent.High} {
		if per < minPercent || per > maxPercent {
			return errors.Errorf("cache limit percentage %d out of range [%d,%d]", per, minPercent, maxPercent)
		}
	}
	if cfg.L3Percent.Low > cfg.L3Percent.Mid || cfg.L3Percent.Mid > cfg.L3Percent.High {
		return errors.Errorf("cache limit config L3Percent does not satisfy constraint low<=mid<=high")
	}
	if cfg.MemBandPercent.Low > cfg.MemBandPercent.Mid || cfg.MemBandPercent.Mid > cfg.MemBandPercent.High {
		return errors.Errorf("cache limit config MemBandPercent does not satisfy constraint low<=mid<=high")
	}

	return nil
}

// initCacheLimitDir init multi-level cache limit directories
func initCacheLimitDir(cfg *config.CacheConfig) error {
	log.Infof("init cache limit directory")

	var err error
	if numaNum, err = getNUMANum(numaNodeDir); err != nil {
		return errors.Errorf("get NUMA nodes number error: %v", err)
	}

	l3PercentDynamic = cfg.L3Percent.Low
	mbPercentDynamic = cfg.MemBandPercent.Low
	cacheLimitList := []*cacheLimitSet{
		newCacheLimitSet(cfg.DefaultResctrlDir, dynamicLevel, l3PercentDynamic, mbPercentDynamic),
		newCacheLimitSet(cfg.DefaultResctrlDir, lowLevel, cfg.L3Percent.Low, cfg.MemBandPercent.Low),
		newCacheLimitSet(cfg.DefaultResctrlDir, middleLevel, cfg.L3Percent.Mid, cfg.MemBandPercent.Mid),
		newCacheLimitSet(cfg.DefaultResctrlDir, highLevel, cfg.L3Percent.High, cfg.MemBandPercent.High),
		newCacheLimitSet(cfg.DefaultResctrlDir, maxLevel, defaultL3PercentMax, defaultMbPercentMax),
	}

	for _, cl := range cacheLimitList {
		if err = cl.writeResctrlSchemata(numaNum); err != nil {
			return err
		}
	}

	log.Infof("init cache limit directory success")
	return nil
}

func newCacheLimitSet(basePath, level string, l3Per, mbPer int) *cacheLimitSet {
	return &cacheLimitSet{
		level:     level,
		L3Percent: l3Per,
		MbPercent: mbPer,
		clDir:     filepath.Join(filepath.Clean(basePath), dirPrefix+level),
	}
}

// calcLimitedCacheValue calculate number of cache way could be used according to L3 limit percent
func calcLimitedCacheValue(path string, l3Percent int) (string, error) {
	l3BinaryMask, err := getBinaryMask(path)
	if err != nil {
		return "", err
	}
	ten, hundred, binValue := 10, 100, 0
	binLen := l3BinaryMask * l3Percent / hundred
	if binLen == 0 {
		binLen = 1
	}
	for i := 0; i < binLen; i++ {
		binValue = binValue*ten + 1
	}
	decValue, err := strconv.ParseInt(strconv.Itoa(binValue), base2, bitSize)
	if err != nil {
		return "", errors.Errorf("transfer %v to decimal format error: %v", binValue, err)
	}

	return strconv.FormatInt(decValue, base16), nil
}

func (cl *cacheLimitSet) setClDir() error {
	if len(cl.clDir) == 0 {
		return errors.Errorf("cache limit path empty")
	}
	if err := os.Mkdir(cl.clDir, constant.DefaultDirMode); err != nil && !os.IsExist(err) {
		return errors.Errorf("create cache limit directory error: %v", err)
	}
	return nil
}

func (cl *cacheLimitSet) writeResctrlSchemata(numaNum int) error {
	// get cbm mask like "fffff" means 20 cache way
	maskFile := filepath.Join(filepath.Dir(cl.clDir), "info", "L3", "cbm_mask")
	llc, err := calcLimitedCacheValue(maskFile, cl.L3Percent)
	if err != nil {
		return errors.Errorf("get limited cache value from L3 percent error: %v", err)
	}

	if err := cl.setClDir(); err != nil {
		return err
	}
	schemetaFile := filepath.Join(cl.clDir, schemataFile)
	var content string
	for i := 0; i < numaNum; i++ {
		content = content + fmt.Sprintf("L3:%d=%s\n", i, llc) + fmt.Sprintf("MB:%d=%d\n", i, cl.MbPercent)
	}
	if err := ioutil.WriteFile(schemetaFile, []byte(content), constant.DefaultFileMode); err != nil {
		return errors.Errorf("write %s to file %s error: %v", content, schemetaFile, err)
	}

	return nil
}

func (cl *cacheLimitSet) doFlush() error {
	if err := cl.writeResctrlSchemata(numaNum); err != nil {
		return errors.Errorf("adjust dynamic cache limit to l3:%v mb:%v error: %v",
			cl.L3Percent, cl.MbPercent, err)
	}
	l3PercentDynamic = cl.L3Percent
	mbPercentDynamic = cl.MbPercent

	return nil
}

func (cl *cacheLimitSet) flush(cfg *config.CacheConfig, step int) error {
	l3 := nextPercent(l3PercentDynamic, cfg.L3Percent.Low, cfg.L3Percent.High, step)
	mb := nextPercent(mbPercentDynamic, cfg.MemBandPercent.Low, cfg.MemBandPercent.High, step)
	if l3PercentDynamic == l3 && mbPercentDynamic == mb {
		return nil
	}
	log.Infof("flush L3 from %v to %v, Mb from %v to %v", cl.L3Percent, l3, cl.MbPercent, mb)
	cl.L3Percent, cl.MbPercent = l3, mb
	return cl.doFlush()
}

func nextPercent(value, min, max, step int) int {
	value += step
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// startDynamic start monitor online pod qos and adjust dynamic cache limit value
func startDynamic(cfg *config.CacheConfig, missMax, missMin int) {
	if !dynamicExist() {
		return
	}

	ipcMax := 2.1
	ipcMin := 1.6
	stepMore, stepLess := 5, -50
	needMore := true
	limiter := newCacheLimitSet(cfg.DefaultResctrlDir, dynamicLevel, l3PercentDynamic, mbPercentDynamic)

	onlinePods := cpm.ListOnlinePods()
	for _, p := range onlinePods {
		ipc, cacheMiss, LLCMiss, err := getPodPerf(p, cfg.PerfDuration)
		if err != nil {
			log.Errorf(err.Error())
		}

		if estimateQosViolation(p, missMax, cacheMiss, LLCMiss, ipcMin, ipc) {
			if err := limiter.flush(cfg, stepLess); err != nil {
				log.Errorf(err.Error())
			}
			return
		}
		if (cacheMiss >= missMin || LLCMiss >= missMin) && ipc >= ipcMax {
			needMore = false
		}
	}

	if !needMore {
		return
	}
	if err := limiter.flush(cfg, stepMore); err != nil {
		log.Errorf(err.Error())
	}
}

func estimateQosViolation(p *typedef.PodInfo, missMax, cacheMiss, LLCMiss int, ipcMin, ipc float64) bool {
	if ipc < ipcMin {
		log.Infof("online pod %v ipc down: %v lower offline cache limit",
			p.UID, ipc)
		return true
	}
	if cacheMiss >= missMax || LLCMiss >= missMax {
		log.Infof("online pod %v cache miss: %v LLC miss: %v exceeds maxmiss, lower offline cache limit",
			p.UID, cacheMiss, LLCMiss)
		return true
	}
	return false
}

func dynamicExist() bool {
	offlinePods := cpm.ListOfflinePods()
	for _, p := range offlinePods {
		err := SyncLevel(p)
		if err != nil {
			continue
		}
		if p.CacheLimitLevel == dynamicLevel {
			return true
		}
	}
	return false
}

// getPodPerf return ipc, cache miss, llc miss of the pod
func getPodPerf(pi *typedef.PodInfo, perfDu int) (float64, int, int, error) {
	cgroupPath := filepath.Join(config.CgroupRoot, perfEvent, pi.CgroupPath)
	if !util.PathExist(cgroupPath) {
		return 0, 0, 0.0, errors.Errorf("path %v not exist, cannot get perf statistics", cgroupPath)
	}

	stat, err := perf.CgroupStat(cgroupPath, time.Duration(perfDu)*time.Millisecond)
	if err != nil {
		return 0, 0, 0.0, err
	}

	return float64(stat.Instructions) / (1.0 + float64(stat.CpuCycles)),
		int(100.0 * float64(stat.CacheMisses) / (1.0 + float64(stat.CacheReferences))),
		int(100.0 * float64(stat.LLCMiss) / (1.0 + float64(stat.LLCAccess))),
		nil
}

func getPodCacheMiss(pi *typedef.PodInfo, perfDu int) (int, int) {
	cgroupPath := filepath.Join(config.CgroupRoot, perfEvent, pi.CgroupPath)
	if !util.PathExist(cgroupPath) {
		return 0, 0
	}

	stat, err := perf.CgroupStat(cgroupPath, time.Duration(perfDu)*time.Millisecond)
	if err != nil {
		return 0, 0
	}

	return int(100.0 * float64(stat.CacheMisses) / (1.0 + float64(stat.CacheReferences))),
		int(100.0 * float64(stat.LLCMiss) / (1.0 + float64(stat.LLCAccess)))
}

// ClEnabled return if cache limit is enabled
func ClEnabled() bool {
	return enable
}

// checkResctrlExist check if resctrl directory exists
func checkResctrlExist(cfg *config.CacheConfig) error {
	if !util.PathExist(cfg.DefaultResctrlDir) {
		return errors.Errorf("path %v not exist, not support cache limit", cfg.DefaultResctrlDir)
	}
	schemataPath := filepath.Join(cfg.DefaultResctrlDir, schemataFile)
	if !util.PathExist(schemataPath) {
		return errors.Errorf("path %v not exist, check if %v directory is mounted",
			schemataPath, cfg.DefaultResctrlDir)
	}
	return nil
}

func getNUMANum(path string) (int, error) {
	files, err := filepath.Glob(filepath.Join(path, "node*"))
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// getBinaryMask get l3 limit mask like "7ff" and transfer it to binary like "111 1111 1111", return binary length 11
func getBinaryMask(path string) (int, error) {
	maskValue, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return -1, errors.Errorf("get L3 mask value error: %v", err)
	}

	// transfer mask to binary format
	decMask, err := strconv.ParseInt(strings.TrimSpace(string(maskValue)), base16, bitSize)
	if err != nil {
		return -1, errors.Errorf("transfer L3 mask value %v to decimal format error: %v", string(maskValue), err)
	}
	return len(strconv.FormatInt(decMask, base2)), nil
}

// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Xiang Li
// Create: 2023-02-21
// Description: This file will init cache limit directories before services running

// Package dyncache is the service used for cache limit setting
package dyncache

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/perf"
	"isula.org/rubik/pkg/common/util"
)

const (
	base2, base16, bitSize = 2, 16, 32
)

type limitSet struct {
	dir       string
	level     string
	l3Percent int
	mbPercent int
}

// InitCacheLimitDir init multi-level cache limit directories
func (c *DynCache) InitCacheLimitDir() error {
	log.Debugf("init cache limit directory")
	const (
		defaultL3PercentMax = 100
		defaultMbPercentMax = 100
	)
	if !perf.Support() {
		return fmt.Errorf("current os does not support perf hw pmu events")
	}
	if err := checkHostPidns(c.config.DefaultPidNameSpace); err != nil {
		return err
	}
	if err := checkResctrlPath(c.config.DefaultResctrlDir); err != nil {
		return err
	}
	numaNum, err := getNUMANum(c.Attr.NumaNodeDir)
	if err != nil {
		return fmt.Errorf("failed to get NUMA nodes number: %v", err)
	}
	c.Attr.NumaNum = numaNum
	c.Attr.L3PercentDynamic = c.config.L3Percent.Low
	c.Attr.MemBandPercentDynamic = c.config.MemBandPercent.Low

	cacheLimitList := []*limitSet{
		c.newCacheLimitSet(levelDynamic, c.Attr.L3PercentDynamic, c.Attr.MemBandPercentDynamic),
		c.newCacheLimitSet(levelLow, c.config.L3Percent.Low, c.config.MemBandPercent.Low),
		c.newCacheLimitSet(levelMiddle, c.config.L3Percent.Mid, c.config.MemBandPercent.Mid),
		c.newCacheLimitSet(levelHigh, c.config.L3Percent.High, c.config.MemBandPercent.High),
		c.newCacheLimitSet(levelMax, defaultL3PercentMax, defaultMbPercentMax),
	}

	for _, cl := range cacheLimitList {
		if err := cl.writeResctrlSchemata(c.Attr.NumaNum); err != nil {
			return err
		}
	}

	log.Debugf("initialize cache limit directory successfully")
	return nil
}

func (c *DynCache) newCacheLimitSet(level string, l3Per, mbPer int) *limitSet {
	return &limitSet{
		level:     level,
		l3Percent: l3Per,
		mbPercent: mbPer,
		dir:       filepath.Join(filepath.Clean(c.config.DefaultResctrlDir), resctrlDirPrefix+level),
	}
}

func (cl *limitSet) setDir() error {
	if err := os.Mkdir(cl.dir, constant.DefaultDirMode); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create cache limit directory: %v", err)
	}
	return nil
}

func (cl *limitSet) writeResctrlSchemata(numaNum int) error {
	// get cbm mask like "fffff" means 20 cache way
	maskFile := filepath.Join(filepath.Dir(cl.dir), "info", "L3", "cbm_mask")
	llc, err := calcLimitedCacheValue(maskFile, cl.l3Percent)
	if err != nil {
		return fmt.Errorf("failed to get limited cache value from L3 percent: %v", err)
	}

	if err := cl.setDir(); err != nil {
		return err
	}
	schemetaFile := filepath.Join(cl.dir, schemataFileName)
	var content string
	var l3List, mbList []string
	for i := 0; i < numaNum; i++ {
		l3List = append(l3List, fmt.Sprintf("%d=%s", i, llc))
		mbList = append(mbList, fmt.Sprintf("%d=%d", i, cl.mbPercent))
	}
	l3 := fmt.Sprintf("L3:%s\n", strings.Join(l3List, ";"))
	mb := fmt.Sprintf("MB:%s\n", strings.Join(mbList, ";"))
	content = l3 + mb
	if err := util.WriteFile(schemetaFile, content); err != nil {
		return fmt.Errorf("failed to write %s to file %s: %v", content, schemetaFile, err)
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
	maskValue, err := util.ReadFile(path)
	if err != nil {
		return -1, fmt.Errorf("failed to get L3 mask value: %v", err)
	}

	// transfer mask to binary format
	decMask, err := strconv.ParseInt(strings.TrimSpace(string(maskValue)), base16, bitSize)
	if err != nil {
		return -1, fmt.Errorf("failed to transfer L3 mask value %v to decimal format: %v", string(maskValue), err)
	}
	return len(strconv.FormatInt(decMask, base2)), nil
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
		return "", fmt.Errorf("failed to transfer %v to decimal format: %v", binValue, err)
	}

	return strconv.FormatInt(decValue, base16), nil
}

func checkHostPidns(path string) error {
	ns, err := os.Readlink(path)
	if err != nil {
		return fmt.Errorf("failed to get pid namespace inode: %v", err)
	}
	hostPidInode := "4026531836"
	if strings.Trim(ns, "pid:[]") != hostPidInode {
		return fmt.Errorf("not share pid ns with host")
	}
	return nil
}

func checkResctrlPath(path string) error {
	if !util.PathExist(path) {
		return fmt.Errorf("resctrl path %v does not exist, does not support cache limit", path)
	}
	schemataPath := filepath.Join(path, schemataFileName)
	if !util.PathExist(schemataPath) {
		return fmt.Errorf("resctrl schemata file %v does not exist, please check if %v directory is mounted",
			schemataPath, path)
	}
	return nil
}

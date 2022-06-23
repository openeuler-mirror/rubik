// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Yang Feiyu
// Create: 2022-6-7
// Description: memory setting for pods

package memory

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/typedef"
)

type memoryInfo struct {
	total     int64
	free      int64
	available int64
}

type dynLevel struct {
	m       *MemoryManager
	memInfo memoryInfo
	st      status
}

func newDynLevel(m *MemoryManager) (f *dynLevel) {
	return &dynLevel{
		st: newStatus(),
		m:  m,
	}
}

func (f *dynLevel) Run() {
	go wait.Until(f.run, time.Duration(f.m.checkInterval)*time.Second, f.m.stop)
}

func (f *dynLevel) run() {
	f.updateStatus()
	log.Logf("memory manager updates status with memory free: %v, memory total: %v", f.memInfo.free, f.memInfo.total)
	f.reclaim()
	log.Logf("memory manager reclaims done and pressure level is %s", &f.st)
}

func (f *dynLevel) updateStatus() {
	memInfo, err := getMemoryInfo()
	if err != nil {
		log.Errorf("getMemoryInfo failed with error: %v, it should not happen", err)
		return
	}
	f.memInfo = memInfo
	f.st.transitionStatus(float64(memInfo.free) / float64(memInfo.total))
}

func (f *dynLevel) limitOfflineContainers(ft fileType) {
	containers := f.m.cpm.ListOfflineContainers()
	for _, c := range containers {
		if err := f.limitContainer(c, ft); err != nil {
			log.Errorf("limit memory for container: %v failed, filetype: %v, err: %v", c.ID, ft, err)
		}
	}
}

func (f *dynLevel) limitContainer(c *typedef.ContainerInfo, ft fileType) error {
	path := c.CgroupPath("memory")
	limit, err := readMemoryFile(filepath.Join(path, memoryUsageFile))
	if err != nil {
		return err
	}

	for i := 0; i < maxRetry; i++ {
		limit += int64(float64(f.memInfo.free) * extraFreePercentage)
		if err = writeMemoryLimit(path, typedef.FormatInt64(limit), ft); err == nil {
			break
		}
		log.Errorf("failed to write memory limit from path: %v, will retry now, retry num: %v", path, i)
	}

	return err
}

// dropCaches will echo 3 > /proc/sys/vm/drop_caches
func (f *dynLevel) dropCaches() {
	var err error
	for i := 0; i < maxRetry; i++ {
		if err = ioutil.WriteFile(dropCachesFilePath, []byte("3"), constant.DefaultFileMode); err == nil {
			log.Logf("drop caches success")
			return
		}
		log.Errorf("drop caches failed, error: %v, will retry later, retry num: %v", err, i)
	}
}

func (f *dynLevel) forceEmptyOfflineContainers() {
	containers := f.m.cpm.ListOfflineContainers()
	for _, c := range containers {
		if err := writeForceEmpty(c.CgroupPath("memory")); err != nil {
			log.Errorf("force empty for container: %v failed, err: %v", c.ID, err)
		}

	}
}

func (f *dynLevel) reclaimInPressure() {
	switch f.st.pressureLevel {
	case low:
		// do soft limit
		f.limitOfflineContainers(msoftLimit)
	case mid:
		f.forceEmptyOfflineContainers()
	case high:
		// do hard limit
		f.limitOfflineContainers(mlimit)
	case critical:
		// drop caches and do hard limit
		f.dropCaches()
		f.limitOfflineContainers(mlimit)
	}
}

func (f *dynLevel) reclaimInRelieve() {
	f.st.relieveCnt++
	containers := f.m.cpm.ListOfflineContainers()
	for _, c := range containers {
		recoverContainerMemoryLimit(c, f.st.relieveCnt == relieveMaxCnt)
	}
}

func (f *dynLevel) reclaim() {
	if f.st.isNormal() {
		return
	}

	if f.st.isRelieve() {
		f.reclaimInRelieve()
		return
	}

	f.reclaimInPressure()
}

func writeForceEmpty(cgroupPath string) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		if err = writeMemoryFile(cgroupPath, memoryForceEmptyFile, "0"); err == nil {
			log.Logf("force cgroup memory %v empty success", cgroupPath)
			return nil
		}
		log.Errorf("force clean memory failed for %s: %v, will retry later, retry num: %v", cgroupPath, err, i)
	}

	return err
}

func recoverContainerMemoryLimit(c *typedef.ContainerInfo, reachMax bool) {
	// ratio 0.1 means, newLimit = oldLimit * 1.1
	const ratio = 0.1
	var memLimit int64
	path := c.CgroupPath("memory")
	if reachMax {
		memLimit = maxSysMemLimit
		if err := writeMemoryLimit(path, typedef.FormatInt64(memLimit), mlimit); err != nil {
			log.Errorf("failed to write memory limit from path:%v container:%v", path, c.ID)
		}

		if err := writeMemoryLimit(path, typedef.FormatInt64(memLimit), msoftLimit); err != nil {
			log.Errorf("failed to write memory soft limit from path:%v container:%v", path, c.ID)
		}
		return
	}

	memLimit, err := readMemoryFile(filepath.Join(path, memoryLimitFile))
	if err != nil {
		log.Errorf("failed to read from path:%v container:%v", path, c.ID)
		return
	}

	memLimit = int64(float64(memLimit) * (1 + ratio))
	if memLimit < 0 {
		// it means the limit value has reached max, just return
		return
	}

	if err := writeMemoryLimit(path, typedef.FormatInt64(memLimit), mlimit); err != nil {
		log.Errorf("failed to write memory limit from path:%v container:%v", path, c.ID)
	}
}

package dynmemory

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/log"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	fssrInterval          = 5
	fssrIntervalCount     = 12
	scale                 = 10
	highRatio             = 80
	memcgRootDir          = "memory"
	highMemFile           = "memory.high"
	highMemAsyncRatioFile = "memory.high_async_ratio"
	memInfoFile           = "/proc/meminfo"
)

// fssr structure for policy
type fssrDynMemAdapter struct {
	memTotal    int64
	memHigh     int64
	reservedMem int64
	count       int64
	viewer      api.Viewer
}

// initFssrDynMemAdapter initializes a new fssrDynMemAdapter struct.
func initFssrDynMemAdapter() *fssrDynMemAdapter {
	if total, err := getFieldMemory("MemTotal"); err == nil && total > 0 {
		return &fssrDynMemAdapter{
			memTotal:    total,
			memHigh:     total * 8 / 10,
			reservedMem: total * 1 / 10,
			count:       0,
		}
	}
	return nil
}

// preStart initializes the fssrDynMemAdapter with the provided viewer and
// deals with any existing pods.
func (f *fssrDynMemAdapter) preStart(viewer api.Viewer) error {
	f.viewer = viewer
	return f.dealExistedPods()
}

// getInterval returns the fssrInterval value.
func (f *fssrDynMemAdapter) getInterval() int {
	return fssrInterval
}

// dynamicAdjust adjusts the memory allocation of the fssrDynMemAdapter by
// increasing or decreasing the amount of memory reserved for offline pods
// based on the current amount of free memory available on the system.
func (f *fssrDynMemAdapter) dynamicAdjust() {
	var freeMem int64
	var err error
	if freeMem, err = getFieldMemory("MemFree"); err != nil {
		return
	}

	var memHigh int64 = 0
	if freeMem > 2*f.reservedMem {
		if f.count < fssrIntervalCount {
			f.count++
			return
		}
		// no risk of overflow
		memHigh = f.memHigh + f.memTotal/100
		if memHigh > f.memTotal*8/10 {
			memHigh = f.memTotal * 8 / 10
		}
	} else if freeMem < f.reservedMem {
		memHigh = f.memHigh - f.memTotal/10
		if memHigh < 0 {
			return
		}
		if memHigh < f.memTotal*3/10 {
			memHigh = f.memTotal * 3 / 10
		}
	}
	if memHigh != f.memHigh {
		f.memHigh = memHigh
		f.adjustOfflinePodHighMemory()
	}

	f.count = 0
}

// adjustOfflinePodHighMemory adjusts the memory.high of offline pods.
func (f *fssrDynMemAdapter) adjustOfflinePodHighMemory() error {
	pods := listOfflinePods(f.viewer)
	for _, podInfo := range pods {
		if err := setOfflinePodHighMemory(podInfo.Path, f.memHigh); err != nil {
			return err
		}
	}
	return nil
}

// dealExistedPods handles offline pods by setting their memory.high and memory.high_async_ratio
func (f *fssrDynMemAdapter) dealExistedPods() error {
	pods := listOfflinePods(f.viewer)
	for _, podInfo := range pods {
		if err := f.setOfflinePod(podInfo.Path); err != nil {
			log.Errorf("set fssr of offline pod[%v] error:%v", podInfo.UID, err)
		}
	}
	return nil
}

// listOfflinePods returns a map of offline PodInfo objects.
func listOfflinePods(viewer api.Viewer) map[string]*typedef.PodInfo {
	offlineValue := "true"
	return viewer.ListPodsWithOptions(func(pi *typedef.PodInfo) bool {
		return pi.Annotations[constant.PriorityAnnotationKey] == offlineValue
	})
}

// setOfflinePod sets the offline pod for the given path.
func (f *fssrDynMemAdapter) setOfflinePod(path string) error {
	if err := setOfflinePodHighAsyncRatio(path, highRatio); err != nil {
		return err
	}
	return setOfflinePodHighMemory(path, f.memHigh)
}

// setOfflinePodHighMemory sets the high memory limit for the specified pod in the
// cgroup memory
func setOfflinePodHighMemory(podPath string, memHigh int64) error {
	if err := cgroup.WriteCgroupFile(strconv.FormatUint(uint64(memHigh), scale), memcgRootDir,
		podPath, highMemFile); err != nil {
		return err
	}
	return nil
}

// setOfflinePodHighAsyncRatio sets the high memory async ratio for a pod in an offline state.
func setOfflinePodHighAsyncRatio(podPath string, ratio uint) error {
	if err := cgroup.WriteCgroupFile(strconv.FormatUint(uint64(ratio), scale), memcgRootDir,
		podPath, highMemAsyncRatioFile); err != nil {
		return err
	}
	return nil
}

// getFieldMemory retrieves the amount of memory used by a certain field in the
// memory information file.
func getFieldMemory(field string) (int64, error) {
	if !util.PathExist(memInfoFile) {
		return 0, fmt.Errorf("%v: no such file or diretory", memInfoFile)
	}
	f, err := os.Open(memInfoFile)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var total int64
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		if bytes.HasPrefix(scan.Bytes(), []byte(field+":")) {
			if _, err := fmt.Sscanf(scan.Text(), field+":%d", &total); err != nil {
				return 0, err
			}
			return total * 1024, nil
		}
	}
	return 0, fmt.Errorf("%v file not contain '%v' field", memInfoFile, field)
}

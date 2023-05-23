package dynmemory

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
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
	api         api.Viewer
	count       int64
}

// initFssrDynMemAdapter function
func initFssrDynMemAdapter() *fssrDynMemAdapter {
	if total, err := getFieldMemory("MemTotal"); err == nil && total > 0 {
		return &fssrDynMemAdapter{
			memTotal:    total,
			memHigh:     total * 8 / 10,
			reservedMem: total * 8 / 10,
		}
	}
	return nil
}

// preStart function
func (f *fssrDynMemAdapter) preStart(api api.Viewer) error {
	f.api = api
	return f.dealExistedPods()
}

// getInterval function
func (f *fssrDynMemAdapter) getInterval() int {
	return fssrInterval
}

// dynadjust function
func (f *fssrDynMemAdapter) dynamicAdjust() {
	var freeMem int64
	var err error
	if freeMem, err = getFieldMemory("MemFree"); err != nil {
		return
	}
	if freeMem > 2*f.reservedMem {
		if f.count < fssrIntervalCount {
			f.count++
			return
		}
		memHigh := f.memHigh + f.memTotal/100
		if memHigh > f.memTotal*8/10 {
			memHigh = f.memTotal * 8 / 10
		}
		if memHigh != f.memHigh {
			f.memHigh = memHigh
			f.adjustOfflinePodHighMemory()
		}
	} else if freeMem < f.reservedMem {
		memHigh := f.memHigh - f.memTotal/10
		if memHigh < 0 {
			return
		}
		if memHigh < f.memTotal*3/10 {
			memHigh = f.memTotal * 3 / 10
		}
		if memHigh != f.memHigh {
			f.memHigh = memHigh
			f.adjustOfflinePodHighMemory()
		}
	}
	f.count = 0
}

func (f *fssrDynMemAdapter) adjustOfflinePodHighMemory() error {
	pods := listOfflinePods(f.api)
	for _, podInfo := range pods {
		if err := setOfflinePodHighMemory(podInfo.Path, f.memHigh); err != nil {
			return err
		}
	}
	return nil
}

// dealExistedPods function
func (f *fssrDynMemAdapter) dealExistedPods() error {
	pods := listOfflinePods(f.api)
	for _, podInfo := range pods {
		if err := setOfflinePodHighMemory(podInfo.Path, f.memHigh); err != nil {
			return err
		}
		if err := setOfflinePodHighAsyncRatio(podInfo.Path, highRatio); err != nil {
			return err
		}
	}
	return nil
}

func listOfflinePods(viewer api.Viewer) map[string]*typedef.PodInfo {
	offlineValue := "true"
	return viewer.ListPodsWithOptions(func(pi *typedef.PodInfo) bool {
		return pi.Annotations[constant.PriorityAnnotationKey] == offlineValue
	})
}

func setOfflinePodHighMemory(podPath string, high int64) error {
	if err := cgroup.WriteCgroupFile(strconv.FormatUint(uint64(high), scale), memcgRootDir,
		podPath, highMemFile); err != nil {
		return err
	}
	return nil
}

func setOfflinePodHighAsyncRatio(podPath string, ratio uint64) error {
	if err := cgroup.WriteCgroupFile(strconv.FormatUint(ratio, scale), memcgRootDir,
		podPath, highMemAsyncRatioFile); err != nil {
		return err
	}
	return nil
}

// getFieldMemory function
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

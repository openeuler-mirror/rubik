package iocost

import (
	"fmt"
	"strconv"

	"isula.org/rubik/pkg/core/typedef/cgroup"
)

const (
	// iocost model file
	iocostModelFile  = "blkio.cost.model"
	// iocost weight file
	iocostWeightFile = "blkio.cost.weight"
	// iocost weight qos file
	iocostQosFile    = "blkio.cost.qos"
	// cgroup writeback file 
	wbBlkioinoFile   = "memory.wb_blkio_ino"
)

// ConfigIOCostQoS for config iocost qos.
func ConfigIOCostQoS(devno string, enable bool) error {
	t := 0
	if enable {
		t = 1
	}
	qosParam := fmt.Sprintf("%v enable=%v ctrl=user min=100.00 max=100.00", devno, t)
	return cgroup.WriteCgroupFile(qosParam, blkcgRootDir, iocostQosFile)
}

// ConfigIOCostModel for config iocost model
func ConfigIOCostModel(devno string, p interface{}) error {
	var paramStr string
	switch param := p.(type) {
	case LinearParam:
		if param.Rbps <= 0 || param.Rseqiops <= 0 || param.Rrandiops <= 0 ||
			param.Wbps <= 0 || param.Wseqiops <= 0 || param.Wrandiops <= 0 {
			return fmt.Errorf("invalid params, linear params must be greater than 0")
		}

		paramStr = fmt.Sprintf("%v rbps=%v rseqiops=%v rrandiops=%v wbps=%v wseqiops=%v wrandiops=%v",
			devno,
			param.Rbps, param.Rseqiops, param.Rrandiops,
			param.Wbps, param.Wseqiops, param.Wrandiops,
		)
	default:
		return fmt.Errorf("model param is errror")
	}
	return cgroup.WriteCgroupFile(paramStr, blkcgRootDir, iocostModelFile)
}

// ConfigContainerIOCostWeight for config iocost weight
// cgroup v1 iocost cannot be inherited. Therefore, only the container level can be configured.
func ConfigContainerIOCostWeight(containerRelativePath string, weight uint64) error {
	if err := cgroup.WriteCgroupFile(strconv.FormatUint(weight, scale), blkcgRootDir,
		containerRelativePath, iocostWeightFile); err != nil {
		return err
	}
	if err := bindMemcgBlkcg(containerRelativePath); err != nil {
		return err
	}
	return nil
}

// bindMemcgBlkcg for bind memcg and blkcg
func bindMemcgBlkcg(containerRelativePath string) error {
	blkcgPath := cgroup.AbsoluteCgroupPath(blkcgRootDir, containerRelativePath)
	ino, err := getDirInode(blkcgPath)
	if err != nil {
		return err
	}

	if err := cgroup.WriteCgroupFile(strconv.FormatUint(ino, scale),
		memcgRootDir, containerRelativePath, wbBlkioinoFile); err != nil {
		return err
	}
	return nil
}

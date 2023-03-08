package blkcg

import (
	"os"
	"strings"
	"unicode"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/services"
)

var log api.Logger

const (
	blkcgRootDir  = "blkio"
	memcgRootDir  = "memory"
	offlineWeight = 10
	onlineWeight  = 1000
	scale         = 10
)

// LinearParam for linear model
type LinearParam struct {
	Rbps      int64 `json:"rbps,omitempty"`
	Rseqiops  int64 `json:"rseqiops,omitempty"`
	Rrandiops int64 `json:"rrandiops,omitempty"`
	Wbps      int64 `json:"wbps,omitempty"`
	Wseqiops  int64 `json:"wseqiops,omitempty"`
	Wrandiops int64 `json:"wrandiops,omitempty"`
}

// IOCostConfig define iocost for node
type IOCostConfig struct {
	Dev    string      `json:"dev,omitempty"`
	Enable bool        `json:"enable,omitempty"`
	Model  string      `json:"model,omitempty"`
	Param  LinearParam `json:"param,omitempty"`
}

// NodeConfig define the config of node, include iocost
type NodeConfig struct {
	NodeName     string         `json:"nodeName,omitempty"`
	IOCostEnable bool           `json:"iocostEnable,omitempty"`
	IOCostConfig []IOCostConfig `json:"iocostConfig,omitempty"`
}

// IOCost for iocost class
type IOCost struct {
	name       string
	nodeConfig []NodeConfig
}

var (
	nodeName string
)

// init iocost: register service and ensure this platform support iocost.
func init() {
	services.Register("iocost", func() interface{} {
		return &IOCost{name: "iocost"}
	})
}

// IOCostSupport tell if the os support iocost.
func IOCostSupport() bool {
	qosFile := cgroup.AbsoluteCgroupPath(blkcgRootDir, iocostQosFile)
	modelFile := cgroup.AbsoluteCgroupPath(blkcgRootDir, iocostModelFile)
	return util.PathExist(qosFile) && util.PathExist(modelFile)
}

func SetLogger(l api.Logger) {
	log = l
}

// ID for get the name of iocost
func (b *IOCost) ID() string {
	return b.name
}

func (b *IOCost) PreStart(viewer api.Viewer) error {
	nodeName = os.Getenv(constant.NodeNameEnvKey)
	if err := b.loadConfig(); err != nil {
		return err
	}
	return b.dealExistedPods(viewer)
}

func (b *IOCost) loadConfig() error {
	var nodeConfig *NodeConfig
	// global will set all node
	for _, config := range b.nodeConfig {
		if config.NodeName == nodeName {
			nodeConfig = &config
			break
		}
		if config.NodeName == "global" {
			nodeConfig = &config
		}
	}

	// no config, return
	if nodeConfig == nil {
		log.Warnf("no matching node exist:%v", nodeName)
		return nil
	}

	// ensure that previous configuration is cleared.
	if err := b.clearIOCost(); err != nil {
		log.Errorf("clear iocost err:%v", err)
		return err
	}

	if !nodeConfig.IOCostEnable {
		// clear iocost before
		return nil
	}

	b.configIOCost(nodeConfig.IOCostConfig)
	return nil
}

func (b *IOCost) Terminate(viewer api.Viewer) error {
	if err := b.clearIOCost(); err != nil {
		return err
	}
	return nil
}

func (b *IOCost) dealExistedPods(viewer api.Viewer) error {
	pods := viewer.ListPodsWithOptions()
	for _, pod := range pods {
		b.configPodIOCostWeight(pod)
	}
	return nil
}

func (b *IOCost) AddFunc(podInfo *typedef.PodInfo) error {
	return b.configPodIOCostWeight(podInfo)
}

func (b *IOCost) UpdateFunc(old, new *typedef.PodInfo) error {
	return b.configPodIOCostWeight(new)
}

// deal with  deleted pod.
func (b *IOCost) DeleteFunc(podInfo *typedef.PodInfo) error {
	return nil
}

func (b *IOCost) configIOCost(configs []IOCostConfig) {
	for _, config := range configs {
		devno, err := getBlkDeviceNo(config.Dev)
		if err != nil {
			log.Errorf("this device not found:%v", config.Dev)
			continue
		}
		if config.Model == "linear" {
			if err := ConfigIOCostModel(devno, config.Param); err != nil {
				log.Errorf("this device not found:%v", err)
				continue
			}
		} else {
			log.Errorf("non-linear models are not supported")
			continue
		}
		if err := ConfigIOCostQoS(devno, config.Enable); err != nil {
			log.Errorf("Config iocost qos failed:%v", err)
		}
	}
}

// clearIOCost used to disable all iocost
func (b *IOCost) clearIOCost() error {
	qosbytes, err := cgroup.ReadCgroupFile(blkcgRootDir, iocostQosFile)
	if err != nil {
		return err
	}

	if len(qosbytes) == 0 {
		return nil
	}

	qosParams := strings.Split(string(qosbytes), "\n")
	for _, qosParam := range qosParams {
		words := strings.FieldsFunc(qosParam, unicode.IsSpace)
		if len(words) != 0 {
			if err := ConfigIOCostQoS(words[0], false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *IOCost) configPodIOCostWeight(podInfo *typedef.PodInfo) error {
	var weight uint64 = offlineWeight
	if podInfo.Annotations[constant.PriorityAnnotationKey] == "true" {
		weight = onlineWeight
	}
	for _, container := range podInfo.IDContainersMap {
		if err := ConfigContainerIOCostWeight(container.CgroupPath, weight); err != nil {
			return err
		}
	}
	return nil
}

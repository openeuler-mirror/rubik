// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: weiyuan
// Create: 2024-05-28
// Description:  This file defines NRIRawPod and NRIRawContainerwhich encapsulate nri pod and container info

// Package typedef defines core struct and methods for rubik
package typedef

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/nri/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type (
	// NRIRawContainer is nri container structure
	NRIRawContainer api.Container
	// NRIRawPod is nri pod structure
	NRIRawPod api.PodSandbox
)

const (
	// nri Container Annotations "kubectl.kubernetes.io/last-applied-configuration" means container applied config
	containerAppliedConfiguration string = "kubectl.kubernetes.io/last-applied-configuration"
	// /proc/self/cgroup file
	procSelfCgroupFile = "/proc/self/cgroup"

	containerSpec string      = "containers"
	resourcesSpec string      = "resources"
	nanoToMicro   float64     = 1000
	fileMode      os.FileMode = 0666
)

// convert NRIRawPod structure to PodInfo structure
func (pod *NRIRawPod) ConvertNRIRawPod2PodInfo() *PodInfo {
	if pod == nil {
		return nil
	}
	return &PodInfo{
		Hierarchy: cgroup.Hierarchy{
			Path: pod.CgroupPath(),
		},
		Name:            pod.Name,
		UID:             pod.Uid,
		Namespace:       pod.Namespace,
		IDContainersMap: make(map[string]*ContainerInfo, 0),
		Annotations:     pod.Annotations,
		Labels:          pod.Labels,
		ID:              pod.Id,
	}
}

// get pod Qos
func (pod *NRIRawPod) GetQosClass() string {
	var podQosClass string
	podQosClass = "Guaranteed"
	if strings.Contains(pod.Linux.CgroupsPath, "burstable") {
		podQosClass = "Burstable"
	}
	if strings.Contains(pod.Linux.CgroupsPath, "besteffort") {
		podQosClass = "BestEffort"
	}
	return podQosClass
}

// get pod cgroupPath
func (pod *NRIRawPod) CgroupPath() string {
	id := pod.Uid

	qosClassPath := ""
	switch corev1.PodQOSClass(pod.GetQosClass()) {
	case corev1.PodQOSGuaranteed:
	case corev1.PodQOSBurstable:
		qosClassPath = strings.ToLower(string(corev1.PodQOSBurstable))
	case corev1.PodQOSBestEffort:
		qosClassPath = strings.ToLower(string(corev1.PodQOSBestEffort))
	default:
		return ""
	}
	return cgroup.ConcatPodCgroupPath(qosClassPath, id)
}

// get pod running state
func (pod *NRIRawPod) Running() bool {
	return true
}

// get pod UID
func (pod *NRIRawPod) ID() string {
	if pod == nil {
		return ""
	}
	return string(pod.Uid)
}

// convert NRIRawContainer structure to ContainerInfo structure
func (container *NRIRawContainer) ConvertNRIRawContainer2ContainerInfo() *ContainerInfo {
	if container == nil {
		return nil
	}
	requests, limits := container.GetResourceMaps()
	return &ContainerInfo{
		Hierarchy: cgroup.Hierarchy{
			Path: container.CgroupPath(),
		},
		Name:             container.Name,
		ID:               container.Id,
		RequestResources: requests,
		LimitResources:   limits,
		PodSandboxId:     container.PodSandboxId,
	}
}

// get container cgroupPath
func (container *NRIRawContainer) CgroupPath() string {
	var path string
	/*
		When using systemd cgroup driver with burstable qos:
		kubepods-burstable-podbb29f378_0c50_4da4_b070_be919e350db2.slice:crio:a575c8505d48fd0f75488fcf979dea7a693633e99709bb82308af54f3bafb186
		convert to:
		"kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod84d0ae01_83a0_42a7_8990_abbb16e59923.slice/crio-66ff3a44a254533880e6b50e8fb52e0311c9158eb66426ae244066a4f11b26e5.scope"

		When using cgroupfs as cgroup driver and isula, docker, containerd as runtime wich burstable qos:
		/kubepods/burstable/poda168d109-4d40-4c50-8f89-957b9b0dc5d6/75082fa9e4783ecf3fc2e1ada7cd08fd2dd20d001d36e579e28e3cb00d312ad4
		convert to:
		kubepods/burstable/pod42736679-4475-43cf-afb4-e3744f4352fd/88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec

		When using cgroupfs as cgroup drvier and crio as container runtime wich burstable qos:
		/kubepods/burstable/poda168d109-4d40-4c50-8f89-957b9b0dc5d6/crio-75082fa9e4783ecf3fc2e1ada7cd08fd2dd20d001d36e579e28e3cb00d312ad4
		convert to:
		kubepods/burstable/pod42736679-4475-43cf-afb4-e3744f4352fd/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
	*/

	/*
		When using cgroupfs as cgroup driver and isula, docker, containerd as container runtime:
		1. The Burstable path looks like: kubepods/burstable/pod34152897-dbaf-11ea-8cb9-0653660051c3/88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
		2. The BestEffort path is in the form: kubepods/bestEffort/pod34152897-dbaf-11ea-8cb9-0653660051c3/88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
		3. The Guaranteed path is in the form: kubepods/pod34152897-dbaf-11ea-8cb9-0653660051c3/88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec

		When using cgroupfs as cgroup driver and crio as container runtime:
		1. The Burstable path looks like: kubepods/burstable/pod34152897-dbaf-11ea-8cb9-0653660051c3/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
		2. The BestEffort path is in the form: kubepods/besteffort/pod34152897-dbaf-11ea-8cb9-0653660051c3/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec
		3. The Guaranteed path is in the form: kubepods/pod34152897-dbaf-11ea-8cb9-0653660051c3/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec

		When using systemd as cgroup driver:
		1. The Burstable path looks like: kubepods.slice/kubepods-burstable.slice/kubepods-burstable-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec.scope
		2. The BestEffort path is in the form: kubepods.slice/kubepods-besteffort.slice/kubepods-besteffort-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec.scope
		3. The Guaranteed path is in the form: kubepods.slice/kubepods-podb895995a_e7e5_413e_9bc1_3c3895b3f233.slice/crio-88a791aa2090c928667579ea11a63f0ab67cf0be127743308a6e1a2130489dec.scope
	*/
	qosClassPath := ""
	switch corev1.PodQOSClass(container.getQos()) {
	case corev1.PodQOSGuaranteed:
	case corev1.PodQOSBurstable:
		qosClassPath = strings.ToLower(string(corev1.PodQOSBurstable))
	case corev1.PodQOSBestEffort:
		qosClassPath = strings.ToLower(string(corev1.PodQOSBestEffort))
	default:
		return ""
	}

	if cgroup.GetCgroupDriver() == constant.CgroupDriverSystemd {
		if qosClassPath == "" {
			switch containerEngineScopes[currentContainerEngines] {
			case constant.ContainerEngineContainerd, constant.ContainerEngineCrio, constant.ContainerEngineDocker, constant.ContainerEngineIsula:
				path = filepath.Join(
					constant.KubepodsCgroup+".slice",
					container.getPodCgroupDir(),
					containerEngineScopes[currentContainerEngines]+"-"+container.Id+".scope",
				)
			default:
				path = ""
			}
		} else {
			switch containerEngineScopes[currentContainerEngines] {
			case constant.ContainerEngineContainerd, constant.ContainerEngineCrio, constant.ContainerEngineDocker, constant.ContainerEngineIsula:

				path = filepath.Join(
					constant.KubepodsCgroup+".slice",
					constant.KubepodsCgroup+"-"+qosClassPath+".slice",
					container.getPodCgroupDir(),
					containerEngineScopes[currentContainerEngines]+"-"+container.Id+".scope",
				)
			default:
				path = ""
			}

		}

	} else {
		if qosClassPath == "" {
			switch containerEngineScopes[currentContainerEngines] {
			case constant.ContainerEngineContainerd, constant.ContainerEngineDocker, constant.ContainerEngineIsula:
				path = filepath.Join(
					constant.KubepodsCgroup,
					container.getPodCgroupDir(),
					container.Id,
				)

			case constant.ContainerEngineCrio:
				path = filepath.Join(
					constant.KubepodsCgroup,
					container.getPodCgroupDir(),
					containerEngineScopes[currentContainerEngines]+"-"+container.Id,
				)
			default:
				path = ""
			}
		} else {
			switch containerEngineScopes[currentContainerEngines] {
			case constant.ContainerEngineContainerd, constant.ContainerEngineDocker, constant.ContainerEngineIsula:
				path = filepath.Join(
					constant.KubepodsCgroup,
					qosClassPath,
					container.getPodCgroupDir(),
					container.Id,
				)
			case constant.ContainerEngineCrio:
				path = filepath.Join(
					constant.KubepodsCgroup,
					qosClassPath,
					container.getPodCgroupDir(),
					containerEngineScopes[currentContainerEngines]+"-"+container.Id,
				)

			default:
				path = ""
			}
		}
	}
	return path
}

// get container's pod's cgroup dir
func (container *NRIRawContainer) getPodCgroupDir() string {
	var podPath string
	if cgroup.GetCgroupDriver() == constant.CgroupDriverSystemd {
		podPath = strings.Split(container.Linux.CgroupsPath, ":")[0]
	} else {
		pathInfo := strings.Split(container.Linux.CgroupsPath, "/")
		podPath = pathInfo[len(pathInfo)-2]
	}
	return podPath
}

// get Qos through NRIRawContainer
func (container *NRIRawContainer) getQos() string {
	var podQosClass string
	podQosClass = "Guaranteed"
	if strings.Contains(container.Linux.CgroupsPath, "burstable") {
		podQosClass = "Burstable"
	}
	if strings.Contains(container.Linux.CgroupsPath, "besteffort") {
		podQosClass = "BestEffort"
	}

	return podQosClass
}

// AppliedConfiguration define container applied configure
type AppliedConfiguration struct {
	ApiVersion string
	Kind       string
	Metadata   interface{}
	Spec       map[string][]map[string]interface{}
}

// Resources define container resource info
type Resources struct {
	Limits   Limits
	Requests Requests
}

// Limits define container resource limit info
type Limits struct {
	Memory string
	Cpu    float64
}

// Requests define container resource request info
type Requests struct {
	Memory string
	Cpu    float64
}

// ResourceInfo define get resource interface
type ResourceInfo interface {
	getCpuInfo() float64
	getMemoryInfo() float64
}

// get container cpu request info
func (r *Requests) getCpuInfo() float64 {
	return r.Cpu
}

// get container memory request info
func (r *Requests) getMemoryInfo() float64 {
	var converter = func(value *resource.Quantity) float64 {
		return float64(value.MilliValue()) / 1000
	}

	q, _ := resource.ParseQuantity(r.Memory)

	return converter(&q)
}

// get container cpu limit info
func (r *Limits) getCpuInfo() float64 {
	return r.Cpu
}

// get container memory limit info
func (r *Limits) getMemoryInfo() float64 {
	var converter = func(value *resource.Quantity) float64 {
		return float64(value.MilliValue()) / 1000
	}

	q, _ := resource.ParseQuantity(r.Memory)

	return converter(&q)
}

// get container request and limit info from container applied configuration
func (container *NRIRawContainer) GetResourceMaps() (ResourceMap, ResourceMap) {
	configurations := container.Annotations[containerAppliedConfiguration]
	containerConf := &AppliedConfiguration{}
	_ = json.Unmarshal([]byte(configurations), containerConf)
	resourceInfo := &Resources{}
	if r, ok := containerConf.Spec[containerSpec]; ok {
		for _, containerSpec := range r {
			if containerSpec["name"] == container.Name {
				if containerResourceSpec, ok := containerSpec[resourcesSpec]; ok {
					if resource, err := json.Marshal(containerResourceSpec); err != nil {
						fmt.Printf("get container %v Resource failed ", container.Id)
					} else {
						if err = json.Unmarshal(resource, resourceInfo); err != nil {
							fmt.Printf("container %v data format error", container.Id)
						} else {
							fmt.Printf("get container %v Resource success", container.Id)
						}
					}
				} else {
					fmt.Printf("container %v spec resources info not exist", container.Id)
				}
			} else {
				continue
			}
		}
	} else {
		fmt.Printf("container %v spec containers info not exist", container.Id)
	}

	iterator := func(res ResourceInfo) ResourceMap {
		results := make(ResourceMap)
		results[ResourceCPU] = res.getCpuInfo()
		results[ResourceMem] = res.getMemoryInfo()
		return results
	}
	return iterator(&resourceInfo.Requests), iterator(&resourceInfo.Limits)
}

// get current container engine
func getEngineFromCgroup() {
	file, err := os.OpenFile(procSelfCgroupFile, os.O_RDONLY, fileMode)
	if err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		s := strings.Split(string(line), "/")
		containerEngine := strings.Split(s[len(s)-1], "-")[0]
		for engine, prefix := range containerEngineScopes {
			if containerEngine == prefix {
				currentContainerEngines = engine
				break
			}
		}
		if err == io.EOF {
			break
		}
	}
}

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
// Create: 2023-02-10
// Description: This file contains pod info and cgroup construct

package try

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"

	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/common/util"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// FakePod is used for pod testing
type FakePod struct {
	*typedef.PodInfo
	// Keys is cgroup key list
	Keys []*cgroup.Key
}

const idLen = 8

func genFakeContainerInfo(parentCGPath string) *typedef.ContainerInfo {
	containerID := genContainerID()
	var fakeContainer = &typedef.ContainerInfo{
		Name:             fmt.Sprintf("fakeContainer-%s", containerID[:idLen]),
		ID:               containerID,
		CgroupPath:       filepath.Join(parentCGPath, containerID),
		RequestResources: make(typedef.ResourceMap, 0),
		LimitResources:   make(typedef.ResourceMap, 0),
	}
	return fakeContainer
}

func genFakePodInfo(qosClass corev1.PodQOSClass) *typedef.PodInfo {
	podID := uuid.New().String()
	// generate fake pod info
	var fakePod = &typedef.PodInfo{
		Name:        fmt.Sprintf("fakepod-%s", podID[:idLen]),
		Namespace:   "test",
		UID:         constant.PodCgroupNamePrefix + podID,
		CgroupPath:  genRelativeCgroupPath(qosClass, podID),
		Annotations: make(map[string]string, 0),
	}
	return fakePod
}

// NewFakePod return fake pod info struct
func NewFakePod(keys []*cgroup.Key, qosClass corev1.PodQOSClass) *FakePod {
	return &FakePod{
		Keys:    keys,
		PodInfo: genFakePodInfo(qosClass),
	}
}

func (pod *FakePod) genFakePodCgroupPath() {
	if !util.PathExist(TestRoot) {
		MkdirAll(TestRoot, constant.DefaultDirMode).OrDie()
	}
	// generate fake cgroup path
	for _, key := range pod.Keys {
		// generate pod absolute cgroup path
		podCGPath := filepath.Join(TestRoot, key.SubSys, pod.CgroupPath)
		MkdirAll(podCGPath, constant.DefaultDirMode).OrDie()
		// create pod cgroup file, default qos level is "0"
		podCGFilePath := filepath.Join(podCGPath, key.FileName)
		WriteFile(podCGFilePath, []byte(string("0")), constant.DefaultFileMode).OrDie()
	}
	pod.genFakeContainersCgroupPath()
}

func (pod *FakePod) genFakeContainersCgroupPath() {
	if len(pod.IDContainersMap) == 0 {
		return
	}

	for _, key := range pod.Keys {
		podCGPath := filepath.Join(TestRoot, key.SubSys, pod.CgroupPath)
		if !util.PathExist(podCGPath) {
			return
		}
		for _, container := range pod.IDContainersMap {
			// generate container absolute cgroup path
			containerCGPath := filepath.Join(podCGPath, container.ID)
			MkdirAll(containerCGPath, constant.DefaultDirMode).OrDie()
			// create container cgroup file, default qos level is "0"
			containerCGFilePath := filepath.Join(containerCGPath, key.FileName)
			WriteFile(containerCGFilePath, []byte(string("0")), constant.DefaultFileMode).OrDie()
		}
	}
}

// WithContainers will generate containers under pod with container num
func (pod *FakePod) WithContainers(containerNum int) *FakePod {
	pod.IDContainersMap = make(map[string]*typedef.ContainerInfo, containerNum)
	for i := 0; i < containerNum; i++ {
		fakeContainer := genFakeContainerInfo(pod.CgroupPath)
		pod.IDContainersMap[fakeContainer.ID] = fakeContainer
	}
	pod.genFakeContainersCgroupPath()
	return pod
}

func genContainerID() string {
	const delimiter = "-"
	const containerIDLenLimit = 64
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	containerID := strings.ReplaceAll(uuid1, delimiter, "") + strings.ReplaceAll(uuid2, delimiter, "")
	if len(containerID) < containerIDLenLimit {
		containerID += strings.Repeat("a", containerIDLenLimit-len(containerID))
	}
	return containerID[:containerIDLenLimit]
}

// GenFakePod gen fake pod info
func GenFakePod(keys []*cgroup.Key, qosClass corev1.PodQOSClass) *FakePod {
	fakePod := NewFakePod(keys, qosClass)
	fakePod.genFakePodCgroupPath()
	return fakePod
}

// GenFakeBurstablePod generate pod with qos class burstable
func GenFakeBurstablePod(keys []*cgroup.Key) *FakePod {
	return GenFakePod(keys, corev1.PodQOSBurstable)
}

// GenFakeBestEffortPod generate pod with qos class best effort
func GenFakeBestEffortPod(keys []*cgroup.Key) *FakePod {
	return GenFakePod(keys, corev1.PodQOSBestEffort)
}

// GenFakeGuaranteedPod generate pod with qos class guaranteed
func GenFakeGuaranteedPod(keys []*cgroup.Key) *FakePod {
	return GenFakePod(keys, corev1.PodQOSGuaranteed)
}

// GenFakeOnlinePod generate online pod
func GenFakeOnlinePod(keys []*cgroup.Key) *FakePod {
	fakePod := GenFakeGuaranteedPod(keys)
	fakePod.Annotations[constant.PriorityAnnotationKey] = "false"
	return fakePod
}

// GenFakeOfflinePod generate offline pod
func GenFakeOfflinePod(keys []*cgroup.Key) *FakePod {
	fakePod := GenFakeBurstablePod(keys)
	fakePod.Annotations[constant.PriorityAnnotationKey] = "true"
	return fakePod
}

func genRelativeCgroupPath(qosClass corev1.PodQOSClass, id string) string {
	path := ""
	switch qosClass {
	case corev1.PodQOSGuaranteed:
	case corev1.PodQOSBurstable:
		path = strings.ToLower(string(corev1.PodQOSBurstable))
	case corev1.PodQOSBestEffort:
		path = strings.ToLower(string(corev1.PodQOSBestEffort))
	default:
		return ""
	}
	return filepath.Join(constant.KubepodsCgroup, path, constant.PodCgroupNamePrefix+id)
}

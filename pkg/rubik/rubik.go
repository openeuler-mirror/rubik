// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2021-05-20
// Description: This file is used for rubik struct

// Package rubik is for rubik struct
package rubik

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/coreos/go-systemd/daemon"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"isula.org/rubik/pkg/autoconfig"
	"isula.org/rubik/pkg/blkio"
	"isula.org/rubik/pkg/cachelimit"
	"isula.org/rubik/pkg/checkpoint"
	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	"isula.org/rubik/pkg/iocost"
	"isula.org/rubik/pkg/memory"
	"isula.org/rubik/pkg/perf"
	"isula.org/rubik/pkg/qos"
	"isula.org/rubik/pkg/quota"
	"isula.org/rubik/pkg/sync"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/util"
)

// Rubik defines rubik struct
type Rubik struct {
	config     *config.Config
	kubeClient *kubernetes.Clientset
	cpm        *checkpoint.Manager
	mm         *memory.MemoryManager
	nodeName   string
}

// NewRubik creates a new rubik object
func NewRubik(cfgPath string) (*Rubik, error) {
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		return nil, errors.Errorf("load config failed: %v", err)
	}

	if err = log.InitConfig(cfg.LogDriver, cfg.LogDir, cfg.LogLevel, int64(cfg.LogSize)); err != nil {
		return nil, errors.Errorf("init log config failed: %v", err)
	}

	r := &Rubik{
		config: cfg,
	}

	if err := r.initComponents(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Rubik) initComponents() error {
	if err := r.initKubeClient(); err != nil {
		return err
	}

	if err := r.initCheckpoint(); err != nil {
		return err
	}

	if err := r.initEventHandler(); err != nil {
		return err
	}

	if err := r.initNodeConfig(); err != nil {
		return err
	}

	if r.config.MemCfg.Enable {
		if err := r.initMemoryManager(); err != nil {
			return err
		}
	}

	return nil
}

// Monitor monitors shutdown signal
func (r *Rubik) Monitor() {
	<-config.ShutdownChan
	os.Exit(1)
}

// Sync sync pods qos level
func (r *Rubik) Sync() error {
	if !r.config.AutoCheck {
		return nil
	}
	return sync.Sync(r.cpm.ListOfflinePods())
}

// CacheLimit init cache limit module
func (r *Rubik) CacheLimit() error {
	if r.config.CacheCfg.Enable {
		if r.cpm == nil {
			return fmt.Errorf("checkpoint is not initialized before cachelimit")
		}
		return cachelimit.Init(r.cpm, &r.config.CacheCfg)
	}
	return nil
}

// QuotaBurst sync all pod's cpu burst quota
func (r *Rubik) QuotaBurst() {
	if r.config.AutoCheck {
		quota.SetPodsQuotaBurst(r.cpm.ListAllPods())
	}
}

// initKubeClient initialize kubeClient if autoconfig is enabled
func (r *Rubik) initKubeClient() error {
	conf, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return err
	}

	r.kubeClient = kubeClient
	log.Infof("the kube-client is initialized successfully")
	return nil
}

// initEventHandler initialize the event handler and set the rubik callback function corresponding to the pod event.
func (r *Rubik) initEventHandler() error {
	if r.kubeClient == nil {
		return fmt.Errorf("kube-client is not initialized")
	}

	autoconfig.Backend = r
	if err := autoconfig.Init(r.kubeClient, r.nodeName); err != nil {
		return err
	}

	log.Infof("the event-handler is initialized successfully")
	return nil
}

func (r *Rubik) initMemoryManager() error {
	mm, err := memory.NewMemoryManager(r.cpm, r.config.MemCfg)
	if err != nil {
		return err
	}

	r.mm = mm
	log.Infof("init memory manager ok")
	return nil
}

func (r *Rubik) initCheckpoint() error {
	if r.kubeClient == nil {
		return fmt.Errorf("kube-client is not initialized")
	}

	cpm := checkpoint.NewManager(r.config.CgroupRoot)
	node := os.Getenv(constant.NodeNameEnvKey)
	if node == "" {
		return fmt.Errorf("missing %s", constant.NodeNameEnvKey)
	}

	r.nodeName = node
	const specNodeNameFormat = "spec.nodeName=%s"
	pods, err := r.kubeClient.CoreV1().Pods("").List(context.Background(),
		metav1.ListOptions{FieldSelector: fmt.Sprintf(specNodeNameFormat, node)})
	if err != nil {
		return err
	}
	cpm.SyncFromCluster(pods.Items)

	r.cpm = cpm
	log.Infof("the checkpoint is initialized successfully")
	return nil
}

func (r *Rubik) initNodeConfig() error {
	for _, nc := range r.config.NodeConfig {
		if nc.NodeName == r.nodeName && nc.IOcostEnable {
			iocost.SetIOcostEnable(true)
			return iocost.ConfigIOcost(nc.IOcostConfig)
		}
	}
	return nil
}

// AddEvent handle add event from informer
func (r *Rubik) AddEvent(pod *corev1.Pod) {
	// Rubik does not process add event of pods that are not in the running state.
	if pod.Status.Phase != corev1.PodRunning {
		return
	}
	r.cpm.AddPod(pod)

	pi := r.cpm.GetPod(pod.UID)
	if err := qos.SetQosLevel(pi); err != nil {
		log.Errorf("AddEvent handle error: %v", err)
	}
	quota.SetPodQuotaBurst(pi)

	if r.config.BlkioCfg.Enable {
		blkio.SetBlkio(pod)
	}
	if r.config.MemCfg.Enable {
		r.mm.UpdateConfig(pi)
	}
	iocost.SetPodWeight(pi)
}

// UpdateEvent handle update event from informer
func (r *Rubik) UpdateEvent(oldPod *corev1.Pod, newPod *corev1.Pod) {
	// Rubik does not process updates of pods that are not in the running state
	// if the container is not running, delete it.
	if newPod.Status.Phase != corev1.PodRunning {
		r.cpm.DelPod(newPod.UID)
		return
	}

	// after the Rubik is started, the pod adding events are transferred through the update handler of Kubernetes.
	if !r.cpm.PodExist(newPod.UID) {
		r.cpm.AddPod(newPod)

		if r.config.BlkioCfg.Enable {
			blkio.SetBlkio(newPod)
		}

		pi := r.cpm.GetPod(newPod.UID)
		quota.SetPodQuotaBurst(pi)

		if r.config.MemCfg.Enable {
			r.mm.UpdateConfig(pi)
		}
		iocost.SetPodWeight(pi)
	} else {
		opi := r.cpm.GetPod(oldPod.UID)
		r.cpm.UpdatePod(newPod)
		if r.config.BlkioCfg.Enable {
			blkio.WriteBlkio(oldPod, newPod)
		}
		npi := r.cpm.GetPod(newPod.UID)
		quota.UpdatePodQuotaBurst(opi, npi)
		iocost.SetPodWeight(npi)
	}

	cpmPod := r.cpm.GetPod(newPod.UID)
	if err := qos.UpdateQosLevel(cpmPod); err != nil {
		log.Errorf("UpdateEvent handle error: %v", err)
	}
}

// DeleteEvent handle update event from informer
func (r *Rubik) DeleteEvent(pod *corev1.Pod) {
	r.cpm.DelPod(pod.UID)
}

func run(fcfg string) int {
	rubik, err := NewRubik(fcfg)
	if err != nil {
		fmt.Printf("new rubik failed: %v\n", err)
		return constant.ErrCodeFailed
	}

	if rubik.mm != nil {
		rubik.mm.Run()
	}

	log.Infof("perf hw support = %v", perf.HwSupport())
	if err = rubik.CacheLimit(); err != nil {
		log.Errorf("cache limit init error: %v", err)
		return constant.ErrCodeFailed
	}

	rubik.QuotaBurst()
	if err = rubik.Sync(); err != nil {
		log.Errorf("sync qos level failed: %v", err)
	}

	log.Logf("Start rubik with cfg\n%v", rubik.config)
	go signalHandler()

	// Notify systemd that rubik is ready SdNotify() only tries to
	// notify if the NOTIFY_SOCKET environment is set, so it's
	// safe to call when systemd isn't present.
	// Ignore the return values here because they're not valid for
	// platforms that don't use systemd.  For platforms that use
	// systemd, rubik doesn't log if the notification failed.
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)

	rubik.Monitor()

	return 0
}

// Run start rubik server
func Run(fcfg string) int {
	unix.Umask(constant.DefaultUmask)
	if len(os.Args) > 1 {
		fmt.Println("args not allowed")
		return constant.ErrCodeFailed
	}

	lock, err := util.CreateLockFile(constant.LockFile)
	if err != nil {
		fmt.Printf("set rubik lock failed: %v, check if there is another rubik running\n", err)
		return constant.ErrCodeFailed
	}

	ret := run(fcfg)

	util.RemoveLockFile(lock, constant.LockFile)
	return ret
}

func signalHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	var forceCount int32 = 3
	for sig := range signalChan {
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			if atomic.AddInt32(&config.ShutdownFlag, 1) == 1 {
				log.Infof("Signal %v received and starting exit...", sig)
				close(config.ShutdownChan)
			}

			if atomic.LoadInt32(&config.ShutdownFlag) >= forceCount {
				log.Infof("3 interrupts signal received, forcing rubik shutdown")
				os.Exit(1)
			}
		}
	}
}

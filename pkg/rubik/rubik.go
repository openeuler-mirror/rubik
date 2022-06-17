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
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

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
	"isula.org/rubik/pkg/httpserver"
	"isula.org/rubik/pkg/perf"
	"isula.org/rubik/pkg/qos"
	"isula.org/rubik/pkg/sync"
	log "isula.org/rubik/pkg/tinylog"
	"isula.org/rubik/pkg/util"
	"isula.org/rubik/pkg/workerpool"
)

// Rubik defines rubik struct
type Rubik struct {
	server     *http.Server
	pool       *workerpool.WorkerPool
	sock       *net.Listener
	config     *config.Config
	kubeClient *kubernetes.Clientset
	cpm        *checkpoint.Manager
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

func run(fcfg string) int {
	rubik, err := NewRubik(fcfg)
	if err != nil {
		fmt.Printf("new rubik failed: %v\n", err)
		return constant.ErrCodeFailed
	}

	rubik.initAutoConfig()

	log.Infof("perf hw support = %v", perf.HwSupport())
	if err = rubik.CacheLimit(); err != nil {
		log.Errorf("cache limit init error: %v", err)
		return constant.ErrCodeFailed
	}

	if err = rubik.Sync(); err != nil {
		log.Errorf("sync qos level failed: %v", err)
	}

	go signalHandler()
	go rubik.Monitor()

	if err = rubik.Serve(); err != nil {
		log.Errorf("http serve failed: %v", err)
		return constant.ErrCodeFailed
	}
	return 0
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

	sock, err := httpserver.NewSock()
	if err != nil {
		return nil, errors.Errorf("new sock failed: %v", err)
	}
	server, pool := httpserver.NewServer()

	r := &Rubik{
		server: server,
		pool:   pool,
		sock:   sock,
		config: cfg,
	}

	if err := r.initKubeClient(); err != nil {
		return nil, err
	}

	if err := r.initEventHandler(); err != nil {
		return nil, err
	}

	if err := r.initCheckpoint(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Rubik) initAutoConfig() {
	if r.config.AutoConfig {
		autoconfig.Init(r.kubeClient)
	}
}

// Sync sync pods qos level
func (r *Rubik) Sync() error {
	return sync.Sync(r.config.AutoCheck, r.kubeClient)
}

// Monitor monitors shutdown signal
func (r *Rubik) Monitor() {
	<-config.ShutdownChan
	r.Shutdown()
	os.Exit(1)
}

// Shutdown waits for tasks done or timeout to shutdown rubik
func (r *Rubik) Shutdown() {
	poolDone := make(chan struct{})
	go func() {
		for {
			if len(r.pool.TasksChan) == 0 && atomic.LoadInt32(&r.pool.WorkerBusy) == 0 {
				close(poolDone)
				return
			}
			time.Sleep(time.Second)
		}
	}()

	waitTime := 15
	select {
	case _, ok := <-poolDone:
		if !ok {
			log.Infof("All tasks finished and exit")
		}
	case <-time.After(time.Duration(waitTime) * time.Second):
		log.Errorf("Tasks not finished after 15 seconds, force exit")
	}

	log.DropError(r.server.Shutdown(context.Background()))
}

// Serve starts http server
func (r *Rubik) Serve() error {
	log.Logf("Start http server %s with cfg\n%v", constant.RubikSock, r.config)
	return r.server.Serve(*r.sock)
}

// CacheLimit init cache limit module
func (r *Rubik) CacheLimit() error {
	if r.config.CacheCfg.Enable {
		return cachelimit.Init(&r.config.CacheCfg)
	}
	return nil
}

// initKubeClient initialize kubeClient if autoconfig is enabled
func (r *Rubik) initKubeClient() error {
	r.kubeClient = nil
	if !r.config.AutoConfig {
		return nil
	}

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
	if !r.config.AutoConfig {
		return nil
	}

	if r.kubeClient == nil {
		return fmt.Errorf("kube-client is not initialized")
	}

	autoconfig.Backend = r
	if err := autoconfig.Init(r.kubeClient); err != nil {
		return err
	}
	log.Infof("the event-handler is initialized successfully")
	return nil
}

func (r *Rubik) initCheckpoint() error {
	r.cpm = nil
	if !r.config.AutoConfig {
		log.Infof("the autoconfig must be enabled for checkpoint")
		return nil
	}
	if r.kubeClient == nil {
		return fmt.Errorf("kube-client not initialized")
	}

	cpm := checkpoint.NewManager()

	node := os.Getenv(constant.NodeNameEnvKey)
	if node == "" {
		return fmt.Errorf("missing %s", constant.NodeNameEnvKey)
	}
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

// AddEvent handle add event from informer
func (r *Rubik) AddEvent(pod *corev1.Pod) {
	// Rubik does not process add event of pods that are not in the running state.
	if pod.Status.Phase != corev1.PodRunning {
		return
	}
	r.cpm.AddPod(pod)
	qos.SetQosLevel(pod)
	if r.config.BlkConfig.Limit {
		blkio.SetBlkio(pod)
	}
}

// UpdateEvent handle update event from informer
func (r *Rubik) UpdateEvent(oldPod *corev1.Pod, newPod *corev1.Pod) {
	// Rubik does not process updates of pods that are not in the running state
	// if the container is not running, delete it.
	if newPod.Status.Phase != corev1.PodRunning {
		r.cpm.DelPod(newPod.UID)
		return
	}

	qos.UpdateQosLevel(oldPod, newPod)

	// after the Rubik is started, the pod adding events are transferred through the update handler of Kubernetes.
	if !r.cpm.PodExist(newPod.UID) {
		r.cpm.AddPod(newPod)
		if r.config.BlkConfig.Limit {
			blkio.SetBlkio(newPod)
		}
	} else {
		r.cpm.UpdatePod(newPod)
		if r.config.BlkConfig.Limit {
			blkio.WriteBlkio(oldPod, newPod)
		}
	}
}

// DeleteEvent handle update event from informer
func (r *Rubik) DeleteEvent(pod *corev1.Pod) {
	r.cpm.DelPod(pod.UID)
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

// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: JingRui
// Create: 2022-01-26
// Description: cgroup perf stats

// Package perf provide perf functions
package perf

import (
	"encoding/binary"
	"errors"
	"path/filepath"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"

	"isula.org/rubik/pkg/config"
	"isula.org/rubik/pkg/constant"
	log "isula.org/rubik/pkg/tinylog"
)

var (
	hwSupport = false
)

//HwSupport tells if the os support perf hw pmu events
func HwSupport() bool {
	return hwSupport
}

// PerfStat is perf stat info
type PerfStat struct {
	Instructions    uint64
	CpuCycles       uint64
	CacheMisses     uint64
	CacheReferences uint64
}

type cgEvent struct {
	cgfd   int
	cpu    int
	fds    map[uint64]int
	leader int
}

func newEvent(cgfd, cpu int) (*cgEvent, error) {
	e := cgEvent{
		cgfd:   cgfd,
		cpu:    cpu,
		fds:    make(map[uint64]int),
		leader: -1,
	}

	configList := []uint64{
		unix.PERF_COUNT_HW_INSTRUCTIONS,
		unix.PERF_COUNT_HW_CPU_CYCLES,
		unix.PERF_COUNT_HW_CACHE_REFERENCES,
		unix.PERF_COUNT_HW_CACHE_MISSES,
	}
	for _, config := range configList {
		if err := e.openHardware(config); err != nil {
			return nil, err
		}
	}

	return &e, nil
}

func (e *cgEvent) openHardware(config uint64) error {
	attr := unix.PerfEventAttr{
		Type:   unix.PERF_TYPE_HARDWARE,
		Config: config,
	}

	fd, err := unix.PerfEventOpen(&attr, e.cgfd, e.cpu, e.leader, unix.PERF_FLAG_PID_CGROUP|unix.PERF_FLAG_FD_CLOEXEC)
	if err != nil {
		log.Errorf("perf open for config:0x%x cpu:%d failed: %v", config, e.cpu, err)
		return err
	}

	if e.leader == -1 {
		e.leader = fd
	}
	e.fds[config] = fd
	return nil
}

func (e *cgEvent) start() error {
	if err := unix.IoctlSetInt(e.leader, unix.PERF_EVENT_IOC_RESET, 0); err != nil {
		return err
	}
	if err := unix.IoctlSetInt(e.leader, unix.PERF_EVENT_IOC_ENABLE, 0); err != nil {
		return err
	}
	return nil
}

func (e *cgEvent) stop() error {
	if err := unix.IoctlSetInt(e.leader, unix.PERF_EVENT_IOC_DISABLE, 0); err != nil {
		return err
	}
	return nil
}

func (e *cgEvent) read(config uint64) uint64 {
	var val uint64

	p := make([]byte, 64)
	num, err := unix.Read(e.fds[config], p)
	if err != nil {
		log.Errorf("read perf data failed %v", err)
		return 0
	}

	if num != int(unsafe.Sizeof(val)) {
		log.Errorf("invalid perf data length %d", num)
		return 0
	}

	return binary.LittleEndian.Uint64(p)
}

func (e *cgEvent) destroy() {
	for _, fd := range e.fds {
		unix.Close(fd)
	}
}

// perf is perf manager
type perf struct {
	Events []*cgEvent
	Cgfd   int
}

// newPerf create the perf manager
func newPerf(cgpath string) (*perf, error) {
	cgfd, err := unix.Open(cgpath, unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	p := perf{
		Cgfd:   cgfd,
		Events: make([]*cgEvent, 0),
	}

	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
		e, err := newEvent(cgfd, cpu)
		if err != nil {
			continue
		}
		p.Events = append(p.Events, e)
	}

	if len(p.Events) == 0 {
		return nil, errors.New("new perf event for all cpus failed")
	}

	return &p, nil
}

// Start start perf
func (p *perf) Start() error {
	for _, e := range p.Events {
		if err := e.start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop stop perf
func (p *perf) Stop() error {
	for _, e := range p.Events {
		if err := e.stop(); err != nil {
			return err
		}
	}
	return nil
}

// Read read perf result
func (p *perf) Read() PerfStat {
	var stat PerfStat
	for _, e := range p.Events {
		stat.Instructions += e.read(unix.PERF_COUNT_HW_INSTRUCTIONS)
		stat.CpuCycles += e.read(unix.PERF_COUNT_HW_CPU_CYCLES)
		stat.CacheReferences += e.read(unix.PERF_COUNT_HW_CACHE_REFERENCES)
		stat.CacheMisses += e.read(unix.PERF_COUNT_HW_CACHE_MISSES)
	}
	return stat
}

// Destroy free resources
func (p *perf) Destroy() {
	for _, e := range p.Events {
		e.destroy()
	}
}

// CgroupStat report perf stat for cgroup
func CgroupStat(cgpath string, dur time.Duration) (*PerfStat, error) {
	p, err := newPerf(cgpath)
	if err != nil {
		log.Errorf("perf init failed: %v", err)
		return nil, err
	}

	defer func() {
		p.Destroy()
	}()

	if err := p.Start(); err != nil {
		log.Errorf("perf start failed: %v", err)
		return nil, err
	}
	time.Sleep(dur)
	if err := p.Stop(); err != nil {
		log.Errorf("perf stop failed: %v", err)
		return nil, err
	}

	stat := p.Read()
	return &stat, nil
}

func init() {
	_, err := CgroupStat(filepath.Join(config.CgroupRoot, "perf_event", constant.KubepodsCgroup), time.Millisecond)
	if err == nil {
		hwSupport = true
	}
	log.Infof("perf hw support = %v", hwSupport)
}
// Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jiaqi Yang
// Date: 2023-03-02
// Description: This file is used for testing quota burst

// Package  quotaburst is for Quota Burst
package quotaburst

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
	"isula.org/rubik/pkg/podmanager"
	"isula.org/rubik/pkg/services/helper"
	"isula.org/rubik/test/try"
)

const (
	moduleName = "quotaburst"
)

var (
	cfsBurstUs  = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_burst_us"}
	cfsQuotaUs  = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_quota_us"}
	cfsPeriodUs = &cgroup.Key{SubSys: "cpu", FileName: "cpu.cfs_period_us"}
)

// TestBurst_AddPod tests AddPod
func TestBurst_AddPod(t *testing.T) {
	type args struct {
		pod   *try.FakePod
		burst string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TC-1: set burst successfully",
			args: args{
				pod: try.GenFakeGuaranteedPod(map[*cgroup.Key]string{
					cfsBurstUs:  "0",
					cfsPeriodUs: "100000",
					cfsQuotaUs:  "100000",
				}).WithContainers(1),
				burst: "1000",
			},
			wantErr: false,
		},
		{
			name: "TC-2.1: parseQuotaBurst invalid burst < 0",
			args: args{
				pod:   try.GenFakeGuaranteedPod(map[*cgroup.Key]string{}),
				burst: "-100",
			},
			wantErr: true,
		},
		{
			name: "TC-2.2: parseQuotaBurst invalid burst non int64",
			args: args{
				pod:   try.GenFakeGuaranteedPod(map[*cgroup.Key]string{}),
				burst: "abc",
			},
			wantErr: true,
		},
		{
			name: "TC-3.1: matchQuota quota file not existed",
			args: args{
				pod: try.GenFakeGuaranteedPod(map[*cgroup.Key]string{
					cfsBurstUs: "0",
				}).WithContainers(1),
				burst: "10000",
			},
			wantErr: false,
		},
		{
			name: "TC-3.2: matchQuota quota value invalid",
			args: args{
				pod: try.GenFakeGuaranteedPod(map[*cgroup.Key]string{
					cfsBurstUs: "0",
					cfsQuotaUs: "abc",
				}),
				burst: "10000",
			},
			wantErr: false,
		},
		{
			name: "TC-3.3: matchQuota period file not existed",
			args: args{
				pod: try.GenFakeGuaranteedPod(map[*cgroup.Key]string{
					cfsBurstUs: "0",
					cfsQuotaUs: "10000",
				}),
				burst: "10000",
			},
			wantErr: false,
		},
		{
			name: "TC-3.4: matchQuota period value invalid",
			args: args{
				pod: try.GenFakeGuaranteedPod(map[*cgroup.Key]string{
					cfsBurstUs:  "0",
					cfsPeriodUs: "abc",
					cfsQuotaUs:  "10000",
				}),
				burst: "10000",
			},
			wantErr: false,
		},
		{
			name: "TC-3.5: matchQuota quota > max",
			args: args{
				pod: try.GenFakeGuaranteedPod(map[*cgroup.Key]string{
					cfsBurstUs:  "0",
					cfsPeriodUs: "0",
					cfsQuotaUs:  "10000",
				}),
				burst: "10000",
			},
			wantErr: false,
		},
		{
			name: "TC-3.6: matchQuota quota < burst",
			args: args{
				pod: try.GenFakeGuaranteedPod(map[*cgroup.Key]string{
					cfsBurstUs:  "0",
					cfsPeriodUs: "10000",
					cfsQuotaUs:  "10000",
				}).WithContainers(1),
				burst: "200000000",
			},
			wantErr: false,
		},
		{
			name: "Tc-4: burst file not existed",
			args: args{
				pod:   try.GenFakeGuaranteedPod(map[*cgroup.Key]string{}),
				burst: "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := Burst{ServiceBase: helper.ServiceBase{Name: moduleName}}
			if tt.args.burst != "" {
				tt.args.pod.Annotations[constant.QuotaBurstAnnotationKey] = tt.args.burst
			}
			if err := conf.AddPod(tt.args.pod.PodInfo); (err != nil) != tt.wantErr {
				t.Errorf("Burst.AddPod() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.pod.CleanPath().OrDie()
		})
	}
	cgroup.InitMountDir(constant.DefaultCgroupRoot)
}

// TestOther tests other function
func TestOther(t *testing.T) {
	const tcName = "TC1-test Other"
	t.Run(tcName, func(t *testing.T) {
		got := Burst{ServiceBase: helper.ServiceBase{Name: moduleName}}
		assert.NoError(t, got.DeletePod(&typedef.PodInfo{}))
		assert.Equal(t, moduleName, got.ID())
	})
}

// TestBurst_UpdatePod tests UpdatePod
func TestBurst_UpdatePod(t *testing.T) {
	type args struct {
		oldPod *typedef.PodInfo
		newPod *typedef.PodInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TC1-same burst",
			args: args{
				oldPod: &typedef.PodInfo{
					Annotations: map[string]string{
						constant.QuotaBurstAnnotationKey: "10",
					},
				},
				newPod: &typedef.PodInfo{
					Annotations: map[string]string{
						constant.QuotaBurstAnnotationKey: "10",
					},
				},
			},
		},
		{
			name: "TC2-different burst",
			args: args{
				oldPod: &typedef.PodInfo{
					Annotations: make(map[string]string),
				},
				newPod: &typedef.PodInfo{
					Annotations: map[string]string{
						constant.QuotaBurstAnnotationKey: "10",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := Burst{ServiceBase: helper.ServiceBase{Name: moduleName}}
			if err := conf.UpdatePod(tt.args.oldPod, tt.args.newPod); (err != nil) != tt.wantErr {
				t.Errorf("Burst.UpdatePod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBurst_PreStart tests PreStart
func TestBurst_PreStart(t *testing.T) {
	type args struct {
		viewer api.Viewer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TC1-set pod",
			args: args{
				viewer: &podmanager.PodManager{
					Pods: &podmanager.PodCache{
						Pods: map[string]*typedef.PodInfo{
							"testPod1": {},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := Burst{ServiceBase: helper.ServiceBase{Name: moduleName}}
			if err := conf.PreStart(tt.args.viewer); (err != nil) != tt.wantErr {
				t.Errorf("Burst.PreStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
# rubik licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Create: 2022-06-28
# Description: 调整带业务的pod为离线业务

set -x
top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh

function pre_fun() {
    rubik_id=$(run_rubik)
    fn_check_result $? 0
    pod_id=$(docker run -tid --cgroup-parent /kubepods "${PAUSE_IMAGE}")
    fn_check_result $? 0
    containers=()
    total_num=50
}

function test_fun() {
    cgroup_path="kubepods/${pod_id}"
    qos_level_offline=-1
    # generate json data
    data_offline=$(gen_single_pod_json "$pod_id" "$cgroup_path" $qos_level_offline)

    # create containers in the pod
    for i in $(seq 1 ${total_num}); do
        containers[$i]=$(docker run -tid --cgroup-parent /kubepods/"${pod_id}" "${OPENEULER_IMAGE}" bash)
        fn_check_result $? 0
    done

    # set pod to offline
    result=$(rubik_qos "$data_offline")
    fn_check_result $? 0
    fn_check_string_not_contain "set qos failed" "$result"

    # check qos level for containers
    for i in $(seq 1 ${total_num}); do
        cpu_qos=$(cat /sys/fs/cgroup/cpu/"${cgroup_path}"/"${containers[$i]}"/cpu.qos_level)
        mem_qos=$(cat /sys/fs/cgroup/memory/"${cgroup_path}"/"${containers[$i]}"/memory.qos_level)
        fn_check_result "$cpu_qos" "$qos_level_offline"
        fn_check_result "$mem_qos" "$qos_level_offline"
    done
}

function post_fun() {
    docker rm -f "$rubik_id"
    fn_check_result $? 0
    docker rm -f "${containers[@]}"
    # Deleting multiple containers may time out.
    [ "$?" -ne "0" ] && docker rm -f "${containers[@]}"
    docker rm -f "$pod_id"
    fn_check_result $? 0
    docker ps -a | grep -v "CONTAINER"
    fn_check_result $? 1 "cleanup"
    exit "$exit_flag"
}

pre_fun
test_fun
post_fun

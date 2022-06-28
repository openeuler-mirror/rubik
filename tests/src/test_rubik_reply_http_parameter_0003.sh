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
# Description: http接口无效参数测试

set -x
top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh

function pre_fun() {
    rubik_id=$(run_rubik)
    fn_check_result $? 0
    logfile=${RUBIK_TEST_ROOT}/rubik_"$rubik_id"_log
    docker logs -f "$rubik_id" > "$logfile" &
    pod_id=$(docker run -tid --cgroup-parent /kubepods "${PAUSE_IMAGE}")
    fn_check_result $? 0
}

function test_fun() {
    # generate validate json data
    container_id=$(docker run -tid --cgroup-parent /kubepods/"${pod_id}" "${OPENEULER_IMAGE}" bash)
    cgroup_path="kubepods/${pod_id}"
    qos_level=-1
    validate_data=$(gen_single_pod_json "$pod_id" "$cgroup_path" $qos_level)

    # generate invalid data
    # case1: pod id not exist
    pod_id_not_exist=$(cat /proc/sys/kernel/random/uuid)
    pod_id_not_exist_cgroup="kubepods/${pod_id_not_exist}"
    invalid_data1=$(gen_single_pod_json "$pod_id_not_exist" "$pod_id_not_exist_cgroup" $qos_level)
    result=$(rubik_qos "$invalid_data1")
    fn_check_result $? 0
    fn_check_string_contain "set qos failed" "$result"
    grep "set qos level error" "$logfile"
    fn_check_result $? 0
    echo > "$logfile"

    # case2: cgroup path not exist
    cgroup_path_not_exist="kubepods/cgroup/path/not/exist"
    invalid_data2=$(gen_single_pod_json "$pod_id" "$cgroup_path_not_exist" $qos_level)
    result=$(rubik_qos "$invalid_data2")
    fn_check_result $? 0
    fn_check_string_contain "set qos failed" "$result"
    grep "set qos level error" "$logfile"
    fn_check_result $? 0
    echo > "$logfile"

    # case3: super long cgroup path
    cgroup_path_super_long="kubepods/$(long_char 10000)"
    invalid_data3=$(gen_single_pod_json "$pod_id" "$cgroup_path_super_long" $qos_level)
    result=$(rubik_qos "$invalid_data3")
    fn_check_result $? 0
    fn_check_string_contain "set qos failed" "$result"
    grep -i "length of cgroup path exceeds max limit 4096" "$logfile"
    fn_check_result $? 0
    echo > "$logfile"

    # case4: invalid qos level
    qos_level_invalid=-999
    invalid_data4=$(gen_single_pod_json "$pod_id" "$cgroup_path" $qos_level_invalid)
    result=$(rubik_qos "$invalid_data4")
    fn_check_result $? 0
    fn_check_string_contain "set qos failed" "$result"
    grep -i "Invalid qos level number" "$logfile"
    fn_check_result $? 0
    echo > "$logfile"

    # generate invalid data
    # case5: pod id empty
    pid_id_empty=""
    pod_id_empty_cgroup="kubepods/${pod_id_empty}"
    invalid_data5=$(gen_single_pod_json "$pod_id_empty" "$pod_id_empty_cgroup" $qos_level)
    result=$(rubik_qos "$invalid_data5")
    fn_check_result $? 0
    fn_check_string_contain "set qos failed" "$result"
    grep "invalid cgroup path" "$logfile"
    fn_check_result $? 0
    echo > "$logfile"

    # rubik will success with validate data
    result=$(rubik_qos "$validate_data")
    fn_check_result $? 0
    fn_check_string_not_contain "set qos failed" "$result"
    echo > "$logfile"
}

function post_fun() {
    clean_all
    docker rm -f "$container_id"
    fn_check_result $? 0
    docker rm -f "$pod_id"
    fn_check_result $? 0
    rm -f "$logfile"
    exit "$exit_flag"
}

pre_fun
test_fun
post_fun

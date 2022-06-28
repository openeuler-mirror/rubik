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
    pod_id=$(docker run -tid --cgroup-parent /kubepods "${PAUSE_IMAGE}")
    fn_check_result $? 0
}

function test_fun() {
    # generate json data
    container_id=$(docker run -tid --cgroup-parent /kubepods/"${pod_id}" "${OPENEULER_IMAGE}" bash)
    cgroup_path="kubepods/${pod_id}"
    qos_level=-1
    data=$(gen_single_pod_json "$pod_id" "$cgroup_path" $qos_level)

    # construct wrong handlers
    super_long_handler="http://longlonglong/$(long_char 1000)"
    https_handler="https://localhost/"
    special_char_handler="!@#$!@#!@#"
    invalid_handler="http://localhost/not_exist"
    debug_handler="http://localhost/debug/pprof/"
    profile_handler="http://localhost/debug/pprof/profile"

    result=$(curl_cmd "$super_long_handler" "$data" "POST")
    fn_check_string_contain "404 page not found" "$result"

    curl_cmd "$https_handler" "$data" "POST"
    fn_check_result_noeq $? 0

    curl_cmd "$special_char_handler" "$data" "POST"
    fn_check_result_noeq $? 0

    result=$(curl_cmd "$invalid_handler" "$data" "POST")
    fn_check_string_contain "404 page not found" "$result"

    result=$(curl_cmd "$debug_handler" "$data" "POST")
    fn_check_string_contain "404 page not found" "$result"

    result=$(curl_cmd "$profile_handler" "$data" "POST")
    fn_check_string_contain "404 page not found" "$result"
}

function post_fun() {
    docker rm -f "$rubik_id"
    fn_check_result $? 0
    docker rm -f "$container_id"
    fn_check_result $? 0
    docker rm -f "$pod_id"
    fn_check_result $? 0
    exit "$exit_flag"
}

pre_fun
test_fun
post_fun

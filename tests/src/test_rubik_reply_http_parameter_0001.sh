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
# Description: http接口有效参数测试

set -x
top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh

pre_fun() {
    rubik_id=$(run_rubik)
    fn_check_result $? 0
    pod_id=$(docker run -tid --cgroup-parent /kubepods "${PAUSE_IMAGE}")
    fn_check_result $? 0
}

test_fun() {
    # generate json data
    container_id=$(docker run -tid --cgroup-parent /kubepods/"${pod_id}" "${OPENEULER_IMAGE}" bash)
    cgroup_path="kubepods/${pod_id}"
    qos_level=-1
    data=$(gen_single_pod_json "$pod_id" "$cgroup_path" $qos_level)

    # check http://localhost/
    result=$(rubik_qos "$data")
    fn_check_result $? 0
    fn_check_string_not_contain "set qos failed" "$result"

    # check http://localhost/ping
    result=$(rubik_ping)
    fn_check_result $? 0
    fn_check_string_contain "ok" "$result"

    # check http://localhost/version
    result=$(rubik_version)
    fn_check_result $? 0
    fn_check_string_contain "Version" "$result"
}

post_fun() {
    clean_all
    docker rm -f "${container_id}"
    fn_check_result $? 0
    docker rm -f "${pod_id}"
    fn_check_result $? 0
    exit "$exit_flag"
}

pre_fun
test_fun
post_fun

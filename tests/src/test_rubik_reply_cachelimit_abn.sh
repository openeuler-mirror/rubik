#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
# rubik licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Create: 2022-05-19
# Description: rubik cachelimit 0002

set -x
top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh

function pre_fun() {
    kernel_check CACHE
    if [ $? -ne 0 ]; then
        echo "Kernel not supported, skip test"
        exit "${SKIP_FLAG}"
    fi
    run_rubik
}

# pod not exist
function test_fun() {
    local pod_name=podrubiktestpod
    local cgroup_path=kubepods/podrubiktestpod
    json_data=$(gen_single_pod_json ${pod_name} ${cgroup_path})
    result=$(rubik_qos "${json_data}")
    if ! echo "$result" | grep "set qos failed"; then
        ((exit_flag++))
    fi
    rmdir /sys/fs/resctrl/rubik_*
}

function post_fun() {
    clean_all
    if [[ $exit_flag -eq 0 ]]; then
        echo "PASS"
    else
        echo "FAILED"
    fi
    exit "$exit_flag"
}

pre_fun
test_fun
post_fun

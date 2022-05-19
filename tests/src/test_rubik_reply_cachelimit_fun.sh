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
# Description: rubik cachelimit 0001

top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh
cg=kubepods/podrubiktestpod

function test_reply() {
    mkdir -p /sys/fs/cgroup/cpu/${cg}
    mkdir -p /sys/fs/cgroup/memory/${cg}
    echo $$ > /sys/fs/cgroup/cpu/${cg}/cgroup.procs
    result=$(curl -s -H "Accept: application/json" -H "Content-type: application/json" -X POST --data '{"Pods": {"podrubiktestpod": {"CgroupPath": "kubepods/podrubiktestpod","QosLevel": -1,"CacheLimitLevel": "max"}}}' --unix-socket /run/rubik/rubik.sock http://localhost/)
    if [[ $? -ne 0 ]]; then
        ((exit_flag++))
    fi
    cat /sys/fs/resctrl/rubik_max/tasks | grep $$
    if [[ $? -ne 0 ]]; then
        ((exit_flag++))
    fi
}

function generate_config() {
    generate_config_file
    sed -i 's/\"enable\": false/\"enable\": true/' /var/lib/rubik/config.json
}

function clean() {
    echo $$ > /sys/fs/cgroup/cpu/cgroup.procs
    echo $$ > /sys/fs/resctrl/tasks
    rmdir /sys/fs/cgroup/cpu/${cg}
    rmdir /sys/fs/cgroup/memory/${cg}
    rmdir /sys/fs/resctrl/rubik_*
    rm -f /var/lib/rubik/config.json
}

env_check
if [ $? -ne 0 ]; then
    echo "Kernel not supported, skip test"
    exit 0
fi
generate_config
set_up
test_reply > /dev/null
tear_down
clean

if [[ $exit_flag -eq 0 ]]; then
    echo "PASS"
else
    echo "FAILED"
fi
exit "$exit_flag"

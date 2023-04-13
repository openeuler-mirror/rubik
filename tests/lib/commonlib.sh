#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2021-2023. All rights reserved.
# rubik licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Create: 2021-05-15
# Description: common lib for integration test

top_dir=$(git rev-parse --show-toplevel)
exit_flag=0

function pre_fun() {
    if pgrep rubik > /dev/null 2>&1; then
        echo "rubik is already running, please stop it first"
        exit 1
    fi
    if [ ! -f "${top_dir}"/rubik ]; then
        pushd "${top_dir}" || exit 1 > /dev/null 2>&1
        make release
        popd || exit 1 > /dev/null 2>&1
    fi
    nohup "${top_dir}"/rubik > /tmp/rubik_log 2>&1 &
    rubik_pid=$!

    # check if rubik is started
    rubik_started=0
    for _ in $(seq 1 30); do
        if ! grep -i "start http server" /tmp/rubik_log > /dev/null 2>&1; then
            sleep 0.1
            continue
        else
            rubik_started=1
            break
        fi
    done
    if [ "${rubik_started}" -eq 0 ]; then
        echo "rubik start failed, log dir /tmp/rubik_log"
    fi
}

function post_fun() {
    if [ -n "${rubik_pid}" ]; then
        kill -15 "${rubik_pid}" > /dev/null 2>&1
    fi
}

function clean_all() {
    rm -rf /tmp/rubik_log
}

function generate_config_file() {
    sed -n '/config.json:/{:a;n;/---/q;p;ba}' "${top_dir}"/hack/rubik-daemonset.yaml > /var/lib/rubik/config.json
}

function env_check() {
    ls /sys/fs/cgroup/cpu/cpu.qos_level > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        return 1
    fi

    if [ ! -f /proc/sys/vm/memcg_qos_enable ]; then
        return 1
    else
        echo -n 1 > /proc/sys/vm/memcg_qos_enable
    fi

    ls /sys/fs/cgroup/memory/memory.qos_level > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        return 1
    fi

    ls /sys/fs/resctrl/schemata > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        return 1
    fi
}

#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2020-2023. All rights reserved.
# rubik licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Create: 2020-08-27
# Description: common functions for fuzz tests

# Description: check the log and return the result
#              if crash, return 1
#              if not, return 0
# Usage: check_result /path/to/log
# $1: the full path of log
function check_result() {
    local log=$1
    local time="[$(date '+%Y-%m-%d %H:%M:%S')]"
    result=$(grep "crash" "$log" | tail -1 | awk '{print $10}')
    result=${result%%,}
    if [[ $result -eq 0 ]]; then
        echo "PASS$time: No crash found"
        return 0
    else
        echo "FAIL$time: Crash found, See details in $log"
        return 1
    fi
}

# Description: sleep x s/m/h and kill the process
# Usage: kill_after 1h 10232
# Input: $1: time to sleep
#        $2: pid to kill
function kill_after() {
    local time_out=$1
    local pid_to_kill=$2
    sleep "$time_out"
    for j in $(seq 1 100); do
        kill -9 "$pid_to_kill" > /dev/null 2>&1
        if pgrep -a "$pid_to_kill"; then
            sleep 0.2
        else
            break
        fi
        if [[ $j -eq 100 ]]; then
            return 1
        fi
    done
}

# Description: compile Fuzz.go
# Usage: make_fuzz_zip $fuzz_file $fuzz_dir $test_dir
# Input: $1: path to Fuzz.go file
#        $2: dir to put the Fuzz.go file
#        $3: dir store the build output
# Return: success 0; failed 1
# Warning: all input should be abs path :-)
function make_fuzz_zip() {
    # $top_dir/tests/data/fuzz-test/fuzz-xxx/Fuzz
    fuzz_file=$1
    # $top_dir/pkg/package_name(config)
    fuzz_dir=$2
    # $top_dir/tests/data/fuzz-test/fuzz-xxx/
    data_dir=$3
    cp "$fuzz_file" "$fuzz_dir"
    pushd "$fuzz_dir" > /dev/null 2>&1 || return 1
    mv Fuzz Fuzz.go
    if ! go-fuzz-build "$fuzz_dir"; then
        echo "go-fuzz-build failed" && return 1
    fi
    mv "$fuzz_dir"/*.zip "$data_dir"
    rm "$fuzz_dir/Fuzz.go"
    popd > /dev/null 2>&1 || return 1
}

# Description: set environment for go fuzz test
# Usage: set_env "fuzz-test-abc" $top_dir
# Input: $1: test name
#        $2: abs path for rubik project
# Note: 1. test_name must start with fuzz-test, for example fuzz-test-abc
#       2. go fuzz file must have name "Fuzz.go"
#       3. top_dir must be the abs path for the rubik project
# shellcheck disable=SC2034
function set_env() {
    test_name=$1
    top_dir=$2

    test_root=$top_dir/tests/data/fuzz-test
    test_dir=$test_root/$test_name
    fuzz_file=$test_dir/Fuzz
    fuzz_dir="$top_dir"/"$(cat "$test_dir"/path)"
    fuzz_corpus="$test_dir/corpus"
    fuzz_log="$test_dir/$test_name.log"
    fuzz_crashers="$test_dir/crashers"
    fuzz_suppressions="$test_dir/suppressions"
    fuzz_zip=""
}

function clean_env() {
    rm -rf "$fuzz_zip" "$fuzz_crashers" "$fuzz_suppressions"
    find /tmp -maxdepth 1 -iname "*fuzz*" -exec rm -rf {} \;
}

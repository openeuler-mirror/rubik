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
# Create: 2021-04-15
# Description: DT test script

top_dir=$(git rev-parse --show-toplevel)

# go fuzz test
function fuzz() {
    failed=0
    while IFS= read -r testfile; do
        printf "%-45s" "test $(basename "$testfile"): " | tee -a ${top_dir}/tests/fuzz.log
        bash "$testfile" "$1" | tee -a ${top_dir}/tests/fuzz.log
        if [ $PIPESTATUS -ne 0 ]; then
            failed=1
        fi
        # delete tmp files to avoid "no space left" problem
        find /tmp -maxdepth 1 -iname "*fuzz*" -exec rm -rf {} \;
    done < <(find "$top_dir"/tests/src -maxdepth 1 -name "fuzz_*.sh" -type f -print | sort)
    exit $failed
}

# integration test
function normal() {
    source "${top_dir}"/tests/lib/commonlib.sh
    failed=0
    while IFS= read -r testfile; do
        printf "%-45s" "$(basename "$testfile"): "
        if ! bash "$testfile"; then
            failed=1
        fi
    done < <(find "$top_dir"/tests/src -maxdepth 1 -name "test_*" -type f -print | sort)
    if [[ ${failed} -ne 0 ]]; then
        exit $failed
    else
        clean_all
    fi
}

# main function to chose which kind of test
function main() {
    case "$1" in
        fuzz)
            fuzz "$2"
            ;;
        *)
            normal
            ;;
    esac
}

main "$@"

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
# Create: 2021-05-30
# Description: rubik flag check 0001

top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh

test_fun() {
    # check rubik binary
    if [ ! -f "${top_dir}"/rubik ]; then
        pushd "${top_dir}" || exit 1 > /dev/null 2>&1
        make release
        popd || exit 1 > /dev/null 2>&1
    fi

    # check rubik flag
    if "${top_dir}"/rubik -v > /dev/null 2>&1; then
        if ! "${top_dir}"/rubik --help > /dev/null 2>&1 && ! "${top_dir}"/rubik -h > /dev/null 2>&1; then
            echo "PASS"
        else
            echo "FAILED"
        fi
    else
        echo "FAILED"
    fi
}

test_fun

exit "$exit_flag"

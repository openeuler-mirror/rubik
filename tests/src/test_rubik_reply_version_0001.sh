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
# Create: 2021-05-15
# Description: rubik reply healthcheck 0001

set -x
top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh

pre_fun() {
    prepare_rubik
    run_rubik
}

test_fun() {
    result=$(rubik_version)
    if [[ $? -eq 0 ]]; then
        field_number=$(echo "${result}" | grep -iE "version|release|commit|buildtime" -o | wc -l)
        if [[ $field_number -eq 4 ]]; then
            echo "PASS"
        else
            echo "FAILED"
            ((exit_flag++))
        fi
    else
        echo "FAILED"
        ((exit_flag++))
    fi
}

post_fun() {
    clean_all
    exit "$exit_flag"
}

pre_fun
test_fun
post_fun

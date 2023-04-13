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
# Create: 2021-05-21
# Description: release check

top_dir=$(git rev-parse --show-toplevel)
source "$top_dir"/tests/lib/commonlib.sh
version_file="${top_dir}/VERSION"
spec_file="${top_dir}/rubik.spec"

test_fun() {
    [[ -z $spec_file ]] || echo "PASS" & exit 0
    spec_release=$(head -10 "$spec_file" | grep Release | awk '{print $NF}')
    VERSION_release=$(awk -F"-" '{print $2}' < "${version_file}")
    if [ "$spec_release" != "$VERSION_release" ]; then
        echo "FAILED: release in spec not consistent with release in VERSION"
        ((exit_flag++))
    else
        echo "PASS"
    fi
}

test_fun

exit "$exit_flag"

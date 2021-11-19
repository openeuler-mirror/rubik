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
# Description: go test script

export GO111MODULE=off

test_log=${PWD}/unit_test_log
rm -rf "${test_log}"
touch "${test_log}"
go_list=$(go list ./...)
for path in ${go_list}; do
    echo "Start to test: ${path}"
    go test -race -cover -count=1 -timeout 300s -v "${path}" >> "${test_log}"
    cat "${test_log}" | grep -E -- "--- FAIL:|^FAIL"
    if [ $? -eq 0 ]; then
        echo "Testing failed... Please check ${test_log}"
        exit 1
    fi
    tail -n 1 "${test_log}"
done

rm -rf "${test_log}"

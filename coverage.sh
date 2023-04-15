#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
# rubik licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Create: 2023-04-15
# Description: go test coverage script

modulename="rubik"
testcasesresultfile="testcasesresult.json"
coveragedir="coverage/"
profile="coverage.out"
funcfile=$coveragedir"coverage.func"
htmlfile=$coveragedir"coverage.html"
coverageresultfile=$coveragedir"coverage.json"
testlogfile="unit_test_log"
gcovfile="gocov.json"

function pre() {
    echo -e "\033[32mThis scripts depends on several commands: jq, gcov, gcov-html\033[0m"
    if command -v git > /dev/null 2>&1; then
        return
    else
        yum install jq
    fi
}

function count_testcases() {
    make tests
    if [[ $? -ne 0 ]]; then
        cat $testlogfile
        exit 1
    fi
    stotal=$(grep -rn "^func Test.*" | grep "^pkg/" | wc -l)
    run=$(grep -rn -- "^=== RUN" $testlogfile | grep -v "\/" | wc -l)
    succeeded=$(grep -rn -- "--- PASS" $testlogfile | grep -v "\/" | wc -l)
    failed=$(grep -rni -- "--- FAIL" $testlogfile | grep -v "\/" | wc -l)
    skip=$(grep -rn -- "--- SKIP" $testlogfile | grep -v "\/" | wc -l)

    jq -n -c -M --arg modulename "${modulename}" --arg stotal "${stotal}" --arg run "${run}" \
        --arg succeeded "${succeeded}" --arg failed "${failed}" --arg skip "${skip}" \
        '[{"subModule":$modulename,"data":[{"type":"Test Cases","stotal":$stotal,"run":$run,
        "succeeded":$succeeded,"failed":$failed,"inactive":$skip}]}]' \
        > $testcasesresultfile
}

function count_line() {
    mkdir -p coverage
    go test -coverprofile=$profile ./...
    if [[ $? -ne 0 ]]; then
        exit 1
    fi

    gocov > /dev/null 2>&1
    if [[ $? -eq 2 ]]; then
        export GOROOT=/usr/local/go
        gocov convert $profile > $gcovfile
        gocov-html $gcovfile > ${htmlfile}
        linesData=$(awk '/percent/' ${htmlfile} | tail -1 | awk -F '<code>' '{print$4}' | awk -F '</code>' '{print$1}')
        linesHit=$(echo ${linesData} | awk -F '/' '{print$1}')
        linesTotal=$(echo ${linesData} | awk -F '/' '{print$2}')
        linesUrl="rubik/${htmlfile}"
    else
        echo "should install gcov first"
    fi

    go tool cover -func=$profile > $funcfile
    lineCoverage=$(grep "total" $funcfile | awk '{print $3}')

    jq -n -c -M --arg lineCoverage "${lineCoverage}" --arg linesHit "${linesHit}" \
        --arg linesTotal "${linesTotal}" --arg linesUrl "${linesUrl}" \
        '{"lines":$lineCoverage, "linesHit":$linesHit, "linesTotal":$linesTotal, "linesUrl":$linesUrl}' \
        > $coverageresultfile
}

function clean_env() {
    rm -rf $coveragedir
    rm -rf $profile
    rm -rf $testcasesresultfile
    rm -rf $testlogfile
    rm -rf $gcovfile
}

function show_results() {
    cat $testcasesresultfile | jq
    cat $coverageresultfile | jq
}

function cover() {
    pre
    count_testcases
    count_line
}

# main function to chose which kind of test
function main() {
    case "$1" in
        clean)
            clean_env
            ;;
        show)
            pre
            show_results
            ;;
        *)
            cover
            ;;
    esac
}

main "$@"

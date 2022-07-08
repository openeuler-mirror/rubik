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
# Description: common lib for integration test

TOP_DIR=$(git rev-parse --show-toplevel)
RUBIK_TEST_ROOT="${TOP_DIR}"/.fortest
RUBIK_LIB=${RUBIK_TEST_ROOT}/rubik_lib
RUBIK_RUN=${RUBIK_TEST_ROOT}/rubik_run
RUBIK_LOG=${RUBIK_TEST_ROOT}/rubik_log
RUBIK_CONFIG="${RUBIK_LIB}"/config.json
RUBIK_NAME="rubik-agent"
QOS_HANDLER="http://localhost"
VERSION_HANDLER="http://localhost/version"
PING_HANDLER="http://localhost/ping"
PAUSE_IMAGE="k8s.gcr.io/pause:3.2"
OPENEULER_IMAGE="openeuler-22.03-lts:latest"
SKIP_FLAG=111

mkdir -p "${RUBIK_TEST_ROOT}"
mkdir -p "${RUBIK_LIB}" "${RUBIK_RUN}" "${RUBIK_LOG}"
exit_flag=0

## Description: build_rubik_img will build rubik image
# Usage: build_rubik_img
# Input: $1: baseimage default is scratch
#        $2: images tag default is rubik_version
# Output: rubik image
# Example: build_rubik_img
function build_rubik_img() {
    image_base=${1:-"scratch"}
    image_tag=${2:-"fortest"}
    rubik_img="rubik:$image_tag"
    if [ "$image_base" != "scratch" ]; then
        cp "${RUBIK_LIB}"/Dockerfile "${RUBIK_LIB}"/Dockerfilebak
        sed -i "s#scratch#${image_base}#g" "${RUBIK_LIB}"/Dockerfile
    fi
    if [ ! -f "${TOP_DIR}"/rubik ]; then
        make release
    fi
    cp "$TOP_DIR"/rubik "$RUBIK_LIB"
    docker images | grep ^rubik | grep "${image_tag}" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        docker build -f "${RUBIK_LIB}"/Dockerfile -t "${rubik_img}" "${RUBIK_LIB}"
        [ "$image_base" != "scratch" ] && rm "${RUBIK_LIB}"/Dockerfile && mv "${RUBIK_LIB}"/Dockerfilebak "${RUBIK_LIB}"/Dockerfile
    fi
}

function generate_config_file() {
    # get config from yaml config map
    sed -n '/config.json:/{:a;n;/---/q;p;ba}' "$RUBIK_LIB"/rubik-daemonset.yaml > "${RUBIK_CONFIG}"
    # disable autoConfig
    sed -i 's/\"autoConfig\": true/\"autoConfig\": false/' "${RUBIK_CONFIG}"
}

function prepare_rubik() {
    runtime_check
    if pgrep rubik > /dev/null 2>&1; then
        echo "rubik is already running, please stop it first"
        exit 1
    fi
    cp "$TOP_DIR"/Dockerfile "$TOP_DIR"/hack/rubik-daemonset.yaml "${RUBIK_LIB}"
    image_base=${1:-"scratch"}
    image_tag=${2:-"fortest"}
    rubik_img="rubik:$image_tag"
    build_rubik_img "${image_base}" "${image_tag}"
    generate_config_file
}

function run_rubik() {
    prepare_rubik
    image_check
    if [ ! -f "${RUBIK_CONFIG}" ]; then
        rubik_pid=$(docker run -tid --name=${RUBIK_NAME} --pid=host --cap-add SYS_ADMIN \
            -v "${RUBIK_RUN}":/run/rubik -v "${RUBIK_LOG}":/var/log/rubik -v /sys/fs:/sys/fs "${rubik_img}")
    else
        rubik_pid=$(docker run -tid --name=${RUBIK_NAME} --pid=host --cap-add SYS_ADMIN \
            -v "${RUBIK_RUN}":/run/rubik -v "${RUBIK_LOG}":/var/log/rubik -v /sys/fs:/sys/fs \
            -v "${RUBIK_CONFIG}":/var/lib/rubik/config.json "${rubik_img}")
    fi
    return_code=$?
    echo -n "$rubik_pid"
    return $return_code
}

function kill_rubik() {
    docker inspect ${RUBIK_NAME} > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        docker logs ${RUBIK_NAME}
        docker stop -t 0 ${RUBIK_NAME}
        docker rm ${RUBIK_NAME}
    fi
}

function clean_all() {
    rm -rf "${RUBIK_LIB}" "${RUBIK_RUN}" "${RUBIK_LOG}"
    kill_rubik
}

function runtime_check() {
    if ! docker info > /dev/null 2>&1; then
        echo "docker is not found, please install it via 'yum install docker'"
        exit 1
    fi
}

function image_check() {
    openEuler_image="openeuler-22.03-lts"
    pause_image="k8s.gcr.io/pause"
    if ! docker images | grep $openEuler_image > /dev/null 2>&1; then
        echo "openEuler image ${OPENEULER_IMAGE} is not found, please prepare it first before test begin"
        exit 1
    fi
    if ! docker images | grep ${pause_image} > /dev/null 2>&1; then
        echo "pause image ${PAUSE_IMAGE} is not found, please prepare it first before test begin"
        exit 1
    fi
}

# Description: kernel_check will check wether the environment is met for testing
# Usage: kernel_check [...]
# Input: list of functional check requirements, default check all
# Output: 0(success) 1(fail)
# Example: kernel_check CPU、kernel_check CPU MEM CACHE、kernel_check ALL
function kernel_check() {
    function cpu_check() {
        ls /sys/fs/cgroup/cpu/cpu.qos_level > /dev/null 2>&1
        if [ $? -ne 0 ]; then
            echo "ls /sys/fs/cgroup/cpu/cpu.qos.level failed"
            return 1
        fi
    }
    function mem_check() {
        if [ ! -f /proc/sys/vm/memcg_qos_enable ]; then
            echo "/proc/sys/vm/memcg_qos_enable is not an ordinary file"
            return 1
        else
            echo -n 1 > /proc/sys/vm/memcg_qos_enable
        fi
        ls /sys/fs/cgroup/memory/memory.qos_level > /dev/null 2>&1
        if [ $? -ne 0 ]; then
            echo "ls /sys/fs/cgroup/memory.qos_level failed"
            return 1
        fi
    }
    function cache_check() {
        ls /sys/fs/resctrl/schemata > /dev/null 2>&1
        if [ $? -ne 0 ]; then
            echo "ls /sys/fs/resctrl/schemata failed"
            return 1
        fi
    }
    function check_all() {
        cpu_check
        mem_check
        cache_check
    }
    for functional in $@; do
        case $functional in
            "CPU")
                cpu_check
                ;;
            "MEM")
                mem_check
                ;;
            "CACHE")
                cache_check
                ;;
            "ALL" | *)
                check_all
                ;;
        esac
    done
}

# Description: curl_cmd performs like curl command
# Usage: curl_cmd $http_handler $json_data $protocol
# Reminds: NOT recommend to use this method directly, use rubik_ping/rubik_version/rubik_qos instead in most occasions
# Input:
#   $1: http_handler
#   $2: json_data
#   $3: protocol
# Output: curl command execute return message
# Return: success(0) or fail(not 0)
# Example:
#   data=$(gen_pods_json json1 json2)
#   QOS_HANDLER="http://localhost"
#   PING_HANDLER="http://localhost/ping"
#   VERSION_HANDLER="http://localhost/version"
#   protocol="GET"
#
#   curl_cmd $QOS_HANDLER $data $protocol
#   curl_cmd $PING_HANDLER $protocol
#   curl_cmd $VERSION_HANDLER $protocol
function curl_cmd() {
    http_handler=$1
    data=$2
    protocol=$3
    result=$(curl -s -H "Accept: application/json" -H "Content-type: application/json" -X "$protocol" --data "$(echo -n "$data")" --unix-socket "${RUBIK_RUN}"/rubik.sock "$http_handler")
    return_code=$?
    echo "$result"
    return $return_code
}

function rubik_ping() {
    curl_cmd "$PING_HANDLER" "" "GET"
}

function rubik_version() {
    curl_cmd "$VERSION_HANDLER" "" "GET"
}

function rubik_qos() {
    curl_cmd "$QOS_HANDLER" "$1" "POST"
}

# Description: gen_single_pod_json will generate single pod qos info for one pod
# Usage: gen_single_pod_json $pod_id $cgroup_path $qos_level
# Input: $1: pod id, $2: cgroup path, $3: qos level
# Output: single pod qos info json data
# Example: json1=$(gen_single_pod_json "podaaaaaa" "this/is/cgroup/path" 1)
function gen_single_pod_json() {
    pod_id=$1
    cgroup_path=$2
    qos_level=$3
    jq -n -c -r --arg pid "$pod_id" --arg cp "$cgroup_path" --arg qos "$qos_level" '{"Pods":{($pid): {"CgroupPath": $cp, "QoSLevel": ($qos|tonumber)}}}'
}

function fn_check_result() {
    if [ "$1" = "$2" ]; then
        echo "PASS"
    else
        echo "FAIL"
        ((exit_flag++))
    fi
}

function fn_check_result_noeq() {
    if [ "$1" != "$2" ]; then
        echo "PASS"
    else
        echo "FAIL"
        ((exit_flag++))
    fi
}

function fn_check_string_contain() {
    if echo "$2" | grep "$1"; then
        echo "PASS"
    else
        echo "FAIL"
        ((exit_flag++))
    fi
}

function fn_check_string_not_contain() {
    if ! echo "$2" | grep "$1"; then
        echo "PASS"
    else
        echo "FAIL"
        ((exit_flag++))
    fi
}

# Description: long_char will generate long string by repeat char 'a' N times
# Usage: long_char $length
# Input: $1: length of string
# Output: repeate string with given length
# Example: long_char 10
function long_char() {
    length=$1
    head -c "$length" < /dev/zero | tr '\0' '\141'
}

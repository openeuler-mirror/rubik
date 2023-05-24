#!/bin/bash
###################################################################################################
# Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
# rubik licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Author: Xiang Li
# Create: 2022-11-29
# Description: Build container image for rubik. Enjoy and cherrs!
# Steps:
#    1. build image and tag it with rubik version and release number
#    2. modify `rubik-daemonset.yaml` file
###################################################################################################
set -e

CURRENT_DIR=$(cd "$(dirname "$0")" && pwd)
BINARY_NAME="rubik"

RUBIK_FILE="${CURRENT_DIR}/build/rubik"
DOCKERFILE="${CURRENT_DIR}/Dockerfile"
YAML_FILE="${CURRENT_DIR}/rubik-daemonset.yaml"

# Get version and release number of rubik binary
VERSION=$(${RUBIK_FILE} -v | grep ^Version | awk '{print $NF}')
RELEASE=$(${RUBIK_FILE} -v | grep ^Release | awk '{print $NF}')
IMG_TAG="${VERSION}-${RELEASE}"

# Get rubik image name and tag
IMG_NAME_AND_TAG="${BINARY_NAME}:${IMG_TAG}"

# Build container image for rubik
docker build -f "${DOCKERFILE}" -t "${IMG_NAME_AND_TAG}" "${CURRENT_DIR}"

echo -e "\n"
# Check image existence
docker images | grep -E "REPOSITORY|${BINARY_NAME}"

# Modify rubik-daemonset.yaml file, set rubik image name
sed -i "/image:/s/:.*/: ${IMG_NAME_AND_TAG}/" "${YAML_FILE}"

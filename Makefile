# Copyright (c) Huawei Technologies Co., Ltd. 2020. All rights reserved.
# rubik licensed under the Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#     http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
# PURPOSE.
# See the Mulan PSL v2 for more details.
# Author: Xiang Li
# Create: 2021-04-17
# Description: Makefile for rubik

CWD=$(realpath .)
TMP_DIR := /tmp/rubik_tmpdir

DEBUG_FLAGS := -gcflags="all=-N -l"
LD_FLAGS := -ldflags '-buildid=none -tmpdir=$(TMP_DIR) -extldflags=-ftrapv -extldflags=-Wl,-z,relro,-z,now -linkmode=external -extldflags=-static'

export GO111MODULE=off

GO_BUILD=CGO_ENABLED=1 \
	CGO_CFLAGS="-fstack-protector-strong -fPIE" \
	CGO_CPPFLAGS="-fstack-protector-strong -fPIE" \
	CGO_LDFLAGS_ALLOW='-Wl,-z,relro,-z,now' \
	CGO_LDFLAGS="-Wl,-z,relro,-z,now -Wl,-z,noexecstack" \
	go build -buildmode=pie

all: release

dev:
	$(GO_BUILD) $(DEBUG_FLAGS) -o rubik $(LD_FLAGS) rubik.go

release:
	rm -rf $(TMP_DIR) && mkdir -p $(ORG_PATH) $(TMP_DIR)
	$(GO_BUILD) -o rubik $(LD_FLAGS) rubik.go 2>/dev/null

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
INSTALL_DIR := /var/lib/rubik
VERSION_FILE := ./VERSION
TEST_FILE := ./TEST
VERSION := $(shell cat $(VERSION_FILE) | awk -F"-" '{print $$1}')
RELEASE := $(shell cat $(VERSION_FILE) | awk -F"-" '{print $$2}')
BUILD_TIME := $(shell date "+%Y-%m-%d")
USAGE := $(shell [ -f $(TEST_FILE) ] && echo 'TestOnly')
GIT_COMMIT := $(if $(shell git rev-parse --short HEAD),$(shell git rev-parse --short HEAD),$(shell cat ./git-commit | head -c 7))

DEBUG_FLAGS := -gcflags="all=-N -l"
LD_FLAGS := -ldflags '-buildid=none -tmpdir=$(TMP_DIR) \
	-X isula.org/rubik/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X isula.org/rubik/pkg/version.BuildTime=$(BUILD_TIME) \
	-X isula.org/rubik/pkg/version.Version=$(VERSION) \
	-X isula.org/rubik/pkg/version.Release=$(RELEASE) \
	-X isula.org/rubik/pkg/version.Usage=$(USAGE) \
	-extldflags=-ftrapv \
	-extldflags=-Wl,-z,relro,-z,now -linkmode=external -extldflags=-static'

export GO111MODULE=off

GO_BUILD=CGO_ENABLED=1 \
	CGO_CFLAGS="-fstack-protector-strong -fPIE" \
	CGO_CPPFLAGS="-fstack-protector-strong -fPIE" \
	CGO_LDFLAGS_ALLOW='-Wl,-z,relro,-z,now' \
	CGO_LDFLAGS="-Wl,-z,relro,-z,now -Wl,-z,noexecstack" \
	go build -buildmode=pie

all: release

help:
	@echo "Usage:"
	@echo
	@echo "make                          # build rubik for debug"
	@echo "make release                  # build rubik for release, open security build option"
	@echo "make image                    # container image build"
	@echo "make check                    # static check for latest commit"
	@echo "make checkall                 # static check for whole project"
	@echo "make tests                    # run all testcases within project"
	@echo "make test-unit                # only run unit test for project"
	@echo "make cover                    # generate cover report"
	@echo

dev:
	$(GO_BUILD) $(DEBUG_FLAGS) -o rubik $(LD_FLAGS) rubik.go

release:
	rm -rf $(TMP_DIR) && mkdir -p $(ORG_PATH) $(TMP_DIR)
	$(GO_BUILD) -o rubik $(LD_FLAGS) rubik.go 2>/dev/null
	@if [ -f ./hack/rubik-daemonset.yaml ]; then sed -i 's/rubik_image_name_and_tag/rubik:$(VERSION)-$(RELEASE)/g' ./hack/rubik-daemonset.yaml; fi;

safe: release

image: release
	docker build -f Dockerfile -t rubik:$(VERSION)-$(RELEASE) .

check:
	@echo "Static check start for last commit"
	@./hack/static_check.sh last
	@echo "Static check last commit finished"

checkall:
	@echo "Static check start for whole project"
	@./hack/static_check.sh all
	@echo "Static check project finished"

tests: test-unit test-integration

test-unit:
	@bash ./hack/unit_test.sh

test-integration:
	@bash ./tests/test.sh

cover:
	go test -p 1 -v ./... -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html
	python3 -m http.server 8080

install:
	install -d -m 0750 $(INSTALL_DIR)
	install -Dp -m 0550 ./rubik $(INSTALL_DIR)
	install -Dp -m 0640 ./hack/rubik-daemonset.yaml $(INSTALL_DIR)
	install -Dp -m 0640 ./hack/cluster-role-binding.yaml $(INSTALL_DIR)
	install -Dp -m 0640 ./Dockerfile $(INSTALL_DIR)

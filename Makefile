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
BUILD_DIR=build
VERSION_FILE := ./VERSION
VERSION := $(shell awk -F"-" '{print $$1}' < $(VERSION_FILE))
RELEASE :=$(if $(shell awk -F"-" '{print $$2}' < $(VERSION_FILE)),$(shell awk -F"-" '{print $$2}' < $(VERSION_FILE)),NA)
BUILD_TIME := $(shell date "+%Y-%m-%d")
GIT_COMMIT := $(if $(shell git rev-parse --short HEAD),$(shell git rev-parse --short HEAD),$(shell cat ./git-commit | head -c 7))

export GO111MODULE=on

DEBUG_FLAGS := -gcflags="all=-N -l"
EXTRALDFLAGS := -buildmode=pie -extldflags=-ftrapv \
	-extldflags=-Wl,-z,relro,-z,now -linkmode=external -extldflags "-static-pie -Wl,-z,now"

LD_FLAGS := -ldflags '-buildid=none -tmpdir=$(TMP_DIR) \
	-X isula.org/rubik/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X isula.org/rubik/pkg/version.BuildTime=$(BUILD_TIME) \
	-X isula.org/rubik/pkg/version.Version=$(VERSION) \
	-X isula.org/rubik/pkg/version.Release=$(RELEASE) \
	$(EXTRALDFLAGS)'

GO_BUILD=CGO_ENABLED=1 \
	CGO_CFLAGS="-fstack-protector-strong -fPIE" \
	CGO_CPPFLAGS="-fstack-protector-strong -fPIE" \
	CGO_LDFLAGS_ALLOW='-Wl,-z,relro,-z,now' \
	CGO_LDFLAGS="-Wl,-z,relro,-z,now -Wl,-z,noexecstack" \
	go build -mod=vendor

all: release

help:
	@echo "Usage:"
	@echo
	@echo "make                          # build rubik with security build option"
	@echo "make image                    # build container image"
	@echo "make check                    # static check for latest commit"
	@echo "make checkall                 # static check for whole project"
	@echo "make tests                    # run all tests"
	@echo "make test-unit                # run unit test"
	@echo "make cover                    # generate coverage report"
	@echo "make install                  # install files to /var/lib/rubik"

prepare:
	mkdir -p $(TMP_DIR) $(BUILD_DIR)
	rm -rf $(TMP_DIR) && mkdir -p $(TMP_DIR)

release: prepare
	$(GO_BUILD) -o $(BUILD_DIR)/rubik $(LD_FLAGS) rubik.go
	sed "/image:/s/:.*/: rubik:$(VERSION)-$(RELEASE)/" hack/rubik-daemonset.yaml > $(BUILD_DIR)/rubik-daemonset.yaml
	cp hack/rubik.service $(BUILD_DIR)

debug: prepare
	EXTRALDFLAGS=""
	go build $(LD_FLAGS) $(DEBUG_FLAGS) -o $(BUILD_DIR)/rubik rubik.go
	sed "/image:/s/:.*/: rubik:$(VERSION)-$(RELEASE)/" hack/rubik-daemonset.yaml > $(BUILD_DIR)/rubik-daemonset.yaml
	cp hack/rubik.service $(BUILD_DIR)

image: release
	docker build -f Dockerfile -t rubik:$(VERSION)-$(RELEASE) .

check:
	@echo "Static check for last commit ..."
	@./hack/static_check.sh last
	@echo "Static check for last commit finished"

checkall:
	@echo "Static check for all ..."
	@./hack/static_check.sh all
	@echo "Static check for all finished"

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
	cp -f $(BUILD_DIR)/* $(INSTALL_DIR)
	cp -f $(BUILD_DIR)/rubik.service /lib/systemd/system/


#/*
#Copyright © 2019 Intel Corporation
#SPDX-License-Identifier: BSD-3-Clause
#*/

# ISecL K8S Extensions Extended Scheduler
# Works in tandem with K8s scheduler to return a filtered list of nodes as per predicates on CRDs

GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
GITCOMMIT := $(shell git describe --always)
VERSION := $(or ${GITTAG}, v0.0.0)
DESCRIPTION="ISecL K8S Extended Scheduler"
GOPATH := $(shell go env GOPATH)
PROXY_EXISTS := $(shell if [[ "${https_proxy}" || "${http_proxy}" ]]; then echo 1; else echo 0; fi)

SERVICE=isecl-k8s-scheduler

.PHONY: $(SERVICE), installer, all, clean
# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE):
	env GOSUMDB=off GOPROXY=direct go mod download
	@rm -f ${GOPATH}/pkg/mod/k8s.io/kubernetes*/plugin/pkg/scheduler/api/zz_generated.deepcopy.go
	env GOOS=linux GOSUMDB=off GOPROXY=direct go build -ldflags "-X intel/isecl/$(SERVICE)/version.Version=$(VERSION) -X intel/isecl/$(SERVICE)/version.GitHash=$(GITCOMMIT)" -o out/$(SERVICE)

.PHONY: docker

docker: $(SERVICE)
	mkdir -p out/ 
ifeq ($(PROXY_EXISTS),1)
	docker build --no-cache -t isecl/k8s-scheduler:$(VERSION) --build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy} -f ./dist/docker/Dockerfile ./out
else
	docker build --no-cache -t isecl/k8s-scheduler:$(VERSION) --network host -f ./dist/docker/Dockerfile ./out
endif
	docker save isecl/k8s-scheduler:$(VERSION) > ./out/docker-isecl-scheduler-$(VERSION).tar

all: clean docker
# Removes the generated service config and binary files
.PHONY: clean
clean:
	rm -rf out/

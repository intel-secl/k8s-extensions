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

SERVICE=isecl-k8s-scheduler

.PHONY: $(SERVICE), installer, all, clean
# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE):
	go mod download
	@rm -f ${GOPATH}/pkg/mod/k8s.io/kubernetes*/plugin/pkg/scheduler/api/zz_generated.deepcopy.go
	env GOOS=linux GOSUMDB=off GOPROXY=direct go build -ldflags "-X intel/isecl/$(SERVICE)/version.Version=$(VERSION) -X intel/isecl/$(SERVICE)/version.GitHash=$(GITCOMMIT)" -o out/$(SERVICE)

# Bundles the service config and the service binary files
.PHONY: installer
installer: $(SERVICE)
	mkdir -p out/installer
	cp dist/linux/install.sh out/installer/ && chmod +x out/installer/install.sh
	cp dist/linux/isecl-extended-scheduler-config.json out/installer/
	cp dist/linux/create_k8s_extsched_cert.sh out/installer/ && chmod +x out/installer/create_k8s_extsched_cert.sh
	cp out/$(SERVICE) out/installer
	cp dist/linux/$(SERVICE).service out/installer/
	cp dist/linux/scheduler-policy.json out/installer/
	makeself out/installer out/isecl-k8s-scheduler-$(VERSION).bin "ISecl K8s Scheduler $(VERSION)" ./install.sh

all: clean installer
# Removes the generated service config and binary files
.PHONY: clean
clean:
	rm -rf out/

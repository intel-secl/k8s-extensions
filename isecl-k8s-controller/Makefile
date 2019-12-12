#/*
#Copyright © 2019 Intel Corporation
#SPDX-License-Identifier: BSD-3-Clause
#*/

# ISecL K8S Custom Controller
# Applies labels as per Custom Resource Definitions

GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
GITCOMMIT := $(shell git describe --always)
VERSION := $(or ${GITTAG}, v0.0.0)
DESCRIPTION="ISecL K8S Custom Controller"

SERVICE=isecl-k8s-controller

.PHONY: $(SERVICE), installer, all, clean
# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE):
	env GOOS=linux go build -ldflags "-X intel/isecl/$(SERVICE)/version.Version=$(VERSION) -X intel/isecl/$(SERVICE)/version.GitHash=$(GITCOMMIT)" -o out/$(SERVICE)

# Bundles all the service config and the service binaries files
.PHONY: installer
installer: $(SERVICE)
	mkdir -p out/installer
	cp dist/linux/install.sh out/installer/ && chmod +x out/installer/install.sh
	cp out/$(SERVICE) out/installer/
	cp dist/linux/$(SERVICE).service out/installer/
	makeself out/installer out/isecl-k8s-controller-$(VERSION).bin "ISecl K8s controller $(VERSION)" ./install.sh

all: clean installer

# Removes the generated service config and binary files
clean:
	rm -rf out/

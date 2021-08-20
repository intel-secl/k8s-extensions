#/*
#Copyright © 2019 Intel Corporation
#SPDX-License-Identifier: BSD-3-Clause
#*/

# ISecL K8S Custom Controller
# Applies labels as per Custom Resource Definitions

SHELL:=/bin/bash
GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
GITCOMMIT := $(shell git describe --always)
VERSION := "v3.6.1"
DESCRIPTION="ISecL K8S Custom Controller"

SERVICE=isecl-k8s-controller

PROXY_EXISTS := $(shell if [[ "${https_proxy}" || "${http_proxy}" ]]; then echo 1; else echo 0; fi)
.PHONY: $(SERVICE), installer, all, clean
# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE):
	env GOOS=linux GOSUMDB=off GOPROXY=direct go build -ldflags "-X intel/isecl/$(SERVICE)/version.Version=$(VERSION) -X intel/isecl/$(SERVICE)/version.GitHash=$(GITCOMMIT)" -o out/$(SERVICE)

# Builds the docker image and creates the tar file from the image built
.PHONY: docker

docker: $(SERVICE)
	mkdir -p out/ 
ifeq ($(PROXY_EXISTS),1)
	docker build --no-cache -t isecl/k8s-controller:$(VERSION) --build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy} -f ./dist/docker/Dockerfile ./out
else
	docker build --no-cache -t isecl/k8s-controller:$(VERSION)  -f ./dist/docker/Dockerfile ./out
endif

oci-archive: docker
	skopeo copy docker-daemon:isecl/k8s-controller:$(VERSION) oci-archive:out/isecl-k8s-controller-$(VERSION)-$(GITCOMMIT).tar

all: clean oci-archive

# Removes the generated binary files and docker image as tar
clean:
	rm -rf out/

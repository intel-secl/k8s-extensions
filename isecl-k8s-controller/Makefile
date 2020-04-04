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
	env GOOS=linux GOSUMDB=off GOPROXY=direct go build -ldflags "-X intel/isecl/$(SERVICE)/version.Version=$(VERSION) -X intel/isecl/$(SERVICE)/version.GitHash=$(GITCOMMIT)" -o out/$(SERVICE)

# Builds the docker image and creates the tar file from the image built
.PHONY: docker

docker: $(SERVICE)
	mkdir -p out/ 
ifeq ($(PROXY_EXISTS),1)
	docker build --no-cache -t isecl/k8s-controller:latest --build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy} -f ./dist/docker/Dockerfile ./out
else
	docker build --no-cache -t isecl/k8s-controller:latest  -f ./dist/docker/Dockerfile ./out
endif
	docker save isecl/k8s-controller:latest > ./out/docker-isecl-controller-$(VERSION).tar
all: clean docker

# Removes the generated binary files and docker image as tar
clean:
	rm -rf out/

# ISecL K8S Extensions
DESCRIPTION="ISecL K8S Extensions"

GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
VERSION := $(or ${GITTAG}, v0.0.0)

.PHONY: scheduler, controller, all, clean

# Install the binary
installer:
	chmod +x build-k8s-extensions.sh
	./build-k8s-extensions.sh	 
	mkdir -p out/k8s-extensions/yamls
	cp -r certificate-generation-scripts/* out/k8s-extensions/
	cp isecl-k8s-scheduler/out/docker-isecl-scheduler-$(VERSION).tar out/k8s-extensions/
	cp isecl-k8s-controller/out/docker-isecl-controller-$(VERSION).tar out/k8s-extensions/
	cp -r isecl-k8s-controller/yamls/* out/k8s-extensions/yamls
	cp -r isecl-k8s-scheduler/yamls/* out/k8s-extensions/yamls
	cp -r isecl-k8s-scheduler/config-files/*  out/k8s-extensions/

all: clean installer

# Removes the generated service config and binary files
.PHONY: clean
clean:
	rm -rf out
	rm -rf isecl-k8s-scheduler/out/
	rm -rf isecl-k8s-controller/out/ 

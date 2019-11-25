# ISecL K8S Extensions
DESCRIPTION="ISecL K8S Extensions"

GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
GITCOMMIT := $(shell git describe --always)
GITCOMMITDATE := $(shell git log -1 --date=short --pretty=format:%cd)
VERSION := $(or ${GITTAG}, v0.0.0)

.PHONY: scheduler, controller, all, clean


# Install the binary
installer:
	chmod +x build-k8s-extensions.sh
	./build-k8s-extensions.sh	 
	mkdir -p out/k8s-extensions
	cp -r certificate-generation-scripts/* out/k8s-extensions/
	cp isecl-k8s-scheduler/out/isecl-k8s-scheduler-$(VERSION).bin out/k8s-extensions/
	cp isecl-k8s-controller/out/isecl-k8s-controller-$(VERSION).bin out/k8s-extensions/
	cp isecl-k8s-extensions.sh install.sh out/k8s-extensions/
	makeself out/k8s-extensions out/isecl-k8s-extensions-$(VERSION).bin "k8s extensions installer $(VERSION)" ./install.sh

all: clean installer

# Removes the generated service config and binary files
.PHONY: clean
clean:
	rm -rf out

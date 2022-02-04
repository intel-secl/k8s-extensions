# ISecL K8S Extensions
DESCRIPTION="ISecL K8S Extensions"

VERSION := "v4.0.2"

.PHONY: scheduler, controller, all, clean

# Install the binary
installer:
	chmod +x build-k8s-extensions.sh
	./build-k8s-extensions.sh	 
	mkdir -p out/isecl-k8s-extensions/yamls
	cp -r certificate-generation-scripts/* out/isecl-k8s-extensions/
	cp isecl-k8s-scheduler/out/*.tar out/isecl-k8s-extensions/
	cp isecl-k8s-controller/out/*.tar out/isecl-k8s-extensions/
	cp -r isecl-k8s-controller/yamls/* out/isecl-k8s-extensions/yamls
	cp -r isecl-k8s-scheduler/yamls/* out/isecl-k8s-extensions/yamls
	cp -r isecl-k8s-scheduler/config-files/*  out/isecl-k8s-extensions/
	cd out && tar -zcvf isecl-k8s-extensions-$(VERSION).tar.gz isecl-k8s-extensions

all: clean installer

# Removes the generated service config and binary files
.PHONY: clean
clean:
	rm -rf out
	rm -rf isecl-k8s-scheduler/out/
	rm -rf isecl-k8s-controller/out/ 

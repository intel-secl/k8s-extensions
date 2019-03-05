# ISecL K8S Extensions
DESCRIPTION="ISecL K8S Extensions"


VERSION := 1.0-SNAPSHOT
BUILD := `date +%FT%T%z`

# LDFLAGS
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.PHONY: installer, all, clean

# Install the binary
installer:
	mkdir -p out/k8s-extensions
	dist/linux/build-k8s-extensions.sh
	cp -r certificate-generation-scripts extended-scheduler custom-controller policy.json isecl-k8s-extensions.sh  out/k8s-extensions/
	makeself out/k8s-extensions out/isecl-k8s-extensions.bin "k8s extensions installer $(VERSION)" ./install.sh

all: installer

# Removes the generated service config and binary files
.PHONY: clean
clean:
	@rm -rf out

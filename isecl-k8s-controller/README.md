# ISecL Custom Controller
The ISecL Custom Controller creates/updates labels and annotation for K8s Worker Nodes whenever isecl.hostattributes crd is created or updated through K8s Kube Api Server.

## System Requirements
- RHEL 7.5/7.6
- Epel 7 Repo
- Proxy settings if applicable

## Software requirements
- git
- Go 11.4

# Step By Step Build Instructions

## Install required shell commands

### Install `go 1.11.4` or newer
The `ISecL Custom Controller` requires Go version 11.4 that has support for `go modules`. The build was validated with version 11.4 version of `go`. It is recommended that you use a newer version of `go` - but please keep in mind that the product has been validated with 1.11.4 and newer versions of `go` may introduce compatibility issues. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz
tar -xzf go1.11.4.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

## Build Custom Controller

- Git clone the Custom Controller
- Run scripts to build the Custom Controller

```shell
git clone https://github.com/intel-secl/k8s-custom-controller.git
cd k8s-custom-controller
make
```

# Installation of binary on kubernetes master machine

Pre-requisites
Kubernetes cluster should be up and running

STEPS:
1. Copy the complete source code to K8s master node and run below command for Installation
	```	
	make install
	```

2. Edit the isecl-k8s-controller.service file as below with tag_prefix.conf file path
	```
	vi /etc/systemd/system/isecl-k8s-controller.service

	ExecStart=/opt/isecl-k8s-extensions/bin/isecl-k8s-controller-1.0-SNAPSHOT -kubeconf=/etc/kubernetes/admin.conf -trustedprefixconf=<path>/tag_prefix.conf
	```

3. Run below commands to enable service daemon (to activate newly added service)
	```
	systemctl daemon-reload
	```

4. Run the custom controller using below command	

	```
	systemctl start isecl-k8s-controller.service
	```

5. To check status of this service run below command
	```
	systemctl status isecl-k8s-controller.service
	```
6. To stop this service run below command
	```
	systemctl stop isecl-k8s-controller.service
	```

# Links
https://01.org/intel-secl/

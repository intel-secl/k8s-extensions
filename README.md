# ISecL K8s Extenstions 

`ISecL K8s Extensions` which includes ISecL K8s extended scheduler, ISecL K8s custom controller components and certification generation scripts for trusted launch of containers.
Key Components:
- ISecL K8s extended scheduler
    The ISecL Extended Scheduler verifies trust report and asset tag signature for each of the K8s Worker Node annotation against Pod matching expressions in pod yaml file using ISecL Attestation hub public key.
    The signature verification ensures the integrity of labels created using isecl hostattribute crds on each of the worker nodes. The verification happens at the time of pod scheduling.
- ISecL K8s custom controller
    The ISecL Custom Controller creates/updates labels and annotation of K8s Worker Nodes whenever isecl.hostattributes crd objects are created or updated through K8s Kube Api Server.
- Certificate generation scripts
    These scripts creates kubernetes hostattributes.crd.isecl.intel.com from which the crd objects will be created for each of the tenant, then it creates the client and server certificates.
    The client certificate is created for root user and root user will be having RBAC on get,list,delete,patch,deletecollection,create and update operations on the hostattributes.crd.isecl.intel.com.

## System Requirements
- RHEL 8.1
- Epel 8 Repo
- Proxy settings if applicable

## Software requirements
- git
- makeself
- `go` version >= `go1.12.12` & <= `go1.13.8`

# Step By Step Build Instructions

## Install required shell commands

### Install tools from `yum`
```shell
sudo yum install -y git wget makeself
```

### Install `go` version >= `go1.12.12` & <= `go1.13.8`
The `ISecL K8s Extensions` requires Go version 1.12.12 that has support for `go modules`. The build was validated with the latest version go1.13.8 of `go`. It is recommended that you use go1.13.8 version of `go`. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.13.8.linux-amd64.tar.gz
tar -xzf go1.13.8.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export GOPATH=<path of project workspace>
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

## Build ISecL K8s Extenstions

```shell
git clone https://github.com/intel-secl/isecl-k8s-extensions.git
cd isecl-k8s-extensions
make all
```

### Deploy
```console
export MASTER_IP=`K8s Master IP`
> ./isecl-k8s-extensions-*.bin
```

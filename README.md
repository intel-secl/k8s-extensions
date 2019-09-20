# ISecL K8s Extenstions 

`ISecL K8s Extenstions` is binary installer package which includes K8s extended scheduler, K8s custom controller components and certification generation scripts

## System Requirements
- RHEL 7.5/7.6
- Epel 7 Repo
- Proxy settings if applicable

## Software requirements
- git
- makeself
- Go 11.4 or newer

# Step By Step Build Instructions

## Install required shell commands

### Install tools from `yum`
```shell
sudo yum install -y git wget makeself
```

### Install `go 1.11.4` or newer
The `ISecL K8s Extenstions` requires Go version 11.4 that has support for `go modules`. The build was validated with version 11.4 version of `go`. It is recommended that you use a newer version of `go` - but please keep in mind that the product has been validated with 1.11.4 and newer versions of `go` may introduce compatibility issues. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz
tar -xzf go1.11.4.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

## Build ISecL K8s Extenstions

```shell
git clone https://github.com/intel-secl/isecl-k8s-extensions.git
cd isecl-k8s-extensions
make
```

### Deploy
```console
> export MASTER_IP=<k8s-master-ip>
> ./isecl-k8s-extensions-*.bin
```

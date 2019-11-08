# ISecL K8s Extenstions 

`ISecL K8s Extenstions` is binary installer package which includes K8s extended scheduler, K8s custom controller components and certification generation scripts

## System Requirements
- RHEL 7.5/7.6
- Epel 7 Repo
- Proxy settings if applicable

## Software requirements
- git
- makeself
- `go` version >= `go1.11.4` & <= `go1.12.12`

# Step By Step Build Instructions

## Install required shell commands

### Install tools from `yum`
```shell
sudo yum install -y git wget makeself
```

### Install `go` version >= `go1.11.4` & <= `go1.12.12`
The `ISecL K8s Extensions` requires Go version 1.11.4 that has support for `go modules`. The build was validated with the latest version 1.12.12 of `go`. It is recommended that you use 1.12.12 version of `go`. More recent versions may introduce compatibility issues. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.12.12.linux-amd64.tar.gz
tar -xzf go1.12.12.linux-amd64.tar.gz
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

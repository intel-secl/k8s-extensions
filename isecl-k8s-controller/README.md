# ISecL Custom Controller
The ISecL Custom Controller creates/updates labels and annotation for K8s Worker Nodes whenever isecl.hostattributes crd is created or updated through K8s Kube Api Server.

## System Requirements
- RHEL 8.1
- Epel 8 Repo
- Proxy settings if applicable

## Software requirements
- git
- `go` version >= `go1.12.1` & <= `go1.14.1`

### Install tools from `yum`
```shell
sudo yum install -y git wget
```

### Install `go` version >= `go1.12.1` & <= `go1.14.1`
The `ISecL K8s Extensions` requires Go version 1.12.1 that has support for `go modules`. The build was validated with the latest version go1.14.1 of `go`. It is recommended that you use go1.14.1 version of `go`. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz
tar -xzf go1.14.1.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export GOPATH=<path of project workspace>
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

## Build Custom Controller

- Git clone the Custom Controller
- Run scripts to build the Custom Controller

```shell
git clone https://github.com/intel-secl/k8s-custom-controller.git
cd k8s-custom-controller
make all
```

## Deployment on Kubernetes master
The isecl-k8s-controller is deployed as a container using k8s deployments.

Pre-requisite 
Create HostAttributeCRD using yaml files located at k8s-custom-controller/yamls

```shell
kubectl apply -f crd-1.14.yml (k8s version < v1.16)
kubectl apply -f crd-1.17.yml (k8s version >= v1.16)
```

Load the controller image 
``` docker load -i docker-isecl-k8s-controller-v2.1.tar
```

Create Deployment.
kubectl apply -f isecl-controller.yml

# Links
https://01.org/intel-secl/

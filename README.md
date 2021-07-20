# ISecL K8s Extenstions 

`ISecL K8s Extensions` which includes ISecL K8s extended scheduler, ISecL K8s custom controller components and certification generation scripts for trusted launch of containers.
Key Components:
- ISecL K8s extended scheduler
    The ISecL Extended Scheduler verifies trust report and asset tag signature for each of the K8s Worker Node annotation against Pod matching expressions in pod yaml file using ISecL Integration hub public key.
    The signature verification ensures the integrity of labels created using isecl hostattribute crds on each of the worker nodes. The verification happens at the time of pod scheduling.
- ISecL K8s custom controller
    The ISecL Custom Controller creates/updates labels and annotation of K8s Worker Nodes whenever isecl.hostattributes crd objects are created or updated through K8s kube-apiserver.
- Certificate generation scripts
    These scripts creates kubernetes hostattributes.crd.isecl.intel.com from which the crd objects will be created for each of the tenant, then it creates the client and server certificates.
    The client certificate is created for root user and root user will be having RBAC on get,list,delete,patch,deletecollection,create and update operations on the hostattributes.crd.isecl.intel.com.

## System Requirements
- RHEL 8.2
- Epel 8 Repo
- Proxy settings if applicable

## Software requirements
- git
- makeself
- `go` version 1.14.4

# Step-By-Step Build Instructions

## Install required shell commands

### Install tools from `yum`
```shell
sudo yum install -y git wget
```

### Install `go` version 1.14.4
The `ISecL K8s Extensions` requires Go version 1.14.4 that has support for `go modules`. The build was validated with the latest version go1.14.4 of `go`. It is recommended that you use go1.14.4 version of `go`. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz
tar -xzf go1.14.4.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export GOPATH=<path of project workspace>
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

## Build ISecL K8s Extenstions

```shell
git clone https://github.com/intel-secl/k8s-extensions.git
cd k8s-extensions
make all
```

### Deploy
Pre-requisites
- cffsl
- cfssljson

Install Pre-requisites
```console
   wget http://pkg.cfssl.org/R1.2/cfssl_linux-amd64
   mv cfssl_linux-amd64 /usr/local/bin/cfssl
   wget http://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
   mv cfssljson_linux-amd64 /usr/local/bin/cfssljson
```

#### Deploy isecl-controller
* Copy isecl-k8s-extensions.tar.gz output build to /opt/ directory and extract the contents
```console
cd /opt/
tar -xvzf isecl-k8s-extensions.tar.gz
```
* Load docker images isecl-controller and isecl-scheduler
```console
docker load -i docker-isecl-controller-v*.tar
docker load -i docker-isecl-scheduler-v*.tar
``` 
* Create hostattributes.crd.isecl.intel.com crd
```console
cd /opt/isecl-k8s-extensions
kubectl apply -f yamls/crd-1.17.yaml
```

* Check whether the crd is created 
```console
kubectl get crds
```
* Fields for isecl-controller configuration in isecl-controller.yaml

Field | Required | Type | Default | Comments |
-------|----------|------|---------|---------|
LOG_LEVEL | `Optional` |`string` | INFO | Determines the log level |
LOG_MAX_LENGTH | `Optional` |`int` | 1500 | Maximum length of characters in a line in log file |
TAG_PREFIX | `Optional` | `string` | isecl. | A custom prefix which can be applied to isecl attributes that are pushed from IH. |
TAINT_UNTRUSTED_NODES | `Optional` | `string` | false | If set to true. NoExec taint applied to the nodes for which trust status is set to false |

* Deploy isecl-controller
```console
kubectl apply -f yamls/isecl-controller.yaml
```

* Check whether the isecl-controller is up and running
```console
kubectl get deploy -n isecl
```

* Create clusterrolebinding for ihub to get access to cluster nodes
```
kubectl create clusterrolebinding isecl-clusterrole --clusterrole=system:node --user=system:serviceaccount:isecl:isecl
```

#### Fetch token required for ihub installation
```console
kubectl get secrets -n isecl
kubectl describe secret default-token-<name> -n isecl
```

#### Deploy isecl-scheduler

* Create a directory for storing certificates
```console
mkdir secrets
```
* Create tls key pair for isecl-scheduler service, which is signed by k8s apiserver.crt
```console
chmod +x create_k8s_extsched_cert.sh
./create_k8s_extsched_cert.sh -n "K8S Extended Scheduler" -s "$MASTER_IP","$HOSTNAME" -c /etc/kubernetes/pki/ca.crt -k /etc/kubernetes/pki/ca.key
```

* Copy ihub_public_key.pem from isecl integration hub to **secrets** directory

* Create kubernetes secrets ```scheduler-secret``` for isecl-scheduler
```console
kubectl create secret generic scheduler-certs --namespace isecl --from-file=secrets
```

* Fields for isecl-scheduler configuration in isecl-scheduler.yaml

Field | Required | Type | Default | Comments
-------|----------|------|---------|--------
PORT | `Optional` | `string` | 8888 | ISecl scheduler service port  |
HVS_IHUB_PUBLIC_KEY_FILE_PATH | `Required` |`string` | | Required for IHub with HVS Attestation |
SGX_IHUB_PUBLIC_KEY_FILE_PATH | `Required` |`string` | | Required for IHub with SGX Attestation |
LOG_LEVEL | `Optional` |`string` | INFO | Determines the log level |
LOG_MAX_LENGTH | `Optional` |`int` | 1500 | Maximum length of characters in a line in log file |
TLS_CERT_PATH | `Required` | `string` | | Tls certificate path for isecl scheduler service |
TLS_KEY_PATH | `Required` | `string` | | Tls key path for isecl scheduler service |
TAG_PREFIX | `Optional` | `string` | isecl. | A custom prefix which can be applied to isecl attributes that are pushed from IH |

* Deploy isecl-scheduler
```console
kubectl apply -f yamls/isecl-scheduler.yaml
```

* Check whether the isecl-scheduler is up and running
```console
kubectl get deploy -n isecl
```

#### Configure kube-scheduler to establish communication with isecl-scheduler
* Add scheduler-policy.json under kube-scheduler section /etc/kubernetes/manifests/kube-scheduler.yaml as mentioned below
```console
	spec:
          containers:
	  - command:
            - kube-scheduler
              --policy-config-file : "/opt/isecl-k8s-extensions/scheduler-policy.json"
```

* Add mount path for isecl extended scheduler under container section /etc/kubernetes/manifests/kube-scheduler.yaml as mentioned below
```console
	containers:
		- mountPath: /opt/isecl-k8s-extensions
		name: extendedsched
		readOnly: true
```

* Add volume path for isecl extended scheduler under volumes section /etc/kubernetes/manifests/kube-scheduler.yaml as mentioned below
```console
	spec:
	volumes:
	- hostPath:
		path: /opt/isecl-k8s-extensions
		type: ""
		name: extendedsched
```

* Restart Kubelet which restart all the k8s services including kube base scheduler
```console
	systemctl restart kubelet
```

#### Uninstalling the isecl-k8s-extensions
* Uninstall the isecl-k8s-extensions by running following commands
    - Delete isecl-scheduler service
    ```kubectl delete svc isecl-scheduler-svc -n isecl```
    - Delete isecl-controller and isecl-scheduler deployments
    ```kubectl delete deployment isecl-controller isecl-scheduler -n isecl```
    - Delete hostattributes.crd.isecl.intel.com crds
     ```kubectl delete crds hostattributes.crd.isecl.intel.com```
    - Remove the directories at /opt/isecl-k8s-extensions/
    ```rm -rf /opt/isecl-k8s-extensions```

### Product Guide: 
For more details on the product, installation and deployment strategies, please go through following, (Refer to latest and use case wise guide)

- [https://01.org/intel-secl/documentation/intel%C2%AE-secl-dc-product-guide](https://01.org/intel-secl/documentation/intel%C2%AE-secl-dc-product-guide)
- [https://github.com/intel-secl/docs](https://github.com/intel-secl/docs)

### Release Notes: 
[https://01.org/intel-secl/documentation/intel%C2%AE-secl-dc-release-notes](https://01.org/intel-secl/documentation/intel%C2%AE-secl-dc-release-notes)

### Issues: 
Feel free to raise build, deploy or even runtime issues here,

[https://github.com/intel-secl/k8s-extensions/issues](https://github.com/intel-secl/k8s-extensions/issues)

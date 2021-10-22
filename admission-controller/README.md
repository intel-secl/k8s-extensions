# K8s Node Tainting using Admission Controller

An admission controller is a piece of code that intercepts requests to the Kubernetes API server prior to persistence of
the object, but after the request is authenticated and authorized.

Mutating admission webhooks are invoked first, and can modify objects sent to the API server to enforce custom defaults.

Node tainting webhook admission controller, when the node joining request comes to k8s, it intercepts the node joining
API request and adds NoSchedule and NoExecute taint to the Node.

To taint a Node, when its rebooted, we should set `TAINT_REBOOTED_NODES` to "true".

As a reconciliation mechanism, we use isecl-controller, by setting `TAINT_REBOOTED_NODES` to "
true" `TAINT_REGISTERED_NODES`  to "true" in isecl-controller.yml.

## System Requirements

-RHEL 8.2 
-Epel 8 Repo 
-Proxy settings if applicable.

Note In no_proxy, add .svc,.svc.cluster.local and then do
```
 kubeadm init
```

## Software requirements

-git 
-makeself 
-go version 1.16 
-Kubernetes 19.0 or later.

## Deploy using boostrap script

```
cd manifest
./isecl-bootstrap.sh up admission-controller
```

## Delete deployment using boostrap script

```
cd manifest
./isecl-bootstrap.sh down admission-controller
```

## Manual deployment step

### TLS certificate notes for Webhook

In order for our webhook to be invoked by Kubernetes, we need a TLS certificate.

```
./create_k8s_extsched_cert.sh -n "K8S Admission Controller" -s "node-tainting-webhook","node-tainting-webhook.isecl.svc.cluster.local","node-tainting-webhook.isecl.svc","localhost","127.0.0.1","$MASTER_IP","$HOSTNAME" -c /etc/kubernetes/pki/ca.crt -k /etc/kubernetes/pki/ca.key
```

certificate and key would be available as ./server.crt and ./server.key

Update the node-tainting-webhook-tls.yaml tls.crt and tls.key using 
```
cat ./server.crt | base64 | tr -d '\n' 
cat ./server.key | base64 | tr -d '\n'
```

### Update the CA bundle in webhook.yml

CA Bundle is used for signing new TLS certificates.

CA bundle can be obtained using the below command, run it on mater node and update in webhook.yaml

```
kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}'
```

### Building Node Tainting Admission Controller
```
cd admission-controller 
make oci-archive
```

### To upload to Private registry

Copy the tar file obtained to the private registry using skopeo copy command. example :

```
skopeo copy oci-archive: node-tainting-webhook-v1.tar docker://10.105.168.18:5000/node-tainting-webhook:v1 --dest-tls-verify=false
```

### Update the image in admission-controller.yaml, example

image: 10.105.168.18:5000/node-tainting-webhook:v1

### Deploy secrets

```
kubectl -n isecl apply -f node-tainting-webhook-tls.yaml
```

### Deploy Admission Controller Code

```
kubectl -n isecl apply -f admission-controller.yaml
```

### Deploy Webhook

```
kubectl -n isecl apply -f webhook.yaml
```

## Node Joining

When the worker node is being joined to the k8s cluster, Untrusted:True NoExecute and NoSchedule taint would be added to
the worker Node.

Note:
Following are the commands used to delete the admission controller deployment and related secrets, rbac and webhook.

```
kubectl delete service node-tainting-webhook -n isecl
kubectl delete deploy node-tainting-webhook -n isecl
kubectl delete MutatingWebhookConfiguration node-tainting-webhook -n isecl
kubectl delete ServiceAccount node-tainting-webhook -n isecl 
kubectl delete ClusterRole node-tainting-webhook 
kubectl delete ClusterRoleBinding node-tainting-webhook 
kubectl delete secret node-tainting-webhook-tls --n=isecl
```

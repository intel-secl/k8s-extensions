#!/bin/bash

echo "Installing Pre-requisites"
which cfssl
if [ $? -ne 0 ]
then
  wget http://pkg.cfssl.org/R1.2/cfssl_linux-amd64
  chmod +x cfssl_linux-amd64
  mv cfssl_linux-amd64 /usr/local/bin/cfssl
fi

which cfssljson
if [ $? -ne 0 ]
then
  wget http://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
  chmod +x cfssljson_linux-amd64
  mv cfssljson_linux-amd64 /usr/local/bin/cfssljson
fi


K8S_EXTENSIONS_DIR=/opt/isecl-k8s-extensions
CERTS=certificate-generation-scripts
ATTESTATION_HUB_KEYSTORES=/opt/isecl-k8s-extensions/attestation-hub-keystores
K8S_EXTENSIONS_CONFIG_DIR=$K8S_EXTENSIONS_DIR/config
K8S_EXTENSIONS_LOG_DIR=/var/log/isecl-k8s-extensions
TAG_PREFIX_CONF=tag_prefix.conf

mkdir -p $K8S_EXTENSIONS_DIR
mkdir -p $K8S_EXTENSIONS_CONFIG_DIR
mkdir -p $K8S_EXTENSIONS_LOG_DIR

kubectl cluster-info 2>/dev/null
if [ $? -ne 0 ]
then
   echo "Error while running kubectl cluster-info command Set Environment variable KUBECONFIG to path of admin.conf"
   exit 1
fi

export KUBECONFIG=/etc/kubernetes/admin.conf

echo ""
echo "Configuring tag prefix"
echo ""

cat > $K8S_EXTENSIONS_CONFIG_DIR/$TAG_PREFIX_CONF<<EOF
{
      "trusted":"isecl."
}
EOF

echo ""
echo "Deploying isecl k8s controller"

#Load isecl k8s controller docker image into local repository
docker load -i docker-isecl-controller-*.tar

#Untaint the master node for deploying isecl k8s controller on master node
HOSTNAME=${HOSTNAME:-$(hostname)}
kubectl taint nodes ${HOSTNAME} node-role.kubernetes.io/master:NoSchedule- 2>/dev/null
kubectl apply -f yamls/crd-1.14.yaml
kubectl apply -f yamls/secl-controller.yaml

cp -r yamls $K8S_EXTENSIONS_DIR/
echo ""
echo "Installing Pre requisites for generating certificates"
echo ""

chmod +x create_certs.sh
./create_certs.sh

if [ $? -ne 0 ]
then
  echo "Error while creating certificates."
  exit 1
fi


./isecl-k8s-scheduler-v*.bin

systemctl daemon-reload
systemctl restart kubelet
systemctl restart isecl-k8s-scheduler.service

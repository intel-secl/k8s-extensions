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

# Java configurations
JAVA_CMD=$(type -p java | xargs readlink -f)
JAVA_HOME=$(dirname $JAVA_CMD | xargs dirname | xargs dirname)

K8S_EXTENSIONS_DIR=/opt/isecl-k8s-extensions
CERTS=certificate-generation-scripts
ATTESTATION_HUB_KEYSTORES=/opt/isecl-k8s-extensions/attestation-hub-keystores
K8S_EXTENSIONS_CONFIG_DIR=$K8S_EXTENSIONS_DIR/config
K8S_EXTENSIONS_LOG_DIR=/var/log/isecl-k8s-extensions
TAG_PREFIX_CONF=tag_prefix.conf
K8S_EXTENSIONS_SCHEDULER_DIR=$K8S_EXTENSIONS_DIR/isecl-k8s-scheduler
K8S_EXTENSIONS_SCHEDULER_CONFIG_DIR=${K8S_EXTENSIONS_SCHEDULER_DIR}/config


if [ -f "${JAVA_HOME}/jre/lib/security/java.security" ]; then
  echo "Setting default keystore to pkcs12"
  sed -i -e 's/keystore\.type.*/keystore\.type=pkcs12/' "${JAVA_HOME}/jre/lib/security/java.security" 2>/dev/null
fi

mkdir -p $K8S_EXTENSIONS_DIR
mkdir -p $K8S_EXTENSIONS_CONFIG_DIR
mkdir -p $K8S_EXTENSIONS_LOG_DIR
mkdir -p ${K8S_EXTENSIONS_SCHEDULER_CONFIG_DIR}

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

IFS=' '
k8sversion=$(kubelet --version)
read -ra ADDR <<<"$k8sversion"
version=${ADDR[1]}
IFS='.'
read -ra ADDR <<<"$version"
majorVersion=${ADDR[0]}
minorVersion=${ADDR[1]}

if [[ "$majorVersion" == "v1" && "$minorVersion" -ge 16 ]]; then
  kubectl apply -f yamls/crd-1.17.yaml
else 
  kubectl apply -f yamls/crd-1.14.yaml
fi

kubectl apply -f yamls/isecl-controller.yaml

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


echo "Deploying ISecL K8S Extended Scheduler"

cp isecl-extended-scheduler-config.json ${K8S_EXTENSIONS_SCHEDULER_CONFIG_DIR}/
cp scheduler-policy.json ${K8S_EXTENSIONS_SCHEDULER_CONFIG_DIR}/


chmod +x create_k8s_extsched_cert.sh

echo ./create_k8s_extsched_cert.sh -n "K8S Extended Scheduler" -s "$MASTER_IP","$HOSTNAME" -c /etc/kubernetes/pki/ca.crt -k /etc/kubernetes/pki/ca.key

./create_k8s_extsched_cert.sh -n "K8S Extended Scheduler" -s "$MASTER_IP","$HOSTNAME" -c "/etc/kubernetes/pki/ca.crt"  -k "/etc/kubernetes/pki/ca.key"

if [ $? -ne 0 ]
then
  echo "Error while creating certificates for extended scheduler"
  exit 1
fi

mv server.crt ${K8S_EXTENSIONS_SCHEDULER_CONFIG_DIR}/
mv server.key ${K8S_EXTENSIONS_SCHEDULER_CONFIG_DIR}/

docker load -i docker-isecl-scheduler-*.tar
kubectl apply -f yamls/isecl-scheduler.yaml

systemctl daemon-reload
systemctl restart kubelet
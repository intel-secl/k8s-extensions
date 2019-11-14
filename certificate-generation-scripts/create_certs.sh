#!/bin/bash

K8S_EXTENSION_DIR=/opt/isecl-k8s-extensions
K8S_EXTENDED_SCHEDULER_CONFIG_DIR=$K8S_EXTENSION_DIR/isecl-k8s-scheduler/config/
HUB_KEYSTORE_DIR=$K8S_EXTENSION_DIR/attestation-hub-keystores/

yum -y install java-1.8.0-openjdk openssl

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

chmod +x create_k8s_certs.sh

./create_k8s_certs.sh -u root -s $HOSTNAME,$MASTER_IP \
            -c /etc/kubernetes/pki/ca.crt \
            -k /etc/kubernetes/pki/ca.key \
            -a /etc/kubernetes/pki/apiserver.crt \
            -r "hostattributes.crd.isecl.intel.com" \
            -v "get,list,delete,patch,deletecollection,create,update" \
            -d "$HUB_KEYSTORE_DIR" \
            -f "false" \
            -e "false"

if [ $? -ne 0 ]
then
  exit 1
fi

mkdir $HUB_KEYSTORE_DIR			


rm create_k8s_certs.sh
rm create_certs.sh

cp root_k8s_client.p12 $HUB_KEYSTORE_DIR/
cp root_k8s_trust.p12 $HUB_KEYSTORE_DIR/
cp root_keystore.properties $HUB_KEYSTORE_DIR/


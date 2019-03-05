#!/bin/bash

K8S_EXTENSION_DIR=/opt/isecl-k8s-extensions
K8S_EXTENSION_CONFIG_DIR=$K8S_EXTENSION_DIR/config/
HUB_KEYSTORE_DIR=/root/attestation-hub-keystores/

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
            -r "geolocationcrds.isecl.intel.com,platformcrds.isecl.intel.com" \
            -v "get,list,delete,patch,deletecollection,create,update" \
            -d "/root/ahubkeystore" \
            -f "false" \
            -e "false"

if [ $? -ne 0 ]
then
  exit 1
fi

mkdir $HUB_KEYSTORE_DIR			
chmod +x create_k8s_extsched_cert.sh

echo ./create_k8s_extsched_cert.sh -n "K8S Extended Scheduler" \ -s $MASTER_IP,$HOSTNAME \ -c /etc/kubernetes/pki/ca.crt -k /etc/kubernetes/pki/ca.key

./create_k8s_extsched_cert.sh -s $MASTER_IP,$HOSTNAME \
        -c /etc/kubernetes/pki/ca.crt \
        -k /etc/kubernetes/pki/ca.key

if [ $? -ne 0 ]
then
  exit 1
fi

chmod +755 $K8S_EXTENSION_CONFIG_DIR
cp server.crt $K8S_EXTENSION_CONFIG_DIR
cp server.key $K8S_EXTENSION_CONFIG_DIR


rm create_k8s_extsched_cert.sh
rm create_k8s_certs.sh
rm create_certs.sh
rm server.crt
rm server.key

cp ./* $HUB_KEYSTORE_DIR -r


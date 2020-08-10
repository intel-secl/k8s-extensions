#!/bin/bash

K8S_EXTENSION_DIR=/opt/isecl-k8s-extensions
K8S_EXTENDED_SCHEDULER_CONFIG_DIR=$K8S_EXTENSION_DIR/isecl-k8s-scheduler/config/
HUB_KEYSTORE_DIR=$K8S_EXTENSION_DIR/attestation-hub-keystores/
K8S_API_SERVER_CERT=/etc/kubernetes/pki/apiserver.crt
yum -y install java-1.8.0-openjdk openssl

TRUST_P12=k8s_trust.p12

export TRUST_STORE_PASS=`openssl rand -base64 128 | tr -dc _A-Z-a-z-0-9 | head -c64`
keytool -import -keystore $TRUST_P12  -alias k8s_server -file $K8S_API_SERVER_CERT -noprompt -deststorepass:env TRUST_STORE_PASS


cat > k8s_keystore.properties<<EOF
kubernetes.server.keystore=/opt/attestation-hub/configuration/`ls $TRUST_P12`
kubernetes.server.keystore.password=${TRUST_STORE_PASS}
EOF

export SERVER_KEYSTORE=`ls $TRUST_P12`
export KEYSTORE_CONF=`ls k8s_keystore.properties`


mkdir $HUB_KEYSTORE_DIR			

cp $TRUST_P12 $HUB_KEYSTORE_DIR/
cp k8s_keystore.properties $HUB_KEYSTORE_DIR/

# Clean up

unset TRUST_STORE_PASS
unset SERVER_KEYSTORE
unset KEYSTORE_CONF 

rm create_certs.sh


#!/bin/bash

K8S_EXTENSIONS_DIR=/opt/isecl-k8s-extensions
CERTS=certificate-generation-scripts
ATTESTATION_HUB_KEYSTORES=/opt/isecl-k8s-extensions/attestation-hub-keystores
K8S_EXTENSIONS_CONFIG_DIR=$K8S_EXTENSIONS_DIR/config
TAG_PREFIX_CONF=tag_prefix.conf

mkdir -p $K8S_EXTENSIONS_DIR
mkdir -p $K8S_EXTENSIONS_CONFIG_DIR
cp isecl-k8s-extensions.sh $K8S_EXTENSIONS_DIR/ && chmod +x $K8S_EXTENSIONS_DIR/isecl-k8s-extensions.sh
ln -s $K8S_EXTENSIONS_DIR/isecl-k8s-extensions.sh /usr/local/bin/isecl-k8s-extensions

export KUBECONFIG=/etc/kubernetes/admin.conf

echo ""
echo "Configuring tag prefix"
echo ""

cat > $K8S_EXTENSIONS_CONFIG_DIR/$TAG_PREFIX_CONF<<EOF
{
      "trusted":"isecl."
}
EOF



./isecl-k8s-controller-v*.bin

echo ""
echo "Installing Pre requisites for generating certificates"
echo ""

chmod +x create_certs.sh
./create_certs.sh

if [ $? -ne 0 ]
then
  echo "Error while creating certificates."
  isecl-k8s-extensions uninstall
  exit 1
fi
 


./isecl-k8s-scheduler-v*.bin

systemctl daemon-reload
systemctl restart kubelet
systemctl restart isecl-k8s-scheduler.service
systemctl restart isecl-k8s-controller.service


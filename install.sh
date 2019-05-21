#!/bin/bash

K8S_TEMP_EXTENSIONS=/tmp/k8s-extensions

K8S_EXTENSIONS_DIR=/opt/isecl-k8s-extensions
K8S_EXTENSIONS_BIN_DIR=$K8S_EXTENSIONS_DIR/bin
CUSTOM_CONTROLLER=custom-controller
EXTENDED_SCHEDULER=extended-schedular
CERTS=certificate-generation-scripts
TAG_PREFIX_CONF=tag_prefix.conf
ATTESTATION_HUB_KEYSTORES=/root/attestation-hub-keystores
K8S_EXTENSIONS_CONFIG_DIR=$K8S_EXTENSIONS_DIR/config


mkdir $K8S_TEMP_EXTENSIONS
cp ./* $K8S_TEMP_EXTENSIONS/ -r
mkdir -p $K8S_EXTENSIONS_CONFIG_DIR
chmod -R +755 $K8S_EXTENSIONS_DIR

cd $K8S_TEMP_EXTENSIONS
chmod +x isecl-k8s-extensions.sh
cp isecl-k8s-extensions.sh $K8S_EXTENSIONS_DIR/
ln -s $K8S_EXTENSIONS_DIR/isecl-k8s-extensions.sh /usr/local/bin/isecl-k8s-extensions

export KUBECONFIG=/etc/kubernetes/admin.conf

cd $K8S_TEMP_EXTENSIONS/$CUSTOM_CONTROLLER
make install
echo ""
echo "Installing Custom Controller"
echo ""

echo ""
echo "Starting Custom Controller"
echo ""

systemctl daemon-reload
systemctl start isecl-k8s-controller.service

cd $K8S_TEMP_EXTENSIONS/$CERTS
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
 
cd $K8S_TEMP_EXTENSIONS/$EXTENDED_SCHEDULER
make install
echo ""
echo "Installing Extended Scheduler"
echo ""

echo ""
echo "Configuring tag prefix"
echo ""

cat > $TAG_PREFIX_CONF<<EOF
{
      "trusted":"isecl."
}
EOF

cp $TAG_PREFIX_CONF $K8S_EXTENSIONS_DIR

cd $K8S_TEMP_EXTENSIONS

cat policy.json | sed 's/127.0.0.1/'$MASTER_IP'/g' > scheduler-policy.json
cat $K8S_EXTENSIONS_CONFIG_DIR/isecl-extended-scheduler-config.json | sed 's/127.0.0.1/'$MASTER_IP'/g' >  temp-config.json
cat temp-config.json > $K8S_EXTENSIONS_CONFIG_DIR/isecl-extended-scheduler-config.json
cp scheduler-policy.json $K8S_EXTENSIONS_BIN_DIR/

echo ""
echo "Starting ISecL Extended Scheduler"
echo ""

systemctl daemon-reload
systemctl restart kubelet
systemctl start isecl-k8s-scheduler.service
systemctl restart isecl-k8s-controller.service

rm -rf $K8S_TEMP_EXTENSIONS 

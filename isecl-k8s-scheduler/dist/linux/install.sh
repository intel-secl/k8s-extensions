#!/bin/bash

echo "Installing ISecL K8S Extended Scheduler"

COMPONENT_NAME=isecl-k8s-scheduler
PRODUCT_HOME=/opt/isecl-k8s-extensions/${COMPONENT_NAME}
BIN_PATH=${PRODUCT_HOME}/bin
CONFIG_PATH=${PRODUCT_HOME}/config
SERVICE_INSTALL_DIR=/etc/systemd/system
SERVICE_CONFIG=${COMPONENT_NAME}.service

mkdir -p ${BIN_PATH}
mkdir -p ${CONFIG_PATH}
cp isecl-extended-scheduler-config.json ${CONFIG_PATH}
cp ${SERVICE_CONFIG} ${PRODUCT_HOME}/

cp $COMPONENT_NAME $BIN_PATH/
ln -sfT $BIN_PATH/$COMPONENT_NAME /usr/bin/$COMPONENT_NAME
cp scheduler-policy.json $CONFIG_PATH/

chmod +x create_k8s_extsched_cert.sh

echo ./create_k8s_extsched_cert.sh -n "K8S Extended Scheduler" \ -s $MASTER_IP,$HOSTNAME \ -c /etc/kubernetes/pki/ca.crt -k /etc/kubernetes/pki/ca.key

./create_k8s_extsched_cert.sh -s $MASTER_IP,$HOSTNAME \
        -c /etc/kubernetes/pki/ca.crt \
        -k /etc/kubernetes/pki/ca.key

if [ $? -ne 0 ]
then
  exit 1
fi

cp server.crt $CONFIG_PATH
cp server.key $CONFIG_PATH


systemctl enable $PRODUCT_HOME/$SERVICE_CONFIG

systemctl daemon-reload
systemctl start $SERVICE_CONFIG

echo "ISecL K8S Extended Scheduler Started..."

#!/bin/bash

echo "Installing ISecL K8S Custom controller"

COMPONENT_NAME=isecl-k8s-controller
PRODUCT_HOME=/opt/isecl-k8s-extensions/${COMPONENT_NAME}
BIN_PATH=${PRODUCT_HOME}/bin
SERVICE_INSTALL_DIR=/etc/systemd/system
SERVICE_CONFIG=${COMPONENT_NAME}.service


mkdir -p $PRODUCT_HOME
mkdir -p $BIN_PATH

cp $COMPONENT_NAME $BIN_PATH/
ln -sfT $BIN_PATH/$COMPONENT_NAME /usr/bin/$COMPONENT_NAME

cp $SERVICE_CONFIG $PRODUCT_HOME/
systemctl enable $PRODUCT_HOME/$SERVICE_CONFIG

systemctl daemon-reload
systemctl start $SERVICE_CONFIG

if [ $? -ne 0 ]
then
   echo "Failed to start ISecL K8S Custom Controller"
  exit 1
fi

echo "ISecL K8S Custom Controller Started..."




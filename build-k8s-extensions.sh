#!/bin/bash

TAG=$(git describe --tags --abbrev=0 2> /dev/null)

CURR_DIR=`pwd`
echo "Building isecl-k8s-controller"
cd $CURR_DIR/isecl-k8s-controller
make all

echo "Building isecl-k8s-scheduler"
cd $CURR_DIR/isecl-k8s-scheduler
make all

cd $CURR_DIR
sed -i 's/image: isecl\/k8s-controller.*/image: isecl\/k8s-controller:'${TAG}'/g' isecl-k8s-controller/yamls/isecl-controller.yaml
sed -i 's/image: isecl\/k8s-scheduler.*/image: isecl\/k8s-scheduler:'${TAG}'/g' isecl-k8s-scheduler/yamls/isecl-scheduler.yaml
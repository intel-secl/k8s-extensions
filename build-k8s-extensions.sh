#!/bin/bash

CURR_DIR=`pwd`
echo "Building isecl-k8s-controller"
cd $CURR_DIR/isecl-k8s-controller
make all

echo "Building isecl-k8s-scheduler"
cd $CURR_DIR/isecl-k8s-scheduler
make all

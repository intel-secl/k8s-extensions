#!/bin/bash

CURR_DIR=`pwd`
BUILD_WORKSPACE_DIR=$CURR_DIR/k8s-extensions-build
CUSTOM_CONTROLLER_DIR=$CURR_DIR/custom-controller
EXTENDED_SCHEDULER_DIR=$CURR_DIR/extended-scheduler
OUTPUT_DIR=out

mkdir $BUILD_WORKSPACE_DIR
mkdir $CUSTOM_CONTROLLER_DIR
mkdir $EXTENDED_SCHEDULER_DIR

cd $BUILD_WORKSPACE_DIR
git clone https://github.com/intel-secl/k8s-custom-controller.git
cd k8s-custom-controller
git fetch
git checkout master
make
cp isecl-k8s-controller-1.0-SNAPSHOT isecl-k8s-controller.service Makefile $CUSTOM_CONTROLLER_DIR/


cd $BUILD_WORKSPACE_DIR
git clone https://github.com/intel-secl/k8s-extended-scheduler.git
cd k8s-extended-scheduler
git fetch
git checkout master
make
cp -r config/ isecl-k8s-scheduler-1.0-SNAPSHOT Makefile isecl-extended-scheduler-config.json isecl-k8s-scheduler.service $EXTENDED_SCHEDULER_DIR
cd $CURR_DIR

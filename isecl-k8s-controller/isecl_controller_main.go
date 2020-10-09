/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"fmt"
	commLog "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log"
	commLogMsg "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log/message"
	commLogInt "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log/setup"

	"intel/isecl/k8s-custom-controller/v3/crdController"
	"os"
	"strconv"
	"sync"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"github.com/sirupsen/logrus"
	"github.com/pkg/errors"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

const  logFilePath = "/var/log/isecl-k8s-extensions/isecl-controller.log"
var defaultLog = commLog.GetDefaultLogger()

func configureLogs(logFile *os.File, loglevel string, maxLength int) error {

	lv, err := logrus.ParseLevel(loglevel)
	if err != nil {
		return errors.Wrap(err, "Failed to initiate loggers. Invalid log level: "+loglevel)
	}

	f := commLog.LogFormatter{MaxLength: maxLength}
	commLogInt.SetLogger(commLog.DefaultLoggerName, lv, &f, logFile, false)

	defaultLog.Info(commLogMsg.LogInit)
	return nil
}


func main() {

	fmt.Println("Starting ISecL Custom Controller")

	logLevel := os.Getenv("LOG_LEVEL")
	logMaxLength, err := strconv.Atoi(os.Getenv("LOG_MAX_LENGTH"))
	if err != nil {
		fmt.Printf("Error while parsing variable config LOG_MAX_LENGTH error: %v, setting LOG_MAX_LENGTH to 1500, %v", err)
		logMaxLength = 1500
	}

	skipCrdCreate, err := strconv.ParseBool(os.Getenv("SKIP_CRD_CREATE"))
	if err != nil {
		fmt.Printf("Error while parsing variable config SKIP_CRD_CREATE error: %v, setting SKIP_CRD_CREATE to true", err)
		skipCrdCreate = false
	}
	fmt.Printf("SKIP_CRD_CREATE is set to %v", skipCrdCreate)

	taintUntrustedNodes, err := strconv.ParseBool(os.Getenv("TAINT_UNTRUSTED_NODES"))
	if err != nil {
		fmt.Println("Error while parsing variable config TAINT_UNTRUSTED_NODES error: %v, setting TAINT_UNTRUSTED_NODES to false", err)
		taintUntrustedNodes = false
	}
	fmt.Printf("TAINT_UNTRUSTED_NODES is set to %v", taintUntrustedNodes)

	tagPrefix := os.Getenv("TAG_PREFIX")
	if tagPrefix != "" {
		fmt.Println("Env Variable TAG_PREFIX is empty setting to default value isecl.")
		tagPrefix = "isecl."
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		fmt.Println("Unable to open log file")
		return
	}

	configureLogs(logFile, logLevel, logMaxLength)

	kubeConf := os.Getenv("kubeconf")

	config, err := GetClientConfig(kubeConf)
	if err != nil {
		defaultLog.Errorf("Error in config %v", err)
		return
	}

	cs, err := apiextcs.NewForConfig(config)
	if err != nil {
		defaultLog.Errorf("Error in config %v", err)
		return
	}

	//Create mutex to sync operation between the two CRD threads
	var crdmutex = &sync.Mutex{}

	if !skipCrdCreate {
		CrdDef := crdController.GetHACrdDef()
		//crdController.NewIseclCustomResourceDefinition to create CRD
		err = crdController.NewIseclCustomResourceDefinition(cs, &CrdDef)
		if err != nil {
			defaultLog.Errorf("Error in creating hostattributes CRD %v", err)
			return
		}
	}

	if taintUntrustedNodes {
		crdController.TaintUntrustedNodes = true
	}
	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "iseclcontroller")

	indexer, informer := crdController.NewIseclHAIndexerInformer(config, queue, crdmutex, tagPrefix)

	controller := crdController.NewIseclHAController(queue, indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	defaultLog.Info("Waiting for updates on ISecl Custom Resource Definitions")

	// Wait forever
	select {}
}

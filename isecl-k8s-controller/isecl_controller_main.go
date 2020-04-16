/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"flag"
	"intel/isecl/k8s-custom-controller/v2/crdController"
	"intel/isecl/k8s-custom-controller/v2/util"
	"os"
	"sync"
	"strconv"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

var Log = util.GetLogger()
const TrustedPrefixConf = "/opt/isecl-k8s-extensions/config/tag_prefix.conf"

func main() {

	Log.Infof("Starting ISecL Custom Controller")

        logLevel := os.Getenv("LOG_LEVEL")

        util.SetLogger(logLevel)

	skipCrdCreate, err := strconv.ParseBool(os.Getenv("SKIP_CRD_CREATE"))
	if err != nil {
		Log.Info("Error while parsing variable config SKIP_CRD_CREATE error: %v, setting SKIP_CRD_CREATE to true", err)
		skipCrdCreate = false
	}
	Log.Infof("SKIP_CRD_CREATE is set to %v", skipCrdCreate)

	taintUntrustedNodes, err := strconv.ParseBool(os.Getenv("TAINT_UNTRUSTED_NODES"))
	if err != nil {
		Log.Info("Error while parsing variable config TAINT_UNTRUSTED_NODES error: %v, setting TAINT_UNTRUSTED_NODES to false", err)
		taintUntrustedNodes = false
	}
	Log.Infof("TAINT_UNTRUSTED_NODES is set to %v", taintUntrustedNodes)
	
	kubeConf := flag.String("kubeconf", "", "Path to a kube config. ")
	flag.Parse()

	config, err := GetClientConfig(*kubeConf)
	if err != nil {
		Log.Errorf("Error in config %v", err)
		return
	}

	cs, err := apiextcs.NewForConfig(config)
	if err != nil {
		Log.Errorf("Error in config %v", err)
		return
	}

	//Create mutex to sync operation between the two CRD threads
	var crdmutex = &sync.Mutex{}

        if !skipCrdCreate {
                CrdDef := crdController.GetHACrdDef()
                //crdController.NewIseclCustomResourceDefinition to create CRD
                err = crdController.NewIseclCustomResourceDefinition(cs, &CrdDef)
                if err != nil {
                        Log.Errorf("Error in creating hostattributes CRD %v", err)
                        return
                }
        }

        if taintUntrustedNodes {
                crdController.TaintUntrustedNodes = true
        }
	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "iseclcontroller")

	indexer, informer := crdController.NewIseclHAIndexerInformer(config, queue, crdmutex, TrustedPrefixConf)

	controller := crdController.NewIseclHAController(queue, indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	Log.Infof("Waiting for updates on  ISecl Custom Resource Definitions")

	// Wait forever
	select {}
}

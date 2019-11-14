/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"flag"
	"fmt"
	"k8s_custom_cit_controllers-k8s_custom_controllers/crdController"
	"sync"

	"github.com/golang/glog"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

const TrustedPrefixConf = "/opt/isecl-k8s-extensions/config/tag_prefix.conf"

func main() {

	glog.V(4).Infof("Starting ISecL Custom Controller")

	var Usage = func() {
		fmt.Println("Usage: ./isecl-k8s-controller-1.0-SNAPSHOT -kubeconf=<file path>")
	}

	kubeConf := flag.String("kubeconf", "", "Path to a kube config. ")
	flag.Parse()

	if *kubeConf == "" {
		Usage()
		return
	}

	config, err := GetClientConfig(*kubeConf)
	if err != nil {
		glog.Errorf("Error in config %v", err)
		return
	}

	cs, err := apiextcs.NewForConfig(config)
	if err != nil {
		glog.Errorf("Error in config %v", err)
		return
	}

	//Create mutex to sync operation between the two CRD threads
	var crdmutex = &sync.Mutex{}

	CrdDef := crdController.GetHACrdDef()

	//crdController.NewIseclCustomResourceDefinition to create CRD
	err = crdController.NewIseclCustomResourceDefinition(cs, &CrdDef)
	if err != nil {
		glog.Errorf("Error in creating platform CRD %v", err)
		return
	}

	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "iseclcontroller")

	indexer, informer := crdController.NewIseclHAIndexerInformer(config, queue, crdmutex, TrustedPrefixConf)

	controller := crdController.NewIseclHAController(queue, indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	glog.V(4).Infof("Waiting for updates on  ISecl Custom Resource Definitions")

	// Wait forever
	select {}
}

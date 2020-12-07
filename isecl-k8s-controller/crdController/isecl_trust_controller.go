/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdController

import (
	"fmt"
	"intel/isecl/k8s-custom-controller/v3/crdLabelAnnotate"
	commLog "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log"
	ha_schema "intel/isecl/k8s-custom-controller/v3/crdSchema/api/hostattribute/v1beta1"
	ha_client "intel/isecl/k8s-custom-controller/v3/crdSchema/client/clientset/versioned/typed/hostattribute/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	runtime2 "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var trustLabelRegex = regexp.MustCompile("(^[a-zA-Z0-9_///.-]*$)")
var TaintUntrustedNodes = false

const (
	// lenSGXLabels is the number of SGX Features that are currently supported per node
	lenSGXLabels = 5
	// lenTrustLabels is the number of mandatory ISecL labels that are required per node
	lenTrustLabels = 2
)

type IseclHAController struct {
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface
}

type Config struct {
	Trusted string `json:"trusted"`
}

var defaultLog = commLog.GetDefaultLogger()

func NewIseclHAController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *IseclHAController {
	return &IseclHAController{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func GetHACrdDef() CrdDefinition {
	return CrdDefinition{
		Plural:   ha_schema.HAPlural,
		Singular: ha_schema.HASingular,
		Group:    ha_schema.HAGroup,
		Kind:     ha_schema.HAKind,
	}
}

func (c *IseclHAController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two CRD with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncFromQueue(key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

//processPLQueue : can be extended to validate the crd objects are been acted upon
func (c *IseclHAController) processPLQueue(key string) error {
	defaultLog.Infof("processPLQueue for Key %#v ", key)
	return nil
}

// syncFromQueue is the business defaultLogic of the controller. In this controller it simply prints
// information about the CRD to stdout. In case an error happened, it has to simply return the error.
// The retry defaultLogic should not be part of the business logic.
func (c *IseclHAController) syncFromQueue(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		defaultLog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a CDR, so that we will see a delete for one CRD
		defaultLog.Infof("PL CRD object %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a CRD object was recreated with the same name
		defaultLog.Infof("Sync/Add/Update for PL CRD Object %#v ", obj)
		err = c.processPLQueue(key)
		if err != nil {
			defaultLog.Fatalf("Error while processing queue %v", err)
		}
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *IseclHAController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		defaultLog.Infof("Error syncing CRD %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	defaultLog.Infof("Dropping CRD %q out of the queue: %v", key, err)
}

func (c *IseclHAController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	defaultLog.Info("Starting ISeclHAController")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	defaultLog.Info("Stopping Platform controller")
}

func (c *IseclHAController) runWorker() {
	for c.processNextItem() {
	}
}

//GetHaObjLabel creates labels and annotations map based on HA CRD
func GetHaObjLabel(obj ha_schema.Host, node *corev1.Node, tagPrefix string) (crdLabelAnnotate.Labels, crdLabelAnnotate.Annotations, error) {
	assetTagSize := len(obj.AssetTag)
	hwFeaturesSize := len(obj.HardwareFeatures)

	// allocate labels for:
	// ISecL Trust Status
	// ISecL Asset Tags
	// ISecL Hardware Features
	// SGX Labels
	var lbl = make(crdLabelAnnotate.Labels, lenTrustLabels+assetTagSize+hwFeaturesSize+lenSGXLabels)

	// we need to allocate separate annotations for the SignedTrustReport from iSecL iHub and SGX iHubs
	var annotation = make(crdLabelAnnotate.Annotations, 2)

	if obj.HvsSignedTrustReport != "" {
		annotation[hvsSignTrustReport] = obj.HvsSignedTrustReport

		expiry := strings.Replace(obj.HvsTrustExpiry.Format(time.RFC3339), ":", ".", -1)
		lbl[hvsTrustExpiry] = expiry

		trustLabelWithPrefix := tagPrefix + trustlabel
		lbl[trustLabelWithPrefix] = strconv.FormatBool(obj.Trusted)

		for key, val := range obj.AssetTag {
			labelkey := tagPrefix + key
			lbl[labelkey] = val
		}

		for key, val := range obj.HardwareFeatures {
			labelkey := tagPrefix + key
			lbl[labelkey] = val
		}

		//Remove the older asset tags/ hardware features in node labels
		for key, _ := range node.Labels {
			if 	_, ok := lbl[key]; !ok && strings.Contains(key, tagPrefix){
				delete(node.Labels, key)
			}
		}
	}
	if obj.SgxSignedTrustReport != "" {
		annotation[sgxSignTrustReport] = obj.SgxSignedTrustReport

		expiry := strings.Replace(obj.SgxTrustExpiry.Format(time.RFC3339), ":", ".", -1)

		lbl[sgxTrustExpiry] = expiry
		lbl[sgxEnable] = obj.SgxEnabled
		lbl[sgxSupported] = obj.SGXSupported
		lbl[flcEnabled] = obj.FLCEnabled
		lbl[tcbUpToDate] = obj.TCBUpToDate
		lbl[epcMemory] = obj.EPCSize
	}

	return lbl, annotation, nil
}

//AddHostAttributesTabObj Handler for addition event of the HA CRD
func AddHostAttributesTabObj(haobj *ha_schema.HostAttributesCrd, helper crdLabelAnnotate.APIHelpers, cli *k8sclient.Clientset, mutex *sync.Mutex, tagPrefix string) {

	for index, ele := range haobj.Spec.HostList {
		nodeName := haobj.Spec.HostList[index].Hostname
		node, err := helper.GetNode(cli, nodeName)
		if err != nil {
			defaultLog.Infof("Failed to get node within cluster: %s", err.Error())
			continue
		}
		lbl, ann, err := GetHaObjLabel(ele, node, tagPrefix)
		if err != nil {
			defaultLog.Fatalf("Error: %v", err)
		}
		mutex.Lock()
		helper.AddLabelsAnnotations(node, lbl, ann, tagPrefix)
		// NoExec Taints on nodes enforced optionally
		if TaintUntrustedNodes {
			if !ele.Trusted {
				// Taint the node with no execute
				if err = helper.AddTaint(node, "untrusted", "true", "NoExecute"); err != nil {
					defaultLog.Errorf("Unable to add taints: %s", err.Error())
				}
			} else {
				//Remove Taint from node with no execute
				if err = helper.DeleteTaint(node, "untrusted", "true", "NoExecute"); err != nil {
					defaultLog.Errorf("Unable to delete taints: %s", err.Error())
				}
			}
		}

		err = helper.UpdateNode(cli, node)
		if err != nil {
			defaultLog.Infof("can't update node: %s", err.Error())
		}
		mutex.Unlock()
	}
}

//NewIseclHAIndexerInformer returns informer for HA CRD object
func NewIseclHAIndexerInformer(config *rest.Config, queue workqueue.RateLimitingInterface, crdMutex *sync.Mutex, tagPrefix string) (cache.Indexer, cache.Controller) {
	// Create a new clientset which include our CRD schema
	hacrdclient, err := ha_client.NewForConfig(config)
	if err != nil {
		defaultLog.Fatalf("Failed to create new clientset for Platform CRD %v", err)
	}

	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime2.Object, error) {
			// list all of the host attributes in the default namespace
			return hacrdclient.HostAttributesCrds(metav1.NamespaceDefault).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			// watch all of the host attributes in the default namespace
			return hacrdclient.HostAttributesCrds(metav1.NamespaceDefault).Watch(options)
		},
	}

	//Create a PL CRD Helper object
	hInf, cli := crdLabelAnnotate.Getk8sClientHelper(config)

	return cache.NewIndexerInformer(listWatch, &ha_schema.HostAttributesCrd{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			defaultLog.Info("Received Add event for ", key)
			haobj := obj.(*ha_schema.HostAttributesCrd)
			if err == nil {
				queue.Add(key)
			}
			AddHostAttributesTabObj(haobj, hInf, cli, crdMutex, tagPrefix)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			defaultLog.Info("Received Update event for ", key)
			haobj := new.(*ha_schema.HostAttributesCrd)
			if err == nil {
				queue.Add(key)
			}
			AddHostAttributesTabObj(haobj, hInf, cli, crdMutex, tagPrefix)
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			defaultLog.Info("Received delete event for ", key)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})
}

/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdController

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s_custom_cit_controllers-k8s_custom_controllers/crdLabelAnnotate"
	ha_schema "k8s_custom_cit_controllers-k8s_custom_controllers/crdSchema/iseclHostAttributesSchema"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var StringReg = regexp.MustCompile("(^[a-zA-Z0-9_///.-]*$)")

const MAX_BYTES_LEN = 200

type IseclHAController struct {
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface
}

type Config struct {
	Trusted string `"json":"trusted"`
}

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
	Log.Infof("processPLQueue for Key %#v ", key)
	return nil
}

// syncFromQueue is the business logic of the controller. In this controller it simply prints
// information about the CRD to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *IseclHAController) syncFromQueue(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		Log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a CDR, so that we will see a delete for one CRD
		Log.Infof("PL CRD object %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a CRD object was recreated with the same name
		Log.Infof("Sync/Add/Update for PL CRD Object %#v ", obj)
		c.processPLQueue(key)
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
		Log.Infof("Error syncing CRD %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	Log.Infof("Dropping CRD %q out of the queue: %v", key, err)
}

func (c *IseclHAController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	Log.Info("Starting Platformcrd controller")

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
	Log.Info("Stopping Platform controller")
}

func (c *IseclHAController) runWorker() {
	for c.processNextItem() {
	}
}

//GetHaObjLabel creates lables and annotations map based on HA CRD
func GetHaObjLabel(obj ha_schema.Host, node *api.Node, trustedPrefixConf string) (crdLabelAnnotate.Labels, crdLabelAnnotate.Annotations, error) {
	assetTagsize := len(obj.Assettag)

	var lbl = make(crdLabelAnnotate.Labels, assetTagsize+2)
	var annotation = make(crdLabelAnnotate.Annotations, 1)
	trustPresent := false
	trustLabelWithPrefix, err := getPrefixFromConf(trustedPrefixConf)

	if err != nil {
		return nil, nil, err
	}

	if !StringReg.MatchString(trustLabelWithPrefix) {
		return nil, nil, errors.New("Invalid string formatted input")
	}

	trustLabelWithPrefix = trustLabelWithPrefix + trustlabel

	for key, val := range obj.Assettag {
		labelkey := strings.Replace(key, " ", ".", -1)
		labelkey = strings.Replace(labelkey, ":", ".", -1)
		labelkey = trustLabelWithPrefix + labelkey
		lbl[labelkey] = val
	}

	//Comparing with existing node labels
	for key, value := range node.Labels {
		if key == trustLabelWithPrefix {
			trustPresent = true
			if value == obj.Trusted {
				Log.Info("No change in Trustlabel, updating Trustexpiry time only")
			} else {
				Log.Info("Updating Complete Trustlabel for the node")
				lbl[trustLabelWithPrefix] = obj.Trusted
			}
		}
	}
	if !trustPresent {
		Log.Info("Trust value was not present on node adding for first time")
		lbl[trustLabelWithPrefix] = obj.Trusted
	}
	expiry := strings.Replace(obj.Expiry, ":", ".", -1)
	lbl[trustexpiry] = expiry
	annotation[trustsignreport] = obj.SignedReport

	return lbl, annotation, nil
}

func getPrefixFromConf(path string) (string, error) {
        out, err := os.Open(path)
        if err != nil {
                Log.Errorf("Error: %s %v", path, err)
                return "", err
        }

        defer out.Close()
        readBytes := make([]byte, MAX_BYTES_LEN)
        n, err := out.Read(readBytes)
        if err != nil {
                return "", err
        }
        s := Config{}
        err = json.Unmarshal(readBytes[:n], &s)
        if err != nil {
                Log.Errorf("Error:  %v", err)
                return "", err
        }
        return s.Trusted, nil
}

//AddHostAttributesTabObj Handler for addition event of the HA CRD
func AddHostAttributesTabObj(haobj *ha_schema.HostAttributesCrd, helper crdLabelAnnotate.APIHelpers, cli *k8sclient.Clientset, mutex *sync.Mutex, trustedPrefixConf string) {
	trustLabelWithPrefix, err := getPrefixFromConf(trustedPrefixConf)
	Log.Errorf("Could not get the trustlabel prefix %v", err)

	for index, ele := range haobj.Spec.HostList {
		nodeName := haobj.Spec.HostList[index].Hostname
		node, err := helper.GetNode(cli, nodeName)
		if err != nil {
			Log.Info("Failed to get node within cluster: %s", err.Error())
			continue
		}
		lbl, ann, err := GetHaObjLabel(ele, node, trustedPrefixConf)
		if err != nil {
			Log.Fatalf("Error: %v", err)
		}
		mutex.Lock()
		helper.AddLabelsAnnotations(node, lbl, ann, trustLabelWithPrefix)
		err = helper.UpdateNode(cli, node)
		mutex.Unlock()
		if err != nil {
			Log.Info("can't update node: %s", err.Error())
		}
	}
}

//NewIseclHAIndexerInformer returns informer for HA CRD object
func NewIseclHAIndexerInformer(config *rest.Config, queue workqueue.RateLimitingInterface, crdMutex *sync.Mutex, trustedPrefixConf string) (cache.Indexer, cache.Controller) {
	// Create a new clientset which include our CRD schema
	crdcs, scheme, err := ha_schema.NewHAClient(config)
	if err != nil {
		Log.Fatalf("Failed to create new clientset for Platform CRD %v", err)
	}

	// Create a CRD client interface
	hacrdclient := ha_schema.HAClient(crdcs, scheme, "default")

	//Create a PL CRD Helper object
	hInf, cli := crdLabelAnnotate.Getk8sClientHelper(config)

	return cache.NewIndexerInformer(hacrdclient.NewHAListWatch(), &ha_schema.HostAttributesCrd{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			Log.Info("Received Add event for ", key)
			haobj := obj.(*ha_schema.HostAttributesCrd)
			if err == nil {
				queue.Add(key)
			}
			AddHostAttributesTabObj(haobj, hInf, cli, crdMutex, trustedPrefixConf)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			Log.Info("Received Update event for ", key)
			haobj := new.(*ha_schema.HostAttributesCrd)
			if err == nil {
				queue.Add(key)
			}
			AddHostAttributesTabObj(haobj, hInf, cli, crdMutex, trustedPrefixConf)
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			Log.Info("Received delete event for ", key)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})
}

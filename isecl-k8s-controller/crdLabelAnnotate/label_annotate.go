/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdLabelAnnotate

import (
	"k8s_custom_cit_controllers-k8s_custom_controllers/util"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

var Log = util.GetLogger()
type APIHelpers interface {

	// GetNode returns the Kubernetes node on which this container is running.
	GetNode(*k8sclient.Clientset, string) (*api.Node, error)

	// AddLabelsAnnotations modifies the supplied node's labels and annotations collection.
	// In order to publish the labels, the node must be subsequently updated via the
	// API server using the client library.
	AddLabelsAnnotations(*api.Node, Labels, Annotations, string)

	// UpdateNode updates the node via the API server using a client.
	UpdateNode(*k8sclient.Clientset, *api.Node) error
}

// Implements main.APIHelpers
type K8sHelpers struct{}
type Labels map[string]string
type Annotations map[string]string

//Getk8sClientHelper returns helper object and clientset to fetch node
func Getk8sClientHelper(config *rest.Config) (APIHelpers, *k8sclient.Clientset) {
	helper := APIHelpers(K8sHelpers{})

	cli, err := k8sclient.NewForConfig(config)
	if err != nil {
		Log.Errorf("Error while creating k8s client %v", err)
	}
	return helper, cli
}

//GetNode returns node API based on nodename
func (h K8sHelpers) GetNode(cli *k8sclient.Clientset, NodeName string) (*api.Node, error) {
	// Get the node object using the node name
	node, err := cli.Core().Nodes().Get(NodeName, metav1.GetOptions{})
	if err != nil {
		Log.Errorf("Can't get node: %s", err.Error())
		return nil, err
	}

	return node, nil
}


func cleanupLabelsWithIsecl(n *api.Node, labelPrefix string) Labels{
	var newNodeLabels = make(Labels, len(n.Labels))
	for k, v := range n.Labels{
		if strings.HasPrefix(k, labelPrefix) {
			continue
		}
		newNodeLabels[k] = v
	}
	return newNodeLabels
}

//AddLabelsAnnotations applies labels and annotations to the node
func (h K8sHelpers) AddLabelsAnnotations(n *api.Node, labels Labels, annotations Annotations, labelPrefix string) {
	//Clean up labels with isecl prefix.
	newNodeLabels := cleanupLabelsWithIsecl(n, labelPrefix)
	for k, v := range labels {
		newNodeLabels[k] = v
	}
	for k, v := range annotations {
		n.Annotations[k] = v
	}
	n.Labels = newNodeLabels
	Log.Info(newNodeLabels)
}

//UpdateNode updates the node API
func (h K8sHelpers) UpdateNode(c *k8sclient.Clientset, n *api.Node) error {
	// Send the updated node to the apiserver.
	_, err := c.Core().Nodes().Update(n)
	if err != nil {
		Log.Errorf("Error while updating node label:", err.Error())
		return err
	}
	return nil
}

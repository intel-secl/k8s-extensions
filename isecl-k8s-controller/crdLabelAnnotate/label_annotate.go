/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdLabelAnnotate

import (
	"strings"

	commLog "github.com/intel-secl/intel-secl/v3/pkg/lib/common/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var defaultLog = commLog.GetDefaultLogger()

type APIHelpers interface {

	// GetNode returns the Kubernetes node on which this container is running.
	GetNode(*k8sclient.Clientset, string) (*corev1.Node, error)

	// AddLabelsAnnotations modifies the supplied node's labels and annotations collection.
	// In order to publish the labels, the node must be subsequently updated via the
	// API server using the client library.
	AddLabelsAnnotations(*corev1.Node, Labels, Annotations, string)

	// UpdateNode updates the node via the API server using a client.
	UpdateNode(*k8sclient.Clientset, *corev1.Node) error

	// DeleteNode deletes the node name via the API server using a client.
	DeleteNode(*k8sclient.Clientset, string) error

	// AddTaint modifies the supplied node's taints to add an additional taint
	// effect should be one of: NoSchedule, PreferNoSchedule, NoExecute
	AddTaint(n *corev1.Node, key string, value string, effect string) error

	// DeleteTaint modifies the supplied node's taints to delete a specified taint
	// effect should be one of: NoSchedule, PreferNoSchedule, NoExecute
	DeleteTaint(n *corev1.Node, key string, value string, effect string) error
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
		defaultLog.Errorf("Error while creating k8s client %v", err)
	}
	return helper, cli
}

//GetNode returns node API based on nodename
func (h K8sHelpers) GetNode(cli *k8sclient.Clientset, NodeName string) (*corev1.Node, error) {
	// Get the node object using the node name
	node, err := cli.CoreV1().Nodes().Get(NodeName, metav1.GetOptions{})
	if err != nil {
		defaultLog.Errorf("Can't get node: %s", err.Error())
		return nil, err
	}

	return node, nil
}

func cleanupLabelsWithIsecl(n *corev1.Node, labelPrefix string) Labels {
	var newNodeLabels = make(Labels, len(n.Labels))
	iseclTrustedTag := labelPrefix + "trusted"
	for k, v := range n.Labels {
		if strings.HasPrefix(k, labelPrefix) && k != iseclTrustedTag {
			continue
		}
		newNodeLabels[k] = v
	}
	return newNodeLabels
}

//AddLabelsAnnotations applies labels and annotations to the node
func (h K8sHelpers) AddLabelsAnnotations(n *corev1.Node, labels Labels, annotations Annotations, labelPrefix string) {
	//Clean up labels with isecl prefix.
	newNodeLabels := cleanupLabelsWithIsecl(n, labelPrefix)
	for k, v := range labels {
		newNodeLabels[k] = v
	}
	for k, v := range annotations {
		n.Annotations[k] = v
	}
	n.Labels = newNodeLabels
	defaultLog.Info(newNodeLabels)
}

//AddTaint applies labels and annotations to the node
//effect should be one of: NoSchedule, PreferNoSchedule, NoExecute
func (h K8sHelpers) AddTaint(n *corev1.Node, key string, value string, effect string) error {
	taintEffect, ok := map[string]corev1.TaintEffect{
		"NoSchedule":       corev1.TaintEffectNoSchedule,
		"PreferNoSchedule": corev1.TaintEffectPreferNoSchedule,
		"NoExecute":        corev1.TaintEffectNoExecute,
	}[effect]

	if !ok {
		return errors.Errorf("Taint effect %v not valid", effect)
	}

	n.Spec.Taints = append(n.Spec.Taints, corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: taintEffect,
	})

	return nil
}

//DeleteTaint removes the taint from the node
//effect should be one of: NoSchedule, PreferNoSchedule, NoExecute
func (h K8sHelpers) DeleteTaint(n *corev1.Node, key string, value string, effect string) error {
	taintEffect, ok := map[string]corev1.TaintEffect{
		"NoSchedule":       corev1.TaintEffectNoSchedule,
		"PreferNoSchedule": corev1.TaintEffectPreferNoSchedule,
		"NoExecute":        corev1.TaintEffectNoExecute,
	}[effect]

	if !ok {
		return errors.Errorf("Taint effect %v not valid", effect)
	}

	delT := corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: taintEffect,
	}

	// loop over the taints present on the node appending to a new list
	// skipping the one we don't want
	var newTaints []corev1.Taint
	for _, t := range n.Spec.Taints {
		if t.Key != delT.Key && t.Value != delT.Value && t.Effect != delT.Effect {
			newTaints = append(newTaints, t)
		} else {
			defaultLog.Infof("Dropped %s taint from node %v", effect, n)
		}
	}

	// assign new list of taints back to node
	n.Spec.Taints = newTaints

	return nil
}

//UpdateNode updates the node API
func (h K8sHelpers) UpdateNode(c *k8sclient.Clientset, n *corev1.Node) error {
	// Send the updated node to the apiserver.
	_, err := c.CoreV1().Nodes().Update(n)
	if err != nil {
		defaultLog.Errorf("Error while updating node label:", err.Error())
		return err
	}
	return nil
}

//DeleteNode updates the node API
func (h K8sHelpers) DeleteNode(c *k8sclient.Clientset, nodeName string) error {
	// Send the deleted node to the apiserver.
	err := c.CoreV1().Nodes().Delete(nodeName, &metav1.DeleteOptions{})

	// Node already deleted
	if k8serrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		defaultLog.Errorf("Error while deleting node label:", err.Error())
		return err
	}
	return nil
}

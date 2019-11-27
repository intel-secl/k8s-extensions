/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package algorithm

import (
	"fmt"
	"k8s_scheduler_cit_extension-k8s_extended_scheduler/util"

	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
)

var Log = util.GetLogger()

//FilteredHost is used for getting the nodes and pod details and verify and return if pod key matches with annotations
func FilteredHost(args *schedulerapi.ExtenderArgs, trustPrefix string) (*schedulerapi.ExtenderFilterResult, error) {
	result := []v1.Node{}
	failedNodesMap := schedulerapi.FailedNodesMap{}

	//Get the list of nodes and pods from base scheduler
	nodes := args.Nodes
	pod := args.Pod
	confTrustPrefix := trustPrefix
	//Check for presence of Affinity tag in pod specification
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil && len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {

			//get the nodeselector data
			nodeSelectorData := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms

			for _, node := range nodes.Items {
				//always check for the trust tag signed report
				if cipherVal, ok := node.Annotations["TrustTagSignedReport"]; ok {
					for _, nodeSelector := range nodeSelectorData {
						//match the data from the pod node selector tag to the node annotation
						Log.Infof("Checking annotation for node %s",node)
						if CheckAnnotationAttrib(cipherVal, nodeSelector.MatchExpressions, confTrustPrefix) {
							result = append(result, node)
						} else {
							failedNodesMap[node.Name] = fmt.Sprintf("Annotation validation failed in extended-scheduler")
						}
					}
				}
			}
		} else {
			for _, node := range nodes.Items {
				result = append(result, node)
			}
		}
	} else {
		for _, node := range nodes.Items {
			result = append(result, node)
		}
	}

	Log.Infof("Returning following nodelist from extended scheduler: %v", result)
	if len(result) != 0 {
		return &schedulerapi.ExtenderFilterResult{
			Nodes:       &v1.NodeList{Items: result},
			NodeNames:   nil,
			FailedNodes: failedNodesMap,
		}, nil
	} else {
		return nil, fmt.Errorf("Node validation failed at extended scheduler")
	}
}

/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package algorithm

import (
	"fmt"
	commLog "github.com/intel-secl/intel-secl/v4/pkg/lib/common/log"
	"github.com/intel-secl/k8s-extensions/v4/isecl-k8s-scheduler/constants"

	v1 "k8s.io/api/core/v1"
	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

var defaultLog = commLog.GetDefaultLogger()

//FilteredHost is used for getting the nodes and pod details and verify and return if pod key matches with annotations
func FilteredHost(args *schedulerapi.ExtenderArgs, iHubPubKeys map[string][]byte, tagPrefix string) (*schedulerapi.ExtenderFilterResult, error) {
	result := []v1.Node{}
	failedNodesMap := schedulerapi.FailedNodesMap{}

	//Get the list of nodes and pods from base scheduler
	nodes := args.Nodes
	pod := args.Pod
	//Check for presence of Affinity tag in pod specification
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil && len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions) > 0 {
			//get the nodeselector data
			nodeSelectorData := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			for _, node := range nodes.Items {
				hvsSignedTrustReportExists := false
				sgxSignedTrustReportExists := false
				hvsSignedTrustReportValidated := false
				sgxSignedTrustReportValidated := false
				for _, nodeSelector := range nodeSelectorData {
					if cipherVal, ok := node.Annotations[constants.HvsSignedTrustReport]; ok {
						//match the data from the pod node selector tag to the node annotation HvsSignedTrustReport
						defaultLog.Infof("Checking HVS Trust annotation for node %v", node)
						hvsSignedTrustReportExists = true
						if CheckAnnotationAttrib(cipherVal, nodeSelector.MatchExpressions, iHubPubKeys, tagPrefix, constants.HVSAttestation) {
							hvsSignedTrustReportValidated = true
						} else {
							failedNodesMap[node.Name] = fmt.Sprintf("ISecL Trust Annotation validation failed in extended-scheduler")
						}
					}
					if cipherVal, ok := node.Annotations[constants.SgxSignedTrustReport]; ok {
						//match the data from the pod node selector tag to the node annotation SgxSignedTrustReport
						defaultLog.Infof("Checking SGX Trust annotation for node %v", node)
						sgxSignedTrustReportExists = true
						if CheckAnnotationAttrib(cipherVal, nodeSelector.MatchExpressions, iHubPubKeys, tagPrefix, constants.SGXAttestation) {
							sgxSignedTrustReportValidated = true
						} else {
							failedNodesMap[node.Name] = fmt.Sprintf("SGX Trust Annotation validation failed in extended-scheduler")
						}
					}
				}
				if hvsSignedTrustReportExists && sgxSignedTrustReportExists {
					if hvsSignedTrustReportValidated && sgxSignedTrustReportValidated {
						result = append(result, node)
					}
				} else if hvsSignedTrustReportExists {
					if hvsSignedTrustReportValidated {
						result = append(result, node)
					}
				} else if sgxSignedTrustReportExists {
					if sgxSignedTrustReportValidated {
						result = append(result, node)
					}
				} else {
					//If there is no HvsSignedTrustReport or HvsSignedTrustReport on Node then return the node.
					result = append(result, node)
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

	defaultLog.Infof("Returning following nodelist from extended scheduler: %v", result)
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

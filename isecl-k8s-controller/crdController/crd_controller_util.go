/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdController

import (
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	hvsTrustExpiry      = "HvsTrustExpiry"
	sgxTrustExpiry      = "SgxTrustExpiry"
	trustlabel          = "trusted"
	hvsSignTrustReport  = "HvsSignedTrustReport"
	sgxSignTrustReport  = "SgxSignedTrustReport"
	sgxEnable           = "SGX-Enabled"
	sgxSupported        = "SGX-Supported"
	flcEnabled          = "FLC-Enabled"
	tcbUpToDate         = "TCBUpToDate"
	epcMemory           = "EPC-Memory"
)

type CrdDefinition struct {
	Plural   string
	Singular string
	Group    string
	Kind     string
}

//NewIseclCustomResourceDefinition Creates new ISecL CRD's
func NewIseclCustomResourceDefinition(cs clientset.Interface, crdDef *CrdDefinition) error {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: crdDef.Plural + "." + crdDef.Group},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   crdDef.Group,
			Version: "v1beta1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   crdDef.Plural,
				Singular: crdDef.Singular,
				Kind:     crdDef.Kind,
			},
			Scope: apiextensionsv1beta1.NamespaceScoped,
		},
	}
	_, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil && apierrors.IsAlreadyExists(err) {
		defaultLog.Infof("ISECL HostAttributes CRD object already exists")
		return nil
	} else {
		if err := waitForEstablishedCRD(cs, crd.Name); err != nil {
			defaultLog.Errorf("Failed to establish CRD %v", err)
			return err
		}
	}

	defaultLog.Infof("Successfully created CRD : %#v \n", crd.Name)
	return err
}

//waitForEstablishedCRD polls until the CRD gets created and ready for use
func waitForEstablishedCRD(client clientset.Interface, name string) error {
	return wait.PollImmediate(500*time.Millisecond, wait.ForeverTestTimeout, func() (bool, error) {
		crd, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					defaultLog.Infof("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, nil
	})
}

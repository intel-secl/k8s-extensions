/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package v1beta1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	HAPlural   string = "hostattributes"
	HASingular string = "hostattribute"
	HAKind     string = "HostAttributesCrd"
	HAGroup    string = "crd.isecl.intel.com"
	HAVersion  string = "v1beta1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type HostAttributesCrd struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               Spec `json:"spec"`
}

type Host struct {
	Hostname     string            `json:"hostName"`
	Trusted      bool              `json:"trusted,omitempty"`
	Expiry       string            `json:"validTo,omitempty"`
	SignedReport string            `json:"signedTrustReport"`
	Assettag     map[string]string `json:"assetTags,omitempty"`
	SgxEnabled   string            `json:"sgx-enabled,omitempty"`
	SGXSupported string            `json:"sgx-supported,omitempty"`
	TCBUpToDate  string            `json:"tcbUpToDate,omitempty"`
	EPCSize      string            `json:"epc-size,omitempty"`
	FLCEnabled   string            `json:"flc-enabled,omitempty"`
}

type Spec struct {
	HostList []Host `json:"hostList"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type HostAttributesCrdList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []HostAttributesCrd `json:"items"`
}

/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package crdHostAttributesSchema

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	HAPlural   string = "hostattributes"
	HASingular string = "crd"
	HAKind     string = "HostAttributesCrd"
	HAGroup    string = "crd.isecl.intel.com"
	HAVersion  string = "v1beta1"
)

//HAClient returns CRD clientset required to apply watch on the CRD
func HAClient(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *haclient {
	return &haclient{cl: cl, ns: namespace, plural: HAPlural,
		codec: runtime.NewParameterCodec(scheme)}
}

type haclient struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

type HostAttributesCrd struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               Spec `json:"spec"`
}

type Host struct {
	Hostname     string            `json:"hostName"`
	Trusted      string            `json:"trusted"`
	Expiry       string            `json:"validTo"`
	SignedReport string            `json:"signedTrustReport"`
	Assettag     map[string]string `json:"assetTags"`
}

type Spec struct {
	HostList []Host `json:"hostList"`
}

type HostAttributesCrdList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []HostAttributesCrd `json:"items"`
}

// Create a  Rest client with the new CRD Schema
var SchemeGroupVersion = schema.GroupVersion{Group: HAGroup, Version: HAVersion}

//addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&HostAttributesCrd{},
		&HostAttributesCrdList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

//NewPLClient registers CRD schema and returns rest client for the CRD
func NewHAClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}
	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}
	return client, scheme, nil
}

// Create a new List watch for our HA CRD
func (f *haclient) NewHAListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}

/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	"time"

	v1beta1 "github.com/intel-secl/k8s-extensions/v4/isecl-k8s-controller/crdSchema/api/hostattribute/v1beta1"
	scheme "github.com/intel-secl/k8s-extensions/v4/isecl-k8s-controller/crdSchema/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// HostAttributesCrdsGetter has a method to return a HostAttributesCrdInterface.
// A group's client should implement this interface.
type HostAttributesCrdsGetter interface {
	HostAttributesCrds(namespace string) HostAttributesCrdInterface
}

// HostAttributesCrdInterface has methods to work with HostAttributesCrd resources.
type HostAttributesCrdInterface interface {
	Create(*v1beta1.HostAttributesCrd) (*v1beta1.HostAttributesCrd, error)
	Update(*v1beta1.HostAttributesCrd) (*v1beta1.HostAttributesCrd, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.HostAttributesCrd, error)
	List(opts v1.ListOptions) (*v1beta1.HostAttributesCrdList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.HostAttributesCrd, err error)
	HostAttributesCrdExpansion
}

// hostAttributesCrds implements HostAttributesCrdInterface
type hostAttributesCrds struct {
	client rest.Interface
	ns     string
}

// newHostAttributesCrds returns a HostAttributesCrds
func newHostAttributesCrds(c *CrdV1beta1Client, namespace string) *hostAttributesCrds {
	return &hostAttributesCrds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the hostAttributesCrd, and returns the corresponding hostAttributesCrd object, and an error if there is any.
func (c *hostAttributesCrds) Get(name string, options v1.GetOptions) (result *v1beta1.HostAttributesCrd, err error) {
	result = &v1beta1.HostAttributesCrd{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("hostattributes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(context.Background()).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of HostAttributesCrds that match those selectors.
func (c *hostAttributesCrds) List(opts v1.ListOptions) (result *v1beta1.HostAttributesCrdList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.HostAttributesCrdList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("hostattributes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(context.Background()).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested hostAttributesCrds.
func (c *hostAttributesCrds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("hostattributes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(context.Background())
}

// Create takes the representation of a hostAttributesCrd and creates it.  Returns the server's representation of the hostAttributesCrd, and an error, if there is any.
func (c *hostAttributesCrds) Create(hostAttributesCrd *v1beta1.HostAttributesCrd) (result *v1beta1.HostAttributesCrd, err error) {
	result = &v1beta1.HostAttributesCrd{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("hostattributes").
		Body(hostAttributesCrd).
		Do(context.Background()).
		Into(result)
	return
}

// Update takes the representation of a hostAttributesCrd and updates it. Returns the server's representation of the hostAttributesCrd, and an error, if there is any.
func (c *hostAttributesCrds) Update(hostAttributesCrd *v1beta1.HostAttributesCrd) (result *v1beta1.HostAttributesCrd, err error) {
	result = &v1beta1.HostAttributesCrd{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("hostattributes").
		Name(hostAttributesCrd.Name).
		Body(hostAttributesCrd).
		Do(context.Background()).
		Into(result)
	return
}

// Delete takes name of the hostAttributesCrd and deletes it. Returns an error if one occurs.
func (c *hostAttributesCrds) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("hostattributes").
		Name(name).
		Body(options).
		Do(context.Background()).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *hostAttributesCrds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("hostattributes").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do(context.Background()).
		Error()
}

// Patch applies the patch and returns the patched hostAttributesCrd.
func (c *hostAttributesCrds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.HostAttributesCrd, err error) {
	result = &v1beta1.HostAttributesCrd{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("hostattributes").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do(context.Background()).
		Into(result)
	return
}

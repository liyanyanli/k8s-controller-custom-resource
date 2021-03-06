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

package fake

import (
	democrdv1 "github.com/liyanyanli/k8s-controller-custom-resource/pkg/apis/democrd/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDemocrds implements DemocrdInterface
type FakeDemocrds struct {
	Fake *FakeDemocrdV1
	ns   string
}

var democrdsResource = schema.GroupVersionResource{Group: "democrd.k8s.io", Version: "v1", Resource: "democrds"}

var democrdsKind = schema.GroupVersionKind{Group: "democrd.k8s.io", Version: "v1", Kind: "Democrd"}

// Get takes name of the democrd, and returns the corresponding democrd object, and an error if there is any.
func (c *FakeDemocrds) Get(name string, options v1.GetOptions) (result *democrdv1.Democrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(democrdsResource, c.ns, name), &democrdv1.Democrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*democrdv1.Democrd), err
}

// List takes label and field selectors, and returns the list of Democrds that match those selectors.
func (c *FakeDemocrds) List(opts v1.ListOptions) (result *democrdv1.DemocrdList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(democrdsResource, democrdsKind, c.ns, opts), &democrdv1.DemocrdList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &democrdv1.DemocrdList{ListMeta: obj.(*democrdv1.DemocrdList).ListMeta}
	for _, item := range obj.(*democrdv1.DemocrdList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested democrds.
func (c *FakeDemocrds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(democrdsResource, c.ns, opts))

}

// Create takes the representation of a democrd and creates it.  Returns the server's representation of the democrd, and an error, if there is any.
func (c *FakeDemocrds) Create(democrd *democrdv1.Democrd) (result *democrdv1.Democrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(democrdsResource, c.ns, democrd), &democrdv1.Democrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*democrdv1.Democrd), err
}

// Update takes the representation of a democrd and updates it. Returns the server's representation of the democrd, and an error, if there is any.
func (c *FakeDemocrds) Update(democrd *democrdv1.Democrd) (result *democrdv1.Democrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(democrdsResource, c.ns, democrd), &democrdv1.Democrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*democrdv1.Democrd), err
}

// Delete takes name of the democrd and deletes it. Returns an error if one occurs.
func (c *FakeDemocrds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(democrdsResource, c.ns, name), &democrdv1.Democrd{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDemocrds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(democrdsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &democrdv1.DemocrdList{})
	return err
}

// Patch applies the patch and returns the patched democrd.
func (c *FakeDemocrds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *democrdv1.Democrd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(democrdsResource, c.ns, name, pt, data, subresources...), &democrdv1.Democrd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*democrdv1.Democrd), err
}

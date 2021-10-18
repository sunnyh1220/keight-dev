/*
Copyright 2021 sunnyh.

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
	"context"

	v1alpha1 "github.com/sunnyh1220/keight-dev/controller-sample/pkg/apis/stable/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCronTabs implements CronTabInterface
type FakeCronTabs struct {
	Fake *FakeStableV1alpha1
	ns   string
}

var crontabsResource = schema.GroupVersionResource{Group: "stable.sunnyh.easy", Version: "v1alpha1", Resource: "crontabs"}

var crontabsKind = schema.GroupVersionKind{Group: "stable.sunnyh.easy", Version: "v1alpha1", Kind: "CronTab"}

// Get takes name of the cronTab, and returns the corresponding cronTab object, and an error if there is any.
func (c *FakeCronTabs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.CronTab, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(crontabsResource, c.ns, name), &v1alpha1.CronTab{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CronTab), err
}

// List takes label and field selectors, and returns the list of CronTabs that match those selectors.
func (c *FakeCronTabs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.CronTabList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(crontabsResource, crontabsKind, c.ns, opts), &v1alpha1.CronTabList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CronTabList{ListMeta: obj.(*v1alpha1.CronTabList).ListMeta}
	for _, item := range obj.(*v1alpha1.CronTabList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cronTabs.
func (c *FakeCronTabs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(crontabsResource, c.ns, opts))

}

// Create takes the representation of a cronTab and creates it.  Returns the server's representation of the cronTab, and an error, if there is any.
func (c *FakeCronTabs) Create(ctx context.Context, cronTab *v1alpha1.CronTab, opts v1.CreateOptions) (result *v1alpha1.CronTab, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(crontabsResource, c.ns, cronTab), &v1alpha1.CronTab{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CronTab), err
}

// Update takes the representation of a cronTab and updates it. Returns the server's representation of the cronTab, and an error, if there is any.
func (c *FakeCronTabs) Update(ctx context.Context, cronTab *v1alpha1.CronTab, opts v1.UpdateOptions) (result *v1alpha1.CronTab, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(crontabsResource, c.ns, cronTab), &v1alpha1.CronTab{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CronTab), err
}

// Delete takes name of the cronTab and deletes it. Returns an error if one occurs.
func (c *FakeCronTabs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(crontabsResource, c.ns, name), &v1alpha1.CronTab{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCronTabs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(crontabsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.CronTabList{})
	return err
}

// Patch applies the patch and returns the patched cronTab.
func (c *FakeCronTabs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CronTab, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(crontabsResource, c.ns, name, pt, data, subresources...), &v1alpha1.CronTab{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CronTab), err
}

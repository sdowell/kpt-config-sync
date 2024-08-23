// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1 "kpt.dev/configsync/pkg/api/configmanagement/v1"
)

// FakeNamespaceSelectors implements NamespaceSelectorInterface
type FakeNamespaceSelectors struct {
	Fake *FakeConfigmanagementV1
}

var namespaceselectorsResource = v1.SchemeGroupVersion.WithResource("namespaceselectors")

var namespaceselectorsKind = v1.SchemeGroupVersion.WithKind("NamespaceSelector")

// Get takes name of the namespaceSelector, and returns the corresponding namespaceSelector object, and an error if there is any.
func (c *FakeNamespaceSelectors) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.NamespaceSelector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(namespaceselectorsResource, name), &v1.NamespaceSelector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.NamespaceSelector), err
}

// List takes label and field selectors, and returns the list of NamespaceSelectors that match those selectors.
func (c *FakeNamespaceSelectors) List(ctx context.Context, opts metav1.ListOptions) (result *v1.NamespaceSelectorList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(namespaceselectorsResource, namespaceselectorsKind, opts), &v1.NamespaceSelectorList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.NamespaceSelectorList{ListMeta: obj.(*v1.NamespaceSelectorList).ListMeta}
	for _, item := range obj.(*v1.NamespaceSelectorList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested namespaceSelectors.
func (c *FakeNamespaceSelectors) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(namespaceselectorsResource, opts))
}

// Create takes the representation of a namespaceSelector and creates it.  Returns the server's representation of the namespaceSelector, and an error, if there is any.
func (c *FakeNamespaceSelectors) Create(ctx context.Context, namespaceSelector *v1.NamespaceSelector, opts metav1.CreateOptions) (result *v1.NamespaceSelector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(namespaceselectorsResource, namespaceSelector), &v1.NamespaceSelector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.NamespaceSelector), err
}

// Update takes the representation of a namespaceSelector and updates it. Returns the server's representation of the namespaceSelector, and an error, if there is any.
func (c *FakeNamespaceSelectors) Update(ctx context.Context, namespaceSelector *v1.NamespaceSelector, opts metav1.UpdateOptions) (result *v1.NamespaceSelector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(namespaceselectorsResource, namespaceSelector), &v1.NamespaceSelector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.NamespaceSelector), err
}

// Delete takes name of the namespaceSelector and deletes it. Returns an error if one occurs.
func (c *FakeNamespaceSelectors) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(namespaceselectorsResource, name, opts), &v1.NamespaceSelector{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeNamespaceSelectors) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(namespaceselectorsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1.NamespaceSelectorList{})
	return err
}

// Patch applies the patch and returns the patched namespaceSelector.
func (c *FakeNamespaceSelectors) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.NamespaceSelector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(namespaceselectorsResource, name, pt, data, subresources...), &v1.NamespaceSelector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.NamespaceSelector), err
}
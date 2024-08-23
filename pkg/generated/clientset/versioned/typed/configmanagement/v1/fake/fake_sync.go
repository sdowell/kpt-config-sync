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

// FakeSyncs implements SyncInterface
type FakeSyncs struct {
	Fake *FakeConfigmanagementV1
}

var syncsResource = v1.SchemeGroupVersion.WithResource("syncs")

var syncsKind = v1.SchemeGroupVersion.WithKind("Sync")

// Get takes name of the sync, and returns the corresponding sync object, and an error if there is any.
func (c *FakeSyncs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Sync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(syncsResource, name), &v1.Sync{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Sync), err
}

// List takes label and field selectors, and returns the list of Syncs that match those selectors.
func (c *FakeSyncs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.SyncList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(syncsResource, syncsKind, opts), &v1.SyncList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.SyncList{ListMeta: obj.(*v1.SyncList).ListMeta}
	for _, item := range obj.(*v1.SyncList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested syncs.
func (c *FakeSyncs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(syncsResource, opts))
}

// Create takes the representation of a sync and creates it.  Returns the server's representation of the sync, and an error, if there is any.
func (c *FakeSyncs) Create(ctx context.Context, sync *v1.Sync, opts metav1.CreateOptions) (result *v1.Sync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(syncsResource, sync), &v1.Sync{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Sync), err
}

// Update takes the representation of a sync and updates it. Returns the server's representation of the sync, and an error, if there is any.
func (c *FakeSyncs) Update(ctx context.Context, sync *v1.Sync, opts metav1.UpdateOptions) (result *v1.Sync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(syncsResource, sync), &v1.Sync{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Sync), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSyncs) UpdateStatus(ctx context.Context, sync *v1.Sync, opts metav1.UpdateOptions) (*v1.Sync, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(syncsResource, "status", sync), &v1.Sync{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Sync), err
}

// Delete takes name of the sync and deletes it. Returns an error if one occurs.
func (c *FakeSyncs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(syncsResource, name, opts), &v1.Sync{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSyncs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(syncsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1.SyncList{})
	return err
}

// Patch applies the patch and returns the patched sync.
func (c *FakeSyncs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Sync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(syncsResource, name, pt, data, subresources...), &v1.Sync{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Sync), err
}
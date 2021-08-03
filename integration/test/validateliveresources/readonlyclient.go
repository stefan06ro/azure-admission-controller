package validateliveresources

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReadOnlyCtrlClient struct {
	t      *testing.T
	client client.Client
}

func (c *ReadOnlyCtrlClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return c.client.Get(ctx, key, obj)
}

func (c *ReadOnlyCtrlClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	return c.client.List(ctx, list, opts...)
}

func (c *ReadOnlyCtrlClient) Create(context.Context, runtime.Object, ...client.CreateOption) error {
	c.fatal()
	return nil
}

func (c *ReadOnlyCtrlClient) Delete(context.Context, runtime.Object, ...client.DeleteOption) error {
	c.fatal()
	return nil
}

func (c *ReadOnlyCtrlClient) Update(context.Context, runtime.Object, ...client.UpdateOption) error {
	c.fatal()
	return nil
}

func (c *ReadOnlyCtrlClient) Patch(context.Context, runtime.Object, client.Patch, ...client.PatchOption) error {
	c.fatal()
	return nil
}

func (c *ReadOnlyCtrlClient) DeleteAllOf(context.Context, runtime.Object, ...client.DeleteAllOfOption) error {
	c.fatal()
	return nil
}

func (c *ReadOnlyCtrlClient) Status() client.StatusWriter {
	c.fatal()
	return nil
}

func (c *ReadOnlyCtrlClient) fatal() {
	c.t.Fatalf("It is forbidden to call write functions in a test that is executed on a live installation.")
}

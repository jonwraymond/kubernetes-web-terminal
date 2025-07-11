package client

import (
	"context"
	"fmt"

	terminalv1 "github.com/jraymond/kubernetes-web-terminal/pkg/apis/terminal/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// TerminalConfigClient provides a client for TerminalConfig resources
type TerminalConfigClient struct {
	dynamicClient dynamic.Interface
	namespace     string
}

// NewTerminalConfigClient creates a new TerminalConfig client
func NewTerminalConfigClient(config *rest.Config, namespace string) (*TerminalConfigClient, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return &TerminalConfigClient{
		dynamicClient: dynamicClient,
		namespace:     namespace,
	}, nil
}

// gvr returns the GroupVersionResource for TerminalConfig
func (c *TerminalConfigClient) gvr() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    terminalv1.SchemeGroupVersion.Group,
		Version:  terminalv1.SchemeGroupVersion.Version,
		Resource: "terminalconfigs",
	}
}

// Get retrieves a TerminalConfig by name
func (c *TerminalConfigClient) Get(ctx context.Context, name string) (*terminalv1.TerminalConfig, error) {
	resource := c.dynamicClient.Resource(c.gvr()).Namespace(c.namespace)
	unstructured, err := resource.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get TerminalConfig %s: %v", name, err)
	}

	var terminalConfig terminalv1.TerminalConfig
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.UnstructuredContent(), &terminalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert unstructured to TerminalConfig: %v", err)
	}

	return &terminalConfig, nil
}

// List retrieves all TerminalConfigs in the namespace
func (c *TerminalConfigClient) List(ctx context.Context) (*terminalv1.TerminalConfigList, error) {
	resource := c.dynamicClient.Resource(c.gvr()).Namespace(c.namespace)
	unstructuredList, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list TerminalConfigs: %v", err)
	}

	var terminalConfigList terminalv1.TerminalConfigList
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredList.UnstructuredContent(), &terminalConfigList)
	if err != nil {
		return nil, fmt.Errorf("failed to convert unstructured list to TerminalConfigList: %v", err)
	}

	return &terminalConfigList, nil
}

// Create creates a new TerminalConfig
func (c *TerminalConfigClient) Create(ctx context.Context, tc *terminalv1.TerminalConfig) (*terminalv1.TerminalConfig, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TerminalConfig to unstructured: %v", err)
	}

	resource := c.dynamicClient.Resource(c.gvr()).Namespace(c.namespace)
	unstructured := &unstructured.Unstructured{Object: unstructuredObj}
	created, err := resource.Create(ctx, unstructured, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create TerminalConfig: %v", err)
	}

	var result terminalv1.TerminalConfig
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(created.UnstructuredContent(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created object to TerminalConfig: %v", err)
	}

	return &result, nil
}

// Update updates an existing TerminalConfig
func (c *TerminalConfigClient) Update(ctx context.Context, tc *terminalv1.TerminalConfig) (*terminalv1.TerminalConfig, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TerminalConfig to unstructured: %v", err)
	}

	resource := c.dynamicClient.Resource(c.gvr()).Namespace(c.namespace)
	unstructured := &unstructured.Unstructured{Object: unstructuredObj}
	updated, err := resource.Update(ctx, unstructured, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update TerminalConfig: %v", err)
	}

	var result terminalv1.TerminalConfig
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(updated.UnstructuredContent(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert updated object to TerminalConfig: %v", err)
	}

	return &result, nil
}

// Delete deletes a TerminalConfig by name
func (c *TerminalConfigClient) Delete(ctx context.Context, name string) error {
	resource := c.dynamicClient.Resource(c.gvr()).Namespace(c.namespace)
	err := resource.Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete TerminalConfig %s: %v", name, err)
	}
	return nil
}
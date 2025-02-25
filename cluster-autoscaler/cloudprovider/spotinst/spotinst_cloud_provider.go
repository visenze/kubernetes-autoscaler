/*
Copyright 2016 The Kubernetes Authors.
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

package spotinst

import (
	"io"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
	"os"
)

const (
	// ProviderName is the provide name of spotinst
	ProviderName = "spotinst"

	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "k8s.amazonaws.com/accelerator"
)

// CloudProvider implements CloudProvider interface.
type CloudProvider struct {
	manager         *CloudManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
	}
)

// NewCloudProvider returns CloudProvider implementation for Spotinst.
func NewCloudProvider(manager *CloudManager, resourceLimiter *cloudprovider.ResourceLimiter) (*CloudProvider, error) {
	klog.Info("Building Spotinst cloud provider")
	cloud := &CloudProvider{
		manager:         manager,
		resourceLimiter: resourceLimiter,
	}

	return cloud, nil
}

// Name returns name of the cloud c.
func (c *CloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud c.
func (c *CloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	out := make([]cloudprovider.NodeGroup, len(c.manager.groups))
	for i, group := range c.manager.groups {
		out[i] = group
	}
	return out
}

// NodeGroupForNode returns the node group for the given node.
func (c *CloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	instanceID, err := extractInstanceId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	return c.manager.GetGroupForInstance(instanceID)
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (c *CloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (c *CloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (c *CloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
func (c *CloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (c *CloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return c.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (c *CloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (c *CloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (c *CloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(c, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (c *CloudProvider) Cleanup() error {
	return c.manager.Cleanup()
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (c *CloudProvider) Refresh() error {
	return c.manager.Refresh()
}

// GetInstanceID gets the instance ID for the specified node.
func (c *CloudProvider) GetInstanceID(node *apiv1.Node) string {
	return node.Spec.ProviderID
}

// BuildSpotinst return the spotinst provider
func BuildSpotinst(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := NewCloudManager(config, do)
	if err != nil {
		klog.Fatalf("Failed to create Spotinst manager: %v", err)
	}

	provider, err := NewCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create Spotinst cloud provider: %v", err)
	}

	return provider
}
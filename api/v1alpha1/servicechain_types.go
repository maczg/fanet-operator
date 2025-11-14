package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FunctionSpec defines a function in a service chain
type FunctionSpec struct {
	// Name is the name of the function
	Name string `json:"name"`

	// Image is the container image
	Image string `json:"image"`

	// Args are the container arguments
	Args []string `json:"args,omitempty"`

	// DroneRef references the drone to deploy on
	DroneRef *v1.ObjectReference `json:"droneRef"`

	// Resources are the resource requirements
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// IdleEnergyConsumptionWh is the idle energy consumption
	IdleEnergyConsumptionWh string `json:"idleEnergyConsumptionWh,omitempty"`

	// ActiveEnergyConsumptionWh is the active energy consumption
	ActiveEnergyConsumptionWh string `json:"activeEnergyConsumptionWh,omitempty"`

	// MaxLatencyMs is the maximum latency in milliseconds
	MaxLatencyMs int `json:"maxLatencyMs,omitempty"`
}

// ServiceChainSpec defines the desired state of ServiceChain.
type ServiceChainSpec struct {
	// Functions is the list of virtual functions in this service chain
	Functions []FunctionSpec `json:"functions,omitempty"`
}

// ServiceChainStatus defines the observed state of ServiceChain.
type ServiceChainStatus struct {
	// VFCount is the number of virtual functions in this chain
	VFCount int `json:"vfCount,omitempty"`

	// TotalVFLatency is the total latency of all VFs
	TotalVFLatency int `json:"totalVfLatency,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ServiceChain is the Schema for the servicechains API.
type ServiceChain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceChainSpec   `json:"spec,omitempty"`
	Status ServiceChainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceChainList contains a list of ServiceChain.
type ServiceChainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceChain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceChain{}, &ServiceChainList{})
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChainPerformanceMetrics defines the performance metrics for a Service Chain
type ChainPerformanceMetrics struct {
	// MaxTotalPropagationDelay represents the maximum allowed total propagation delay (milliseconds)
	// +kubebuilder:validation:Minimum=0
	MaxTotalPropagationDelay int32 `json:"maxTotalPropagationDelay"`
}

// FunctionSpec defines the specification for a single function in the service chain
type FunctionSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Image string `json:"image"`
	// +kubebuilder:validation:Required
	UavName string `json:"uavName"`
	// +kubebuilder:validation:Optional
	Args []string `json:"args,omitempty"`

	// Expected propagation delay for this function (milliseconds)
	// +kubebuilder:validation:Minimum=0
	ExpectedPropagationDelay int32 `json:"expectedPropagationDelay,omitempty"`
}

// ServiceChainSpec defines the desired state of ServiceChain.
type ServiceChainSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Functions []FunctionSpec `json:"functions,omitempty"`

	// Performance metrics for the entire chain
	PerformanceMetrics ChainPerformanceMetrics `json:"performanceMetrics"`
}

// ServiceChainStatus defines the observed state of ServiceChain
type ServiceChainStatus struct {
	Functions []FunctionStatus `json:"functions,omitempty"`

	// Total current propagation delay
	TotalPropagationDelay int32 `json:"totalPropagationDelay,omitempty"`

	// Total current energy consumption (in watts, stored as string)
	TotalEnergyConsumption string `json:"totalEnergyConsumption,omitempty"`

	// Chain state
	State string `json:"state,omitempty"`

	// Status message describing the current state
	StatusMessage string `json:"statusMessage,omitempty"`

	// Last update time
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

// FunctionStatus defines the status of a single function in the service chain
type FunctionStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Node   string `json:"node,omitempty"`

	// Current propagation delay
	CurrentPropagationDelay int32 `json:"currentPropagationDelay,omitempty"`

	// Current energy consumption (in watts, stored as string)
	CurrentEnergyConsumption string `json:"currentEnergyConsumption,omitempty"`
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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualFunctionSpec defines the desired state of VirtualFunction.
type VirtualFunctionSpec struct {
	// Image is the container image to use
	Image string `json:"image"`

	// Args are the arguments for the container
	Args []string `json:"args,omitempty"`

	// DroneRef references the drone to deploy on
	DroneRef *v1.ObjectReference `json:"droneRef"`

	// ServiceChainName is the name of the service chain this VF belongs to
	ServiceChainName string `json:"serviceChainName,omitempty"`

	// Resources are the resource requirements
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// IdleEnergyConsumptionWh is the idle energy consumption
	IdleEnergyConsumptionWh string `json:"idleEnergyConsumptionWh,omitempty"`

	// ActiveEnergyConsumptionWh is the active energy consumption
	ActiveEnergyConsumptionWh string `json:"activeEnergyConsumptionWh,omitempty"`

	// MaxLatencyMs is the maximum latency in milliseconds
	MaxLatencyMs int `json:"maxLatencyMs,omitempty"`
}

// VirtualFunctionStatus defines the observed state of VirtualFunction.
type VirtualFunctionStatus struct {
	// Phase is the current phase of the virtual function
	Phase string `json:"phase,omitempty"`

	// LastUpdated is the last update timestamp
	LastUpdated string `json:"lastUpdated,omitempty"`

	// StatusPODVF is the status of the pod
	StatusPODVF string `json:"statusPODVF,omitempty"`

	// LastDroneRef tracks the last drone this VF was deployed on
	LastDroneRef string `json:"lastDroneRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Drone",type="string",JSONPath=".spec.droneRef.name",description="Drone Reference"
// +kubebuilder:printcolumn:name="ServiceChain",type="string",JSONPath=".spec.serviceChainName",description="Service Chain Name"
// +kubebuilder:printcolumn:name="Pod Status",type="string",JSONPath=".status.statusPODVF",description="Pod Status of the Virtual Function"
// +kubebuilder:printcolumn:name="Last Updated",type="string",JSONPath=".status.lastUpdated",description="Last Updated Timestamp"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Current Phase of the Virtual Function"

// VirtualFunction is the Schema for the virtualfunctions API.
type VirtualFunction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualFunctionSpec   `json:"spec,omitempty"`
	Status VirtualFunctionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualFunctionList contains a list of VirtualFunction.
type VirtualFunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualFunction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualFunction{}, &VirtualFunctionList{})
}

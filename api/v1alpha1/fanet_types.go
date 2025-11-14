package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DroneReference contains reference to a Drone
type DroneReference struct {
	// NodeRef is the reference to the node the drone is hosting
	Name string `json:"name"`
	// DroneName is the name of the drone
	NodeRef *NodeReference `json:"nodeRef,omitempty"`
}

// FanetSpec defines the desired state of Fanet.
type FanetSpec struct {
	// Drones is the list of drones in this FANET
	Drones []DroneReference `json:"drones,omitempty"`
}

// DroneStatusInfo contains status information about a drone
type DroneStatusInfo struct {
	DroneName         string            `json:"droneName"`
	OperationalStatus OperationalStatus `json:"operationalStatus,omitempty"`
	BatteryLevel      int               `json:"batteryLevel,omitempty"`
	ComputingDrone    bool              `json:"computingDrone,omitempty"`
	VFCount           int               `json:"vfCount,omitempty"`
	NodeReady         bool              `json:"nodeReady,omitempty"`
	CPUAvailable      string            `json:"cpuAvailable,omitempty"`
	MemoryAvailable   string            `json:"memoryAvailable,omitempty"`
}

// FanetStatus defines the observed state of Fanet.
type FanetStatus struct {
	// TotalDroneCount is the total number of drones
	TotalDroneCount int `json:"totalDroneCount,omitempty"`

	// InFlightCount is the number of drones in flight
	InFlightCount int `json:"inFlightCount,omitempty"`

	// LeavingCount is the number of drones leaving
	LeavingCount int `json:"leavingCount,omitempty"`

	// NotAvailableCount is the number of drones not available
	NotAvailableCount int `json:"notAvailableCount,omitempty"`

	// DroneStatuses is the status of each drone
	DroneStatuses []DroneStatusInfo `json:"droneStatuses,omitempty"`

	// VFTotal is the total number of virtual functions
	VFTotal int `json:"vfTotal,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=fn
// +kubebuilder:printcolumn:name="TotalDrones",type="integer",JSONPath=".status.totalDroneCount",description="Total number of drones in the FANET"
// +kubebuilder:printcolumn:name="InFlight",type="integer",JSONPath=".status.inFlightCount",description="Number of drones in flight"
// +kubebuilder:printcolumn:name="Leaving",type="integer",JSONPath=".status.leavingCount",description="Number of drones leaving"
// +kubebuilder:printcolumn:name="NotAvailable",type="integer",JSONPath=".status.notAvailableCount",description="Number of drones not available"
// +kubebuilder:printcolumn:name="VFTotal",type="integer",JSONPath=".status.vfTotal",description="Total number of virtual functions in the FANET"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age of the FANET"

// Fanet is the Schema for the fanets API.
type Fanet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FanetSpec   `json:"spec,omitempty"`
	Status            FanetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FanetList contains a list of Fanet.
type FanetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Fanet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Fanet{}, &FanetList{})
}

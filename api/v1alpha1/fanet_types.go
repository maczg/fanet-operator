package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DroneReference contains reference to a Drone
type DroneReference struct {
	// NodeRef is the reference to the node the drone is hosting
	NodeRef string `json:"nodeRef"`
	// DroneName is the name of the drone
	DroneName string `json:"name"`
}

// FanetSpec defines the desired state of Fanet.
type FanetSpec struct {
	// Drones is the list of drones in this FANET
	Drones []DroneReference `json:"drones,omitempty"`
}

// DroneStatusSummary contains summary status of a drone in the FANET.
type DroneStatusSummary struct {
	// Name of the drone
	Name string `json:"name"`
	// Status is the operational status
	Status string `json:"status"`
	// BatteryLevel is the battery level
	BatteryLevel string `json:"batteryLevel"`
	// VFCount is the number of VFs
	VFCount int32 `json:"vfCount"`
	// ComputingDrone indicates if this drone is a computing drone or not
	ComputingDrone bool `json:"computingDrone"`
}

// FanetStatus defines the observed state of Fanet.
type FanetStatus struct {
	// TotalDroneCount is the total number of drones in this FANET
	TotalDroneCount int32 `json:"totalDroneCount"`
	// InFlightCount is the number of drones in flight
	InFlightCount int32 `json:"inFlightCount"`
	// LeavingCount is the number of drones leaving
	LeavingCount int32 `json:"leavingCount"`
	// NotAvailableCount is the number of drones not available
	NotAvailableCount int32 `json:"notAvailableCount"`
	// VFTotal is the total number of virtual functions across all drones
	VFTotal int32 `json:"vfTotal"`
	// DroneStatuses contains the status of each drone
	DroneStatuses []DroneStatusSummary `json:"droneStatuses"`
	// LastUpdated is the last time the status was updated
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=fn
// +kubebuilder:printcolumn:name="TotalDrones",type="integer",JSONPath=".status.totalDroneCount",description="Total number of drones in the FANET"
// +kubebuilder:printcolumn:name="InFlight",type="integer",JSONPath=".status.inFlightCount",description="Number of drones in flight"
// +kubebuilder:printcolumn:name="Leaving",type="integer",JSONPath=".status.leavingCount",description="Number of drones leaving"
// +kubebuilder:printcolumn:name="NotAvailable",type="integer",JSONPath=".status.notAvailableCount",description="Number of drones not available"
// +kubebuilder:printcolumn:name="VFTotal",type="integer",JSONPath=".status.vfTotal",description="Total number of virtual functions across all drones"
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

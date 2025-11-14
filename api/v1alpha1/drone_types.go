package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FanetReference references a FANET
type FanetReference struct {
	Name string `json:"name"`
}

// NodeReference references a Kubernetes node
type NodeReference struct {
	Name string `json:"name"`
}

// ResourceCapacity represents resource capacity
type ResourceCapacity struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// OperationalStatus describes the operational status of the drone.
// Only one of the following statuses can be set at a time.
// If none of the following statuses is set, the status is considered unknown.
// +kubebuilder:validation:Enum=Flying;Leaving;NotAvailable
type OperationalStatus string

const (
	// OperationalStatusFlying indicates the drone is in flight
	OperationalStatusFlying OperationalStatus = "Flying"
	// OperationalStatusLeaving indicates the drone is leaving
	OperationalStatusLeaving OperationalStatus = "Leaving"
	// OperationalStatusNotAvailable indicates the drone is not available or its status is unknown
	OperationalStatusNotAvailable OperationalStatus = "NotAvailable"
)

// DroneSpec defines the desired state of Drone.
type DroneSpec struct {
	// BatteryCapacity is the battery capacity in Wh
	BatteryCapacity int `json:"batteryCapacity,omitempty"`

	// MotorIdleConsumptionWh is the motor idle consumption in Wh
	MotorIdleConsumptionWh string `json:"motorIdleConsumptionWh,omitempty"`

	// CEIdleConsumptionWh is the CE idle consumption in Wh
	CEIdleConsumptionWh string `json:"ceIdleConsumptionWh,omitempty"`

	// NodeRef references the Kubernetes node
	NodeRef *NodeReference `json:"nodeRef,omitempty"`

	// FanetRef references the FANET this drone belongs to
	FanetRef *FanetReference `json:"fanetRef,omitempty"`

	// ComputingDrone indicates if this is a computing drone
	ComputingDrone bool `json:"computingDrone,omitempty"`
}

// DroneStatus defines the observed state of Drone.
type DroneStatus struct {
	// OperationalStatus is the current operational status of the drone
	// Valid values are:
	// - Flying: The drone is currently in flight.
	// - Leaving: The drone is in the process of leaving.
	// - NotAvailable (default): The drone is not available or its status is unknown.
	// +optional
	// +kubebuilder:default:=NotAvailable
	OperationalStatus *OperationalStatus `json:"operationalStatus,omitempty"`

	// BatteryLevel is the current battery level percentage
	BatteryLevel int `json:"batteryLevel,omitempty"`

	// VFCount is the number of virtual functions running on this drone
	VFCount int `json:"vfCount,omitempty"`

	// LastUpdated is the last update timestamp
	LastUpdated string `json:"lastUpdated,omitempty"`

	// ComputingDrone indicates if this is a computing drone
	ComputingDrone bool `json:"computingDrone,omitempty"`

	// NodeReady indicates if the node is ready
	NodeReady bool `json:"nodeReady,omitempty"`

	// CPUAvailable is the available CPU
	CPUAvailable string `json:"cpuAvailable,omitempty"`

	// MemoryAvailable is the available memory
	MemoryAvailable string `json:"memoryAvailable,omitempty"`

	// CEMaxCapacity is the maximum capacity of the CE
	CEMaxCapacity *ResourceCapacity `json:"ceMaxCapacity,omitempty"`

	// FanetRef references the FANET this drone is assigned to
	FanetRef *FanetReference `json:"fanetRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=dr
// +kubebuilder:printcolumn:name="OperationalStatus",type="string",JSONPath=".status.operationalStatus",description="Current operational status of the drone"
// +kubebuilder:printcolumn:name="FANET",type="string",JSONPath=".status.fanetRef.name",description="FANET this drone is assigned to"
// +kubebuilder:printcolumn:name="BatteryLevel",type="integer",JSONPath=".status.batteryLevel",description="Current battery level percentage"
// +kubebuilder:printcolumn:name="ComputingDrone",type="boolean",JSONPath=".status.computingDrone",description="Indicates if this is a computing drone"
// +kubebuilder:printcolumn:name="NodeReady",type="boolean",JSONPath=".status.nodeReady",description="Indicates if the node is ready"
// +kubebuilder:printcolumn:name="CPUAvailable",type="string",JSONPath=".status.cpuAvailable",description="Available CPU"
// +kubebuilder:printcolumn:name="MemoryAvailable",type="string",JSONPath=".status.memoryAvailable",description="Available memory"
// +kubebuilder:printcolumn:name="VFCount",type="integer",JSONPath=".status.vfCount",description="Number of virtual functions running on this drone"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age of the drone"
// +kubebuilder:printcolumn:name="LastUpdated",type="string",JSONPath=".status.lastUpdated",description="Last update timestamp"

// Drone is the Schema for the drones API.
type Drone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DroneSpec   `json:"spec,omitempty"`
	Status DroneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DroneList contains a list of Drone.
type DroneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Drone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Drone{}, &DroneList{})
}

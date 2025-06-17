/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UAVFlightStatus defines the observed state of UAV.
// +kubebuilder:validation:Enum=In-flight;Charging;Joining;Leaving;Unavailable
type UAVFlightStatus string

const (
	// InUAVInFlightFlight indicates that the UAV is currently flying.
	UAVInFlight UAVFlightStatus = "In-flight"
	// UAVCharging indicates that the UAV is currently charging.
	UAVCharging UAVFlightStatus = "Charging"
	// UAVJoining indicates that the UAV is in the process of joining a network.
	UAVJoining UAVFlightStatus = "Joining"
	// UAVLeaving indicates that the UAV is in the process of leaving a network.
	UAVLeaving UAVFlightStatus = "Leaving"
	// UAVUnavailable indicates that the UAV is not available for flight.
	UAVUnavailable UAVFlightStatus = "Unavailable"
)

// BatteryLevel defines the battery level of the UAV.
// +kubebuilder:validation:Minimum=0
// +kubebuilder:validation:Maximum=100
// +kubebuilder:validation:ExclusiveMaximum=false
type BatteryLevel int32

// UAVEnergyConsumption defines the energy consumption metrics for a UAV
type UAVEnergyConsumption struct {
	// MotorIdleConsumption represents the energy consumption of the UAV motor in idle state (watts, stored as string)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	MotorIdleConsumption string `json:"motorIdleConsumption"`

	// CEIdleConsumption represents the energy consumption of the Computing Element in idle state (watts, stored as string)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	CEIdleConsumption string `json:"ceIdleConsumption"`
}

// CECapacity defines the computing capabilities of the UAV
type CECapacity struct {
	// MaxCapacity represents the maximum capacity of the Computing Element
	// +kubebuilder:validation:Minimum=0
	MaxCapacity int32 `json:"maxCapacity"`

	// DeployedVFs represents the number of Virtual Functions currently deployed
	// +kubebuilder:validation:Minimum=0
	DeployedVFs int32 `json:"deployedVFs"`
}

// NetworkMetrics defines the network-related metrics for the UAV
type NetworkMetrics struct {
	// InterUAVDelay represents the maximum delay for communication with other UAVs (milliseconds)
	// +kubebuilder:validation:Minimum=0
	InterUAVDelay int32 `json:"interUAVDelay"`
}

// UAVSpec defines the desired state of UAV
type UAVSpec struct {
	// NodeName is the name of the node where the UAV is deployed
	NodeName string `json:"nodeName"`

	// FlightStatus represents the current flight status of the UAV
	FlightStatus UAVFlightStatus `json:"flightStatus"`

	// BatteryConfig defines the battery configuration parameters
	BatteryConfig BatteryConfig `json:"batteryConfig"`

	// EnergyConfig defines the energy consumption configuration
	EnergyConfig EnergyConfig `json:"energyConfig"`

	// CECapacity defines the computing capacity of the UAV
	CECapacity CECapacity `json:"ceCapacity"`

	// NetworkConfig defines the network configuration parameters
	NetworkConfig NetworkConfig `json:"networkConfig"`
}

// BatteryConfig defines the battery configuration parameters
type BatteryConfig struct {
	// MaxCapacity is the maximum battery capacity in percentage
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MaxCapacity int `json:"maxCapacity"`

	// CriticalLevel is the battery level that triggers critical status
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	CriticalLevel int `json:"criticalLevel"`

	// LowLevel is the battery level that triggers low battery warning
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	LowLevel int `json:"lowLevel"`
}

// EnergyConfig defines the energy consumption configuration
type EnergyConfig struct {
	// MotorIdleConsumption is the energy consumption of motors in idle state (in watts)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	MotorIdleConsumption string `json:"motorIdleConsumption"`

	// CEIdleConsumption is the energy consumption of computing elements in idle state (in watts)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	CEIdleConsumption string `json:"ceIdleConsumption"`

	// MaxTotalConsumption is the maximum allowed total energy consumption (in watts)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	MaxTotalConsumption string `json:"maxTotalConsumption"`
}

// NetworkConfig defines the network configuration parameters
type NetworkConfig struct {
	// MaxInterUAVDelay is the maximum allowed delay between UAVs (in milliseconds)
	// +kubebuilder:validation:Minimum=0
	MaxInterUAVDelay int `json:"maxInterUAVDelay"`
}

// UAVStatus defines the observed state of UAV
type UAVStatus struct {
	// NodeReady indicates whether the UAV's node is ready to host virtual functions
	NodeReady bool `json:"nodeReady,omitempty"`

	// FlightStatus represents the current flight status of the UAV
	FlightStatus UAVFlightStatus `json:"flightStatus,omitempty"`

	// CurrentBatteryLevel is the current battery level in percentage
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	CurrentBatteryLevel int `json:"currentBatteryLevel,omitempty"`

	// CurrentEnergyConsumption is the current total energy consumption (in watts)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	CurrentEnergyConsumption string `json:"currentEnergyConsumption,omitempty"`

	// BatteryHealth represents the health status of the battery
	// +kubebuilder:validation:Enum=Good;Degraded;Critical
	BatteryHealth string `json:"batteryHealth,omitempty"`

	// EnergyEfficiency represents the current energy efficiency status
	// +kubebuilder:validation:Enum=Optimal;Suboptimal;Critical
	EnergyEfficiency string `json:"energyEfficiency,omitempty"`

	// LastUpdateTime is the timestamp of the last status update
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Conditions represent the latest available observations of the UAV's current state
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// UAV is the Schema for the uavs API.
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.nodeName`,description="the name of the node where the UAV is running"
// +kubebuilder:printcolumn:name="Battery",type=integer,JSONPath=`.spec.batteryLevel`,description="the battery level of the UAV, in percent"
// +kubebuilder:printcolumn:name="Flight Status",type=string,JSONPath=`.status.flightStatus`,description="the current flight status of the UAV"
type UAV struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UAVSpec   `json:"spec,omitempty"`
	Status UAVStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UAVList contains a list of UAV.
type UAVList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UAV `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UAV{}, &UAVList{})
}

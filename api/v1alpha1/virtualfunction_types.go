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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnergyConsumption defines the energy consumption metrics for a Virtual Function
type EnergyConsumption struct {
	// IdleConsumption represents the minimum energy consumption in idle state (watts, stored as string)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	IdleConsumption string `json:"idleConsumption"`

	// ActiveConsumption represents the maximum energy consumption in active state (watts, stored as string)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?$`
	ActiveConsumption string `json:"activeConsumption"`
}

// PerformanceMetrics defines the performance metrics for a Virtual Function
type PerformanceMetrics struct {
	// MaxPropagationDelay represents the maximum allowed propagation delay (milliseconds)
	// +kubebuilder:validation:Minimum=0
	MaxPropagationDelay int32 `json:"maxPropagationDelay"`
}

// VirtualFunctionSpec defines the desired state of VirtualFunction.
type VirtualFunctionSpec struct {
	Image     string                      `json:"image"`
	UAVName   string                      `json:"uavName"`
	Args      []string                    `json:"args,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Energy consumption metrics
	EnergyConsumption EnergyConsumption `json:"energyConsumption"`

	// Performance metrics
	PerformanceMetrics PerformanceMetrics `json:"performanceMetrics"`
}

// VirtualFunctionStatus defines the observed state of VirtualFunction.
type VirtualFunctionStatus struct {
	// Current energy consumption (in watts, stored as string)
	CurrentEnergyConsumption string `json:"currentEnergyConsumption,omitempty"`

	// Current propagation delay
	CurrentPropagationDelay int32 `json:"currentPropagationDelay,omitempty"`

	// Function state
	State string `json:"state,omitempty"`

	// Status message providing additional information about the current state
	StatusMessage string `json:"statusMessage,omitempty"`

	// Last update time
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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

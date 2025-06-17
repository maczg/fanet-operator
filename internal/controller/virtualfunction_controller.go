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

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fanetv1alpha1 "github.com/maczg/fanet-operator/api/v1alpha1"
)

// VirtualFunctionReconciler reconciles a VirtualFunction object
type VirtualFunctionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fanet.restart.eu,resources=virtualfunctions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fanet.restart.eu,resources=virtualfunctions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fanet.restart.eu,resources=virtualfunctions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *VirtualFunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the VirtualFunction instance
	var vf fanetv1alpha1.VirtualFunction
	if err := r.Get(ctx, req.NamespacedName, &vf); err != nil {
		if errors.IsNotFound(err) {
			log.Info("VirtualFunction resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get VirtualFunction")
		return ctrl.Result{}, err
	}

	// Check if the assigned UAV exists and is ready
	var uav fanetv1alpha1.UAV
	if err := r.Get(ctx, types.NamespacedName{Name: vf.Spec.UAVName}, &uav); err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "Assigned UAV not found")
			vf.Status.State = "error"
			vf.Status.StatusMessage = fmt.Sprintf("Assigned UAV %s not found", vf.Spec.UAVName)
			if err := r.Status().Update(ctx, &vf); err != nil {
				log.Error(err, "Failed to update VirtualFunction status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get UAV")
		return ctrl.Result{}, err
	}

	// Validate UAV state
	if !uav.Status.NodeReady {
		log.Info("UAV node is not ready")
		vf.Status.State = "pending"
		vf.Status.StatusMessage = "UAV node is not ready"
		if err := r.Status().Update(ctx, &vf); err != nil {
			log.Error(err, "Failed to update VirtualFunction status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	if uav.Status.FlightStatus != fanetv1alpha1.UAVInFlight {
		log.Info("UAV is not in flight")
		vf.Status.State = "pending"
		vf.Status.StatusMessage = fmt.Sprintf("UAV is not in flight (current status: %s)", uav.Status.FlightStatus)
		if err := r.Status().Update(ctx, &vf); err != nil {
			log.Error(err, "Failed to update VirtualFunction status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	// Check if UAV has sufficient resources
	if !r.hasSufficientResources(&uav, &vf) {
		log.Info("UAV does not have sufficient resources")
		vf.Status.State = "error"
		vf.Status.StatusMessage = "UAV does not have sufficient resources"
		if err := r.Status().Update(ctx, &vf); err != nil {
			log.Error(err, "Failed to update VirtualFunction status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Calculate current metrics
	currentEnergyConsumption := r.calculateEnergyConsumption(&vf)
	currentPropagationDelay := r.calculatePropagationDelay(&vf)

	// Update VF status
	vf.Status.CurrentEnergyConsumption = currentEnergyConsumption
	vf.Status.CurrentPropagationDelay = currentPropagationDelay
	vf.Status.LastUpdate = metav1.Now()

	// Check if the VF is within its performance constraints
	if currentPropagationDelay > vf.Spec.PerformanceMetrics.MaxPropagationDelay {
		log.Info("VirtualFunction exceeds maximum propagation delay",
			"current", currentPropagationDelay,
			"max", vf.Spec.PerformanceMetrics.MaxPropagationDelay)
		vf.Status.State = "warning"
		vf.Status.StatusMessage = fmt.Sprintf("Propagation delay (%dms) exceeds maximum allowed (%dms)",
			currentPropagationDelay, vf.Spec.PerformanceMetrics.MaxPropagationDelay)
	} else {
		vf.Status.State = "active"
		vf.Status.StatusMessage = "Virtual function is running normally"
	}

	if err := r.Status().Update(ctx, &vf); err != nil {
		log.Error(err, "Failed to update VirtualFunction status")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("[%s] Energy=%sW | Delay=%dms | State=%s",
		vf.Name, currentEnergyConsumption, currentPropagationDelay, vf.Status.State))

	return ctrl.Result{}, nil
}

// calculateEnergyConsumption calculates the current energy consumption of the VF
func (r *VirtualFunctionReconciler) calculateEnergyConsumption(vf *fanetv1alpha1.VirtualFunction) string {
	// In a real implementation, this would measure actual energy consumption
	// For now, we'll use a simple model based on the VF's state
	if vf.Status.State == "active" {
		return vf.Spec.EnergyConsumption.ActiveConsumption
	}
	return vf.Spec.EnergyConsumption.IdleConsumption
}

// calculatePropagationDelay calculates the current propagation delay of the VF
func (r *VirtualFunctionReconciler) calculatePropagationDelay(vf *fanetv1alpha1.VirtualFunction) int32 {
	// In a real implementation, this would measure actual network delay
	// For now, we'll use a simple model based on the VF's state and resources
	baseDelay := int32(50) // Base delay in milliseconds

	// Add delay based on resource usage
	if vf.Spec.Resources.Requests != nil {
		if cpu, ok := vf.Spec.Resources.Requests[corev1.ResourceCPU]; ok {
			// More CPU usage = more processing delay
			baseDelay += int32(cpu.MilliValue() / 10)
		}
	}

	return baseDelay
}

// hasSufficientResources checks if the UAV has sufficient resources for the VF
func (r *VirtualFunctionReconciler) hasSufficientResources(uav *fanetv1alpha1.UAV, vf *fanetv1alpha1.VirtualFunction) bool {
	// Get current resource usage
	var vfs fanetv1alpha1.VirtualFunctionList
	if err := r.List(context.Background(), &vfs, client.MatchingFields{"spec.uavName": uav.Name}); err != nil {
		return false
	}

	// Calculate total resource requirements
	var totalCPU int64
	for _, existingVF := range vfs.Items {
		if existingVF.Name != vf.Name { // Exclude current VF
			if cpu, ok := existingVF.Spec.Resources.Requests[corev1.ResourceCPU]; ok {
				totalCPU += cpu.MilliValue()
			}
		}
	}

	// Add current VF requirements
	if cpu, ok := vf.Spec.Resources.Requests[corev1.ResourceCPU]; ok {
		totalCPU += cpu.MilliValue()
	}

	// Check against UAV's CE capacity
	// Convert totalCPU to int32 for comparison with MaxCapacity
	if int32(totalCPU) > uav.Spec.CECapacity.MaxCapacity {
		return false
	}

	// Check if we're not exceeding the maximum number of VFs
	if uav.Spec.CECapacity.DeployedVFs >= uav.Spec.CECapacity.MaxCapacity {
		return false
	}

	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualFunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fanetv1alpha1.VirtualFunction{}).
		Named("virtualfunction").
		Complete(r)
}

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
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fanetv1alpha1 "github.com/maczg/fanet-operator/api/v1alpha1"
)

// UAVReconciler reconciles a UAV object
type UAVReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fanet.restart.eu,resources=uavs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fanet.restart.eu,resources=uavs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fanet.restart.eu,resources=uavs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles the reconciliation of UAV resources
func (r *UAVReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the UAV instance
	uav := &fanetv1alpha1.UAV{}
	if err := r.Get(ctx, req.NamespacedName, uav); err != nil {
		if errors.IsNotFound(err) {
			log.Info("UAV resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get UAV")
		return ctrl.Result{}, err
	}

	// Update UAV status
	if err := r.updateUAVStatus(ctx, uav); err != nil {
		log.Error(err, "Failed to update UAV status")
		return ctrl.Result{}, err
	}

	// Log the current state
	log.Info("UAV status updated",
		"name", uav.Name,
		"batteryLevel", uav.Status.CurrentBatteryLevel,
		"energyConsumption", uav.Status.CurrentEnergyConsumption,
		"batteryHealth", uav.Status.BatteryHealth,
		"energyEfficiency", uav.Status.EnergyEfficiency)

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

// calculateTotalEnergyConsumption calculates the total energy consumption based on the UAV's configuration
func (r *UAVReconciler) calculateTotalEnergyConsumption(uav *fanetv1alpha1.UAV) string {
	motorIdle, _ := strconv.ParseFloat(uav.Spec.EnergyConfig.MotorIdleConsumption, 64)
	ceIdle, _ := strconv.ParseFloat(uav.Spec.EnergyConfig.CEIdleConsumption, 64)
	total := motorIdle + ceIdle
	return fmt.Sprintf("%.2f", total)
}

// calculateBatteryHealth determines the battery health based on current level and thresholds
func calculateBatteryHealth(uav *fanetv1alpha1.UAV, currentLevel int) string {
	if currentLevel <= uav.Spec.BatteryConfig.CriticalLevel {
		return "Critical"
	} else if currentLevel <= uav.Spec.BatteryConfig.LowLevel {
		return "Degraded"
	}
	return "Good"
}

// calculateEnergyEfficiency determines the energy efficiency based on current consumption and max allowed
func calculateEnergyEfficiency(uav *fanetv1alpha1.UAV, currentConsumption string) string {
	current, _ := strconv.ParseFloat(currentConsumption, 64)
	maxAllowed, _ := strconv.ParseFloat(uav.Spec.EnergyConfig.MaxTotalConsumption, 64)

	if current >= maxAllowed {
		return "Critical"
	} else if current >= (maxAllowed * 0.8) {
		return "Suboptimal"
	}
	return "Optimal"
}

// updateUAVStatus updates the UAV's status with current metrics
func (r *UAVReconciler) updateUAVStatus(ctx context.Context, uav *fanetv1alpha1.UAV) error {
	// Get current status or initialize if nil
	status := uav.Status
	if status.CurrentBatteryLevel == 0 {
		status.CurrentBatteryLevel = uav.Spec.BatteryConfig.MaxCapacity
	}

	// Calculate current energy consumption
	currentConsumption := r.calculateTotalEnergyConsumption(uav)

	// Update status fields
	status.CurrentEnergyConsumption = currentConsumption
	status.BatteryHealth = calculateBatteryHealth(uav, status.CurrentBatteryLevel)
	status.EnergyEfficiency = calculateEnergyEfficiency(uav, currentConsumption)
	status.LastUpdateTime = metav1.Now()

	// Update conditions
	conditions := []metav1.Condition{}

	// Battery conditions
	if status.CurrentBatteryLevel <= 0 {
		conditions = append(conditions, metav1.Condition{
			Type:    "BatteryCritical",
			Status:  metav1.ConditionTrue,
			Reason:  "BatteryDepleted",
			Message: "UAV battery is depleted",
		})
	} else if status.CurrentBatteryLevel <= uav.Spec.BatteryConfig.CriticalLevel {
		conditions = append(conditions, metav1.Condition{
			Type:    "BatteryLow",
			Status:  metav1.ConditionTrue,
			Reason:  "BatteryCritical",
			Message: "UAV battery is critically low",
		})
	}

	// Energy efficiency conditions
	if status.EnergyEfficiency == "Critical" {
		conditions = append(conditions, metav1.Condition{
			Type:    "EnergyEfficiencyCritical",
			Status:  metav1.ConditionTrue,
			Reason:  "HighConsumption",
			Message: "Energy consumption exceeds maximum allowed",
		})
	}

	// Flight status conditions
	if uav.Spec.FlightStatus == "Emergency" {
		conditions = append(conditions, metav1.Condition{
			Type:    "EmergencyMode",
			Status:  metav1.ConditionTrue,
			Reason:  "EmergencyStatus",
			Message: "UAV is in emergency mode",
		})
	}

	status.Conditions = conditions

	// Update the UAV status
	if err := r.Status().Update(ctx, uav); err != nil {
		return fmt.Errorf("failed to update UAV status: %w", err)
	}

	return nil
}

// checkCECapacity checks if the UAV's CE has sufficient capacity for its VFs
func (r *UAVReconciler) checkCECapacity(uav *fanetv1alpha1.UAV) (bool, int32) {
	var vfs fanetv1alpha1.VirtualFunctionList
	if err := r.List(context.Background(), &vfs, client.MatchingFields{"spec.uavName": uav.Name}); err != nil {
		return false, 0
	}

	currentUtilization := int32(len(vfs.Items))
	return currentUtilization <= uav.Spec.CECapacity.MaxCapacity, currentUtilization
}

// checkNetworkDelay checks if the UAV's network delay is within acceptable range
func (r *UAVReconciler) checkNetworkDelay(uav *fanetv1alpha1.UAV) bool {
	// Get all UAVs in the network
	var uavs fanetv1alpha1.UAVList
	if err := r.List(context.Background(), &uavs); err != nil {
		return false
	}

	// Check delay to all other UAVs
	for _, otherUAV := range uavs.Items {
		if otherUAV.Name != uav.Name {
			// Here you would implement actual network delay measurement
			// For now, we'll use the configured maximum delay
			if uav.Spec.NetworkConfig.MaxInterUAVDelay > 1000 { // Example threshold: 1 second
				return false
			}
		}
	}

	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *UAVReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fanetv1alpha1.UAV{}).
		Complete(r)
}

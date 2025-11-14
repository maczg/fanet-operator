package controller

import (
	"context"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fanetv1alpha1 "github.com/maczg/fanet-operator/api/v1alpha1"
)

const (
	FanetNotAssigned = "not-assigned"
)

// DroneReconciler reconciles a Drone object
type DroneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// convertCPU converts CPU values to millicores format
func convertCPU(cpuQuantity resource.Quantity) int {
	milliValue := cpuQuantity.MilliValue()
	return int(milliValue)
}

// convertMemory converts memory values to megabytes
func convertMemory(memQuantity resource.Quantity) int {
	bytes := memQuantity.Value()
	megabytes := bytes / (1024 * 1024)
	return int(megabytes)
}

// +kubebuilder:rbac:groups=fanet.restart.io,resources=drones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fanet.restart.io,resources=drones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fanet.restart.io,resources=drones/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

func (r *DroneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("reconciling drone", "drone", req.NamespacedName)

	var drone fanetv1alpha1.Drone
	if err := r.Get(ctx, req.NamespacedName, &drone); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("drone resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get drone")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !drone.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("drone is being deleted", "name", drone.Name)
		return ctrl.Result{}, nil
	}

	// Update status based on the spec
	if err := r.updateDroneStatus(ctx, drone); err != nil {
		logger.Error(err, "Failed to update Drone status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *DroneReconciler) updateDroneStatus(ctx context.Context, drone fanetv1alpha1.Drone) error {
	logger := logf.FromContext(ctx)

	// Retry logic for status updates
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		// Re-fetch the drone to get the latest version
		latestDrone := &fanetv1alpha1.Drone{}
		if err := r.Get(ctx, types.NamespacedName{Name: drone.Name, Namespace: drone.Namespace}, latestDrone); err != nil {
			return err
		}

		// Initialize status if needed
		if latestDrone.Status.BatteryLevel == 0 {
			latestDrone.Status.BatteryLevel = 100
		}

		// Update FANET reference only if not already set by FANET controller
		if latestDrone.Status.FanetRef == nil {
			latestDrone.Status.FanetRef = &fanetv1alpha1.FanetReference{
				Name: "not-assigned",
			}
		} else if latestDrone.Status.FanetRef.Name != FanetNotAssigned {
			// Verify FANET still exists
			fanet := &fanetv1alpha1.Fanet{}
			if err := r.Get(ctx, types.NamespacedName{
				Name:      latestDrone.Status.FanetRef.Name,
				Namespace: latestDrone.Namespace,
			}, fanet); err != nil {
				if apierrors.IsNotFound(err) {
					logger.Info("FANET not found", "name", latestDrone.Status.FanetRef.Name)
					latestDrone.Status.FanetRef = &fanetv1alpha1.FanetReference{
						Name: FanetNotAssigned,
					}
				}
			}
		}

		// Check node status if NodeRef is specified
		if drone.Spec.NodeRef != nil {
			node := &corev1.Node{}
			if err := r.Get(ctx, types.NamespacedName{
				Name: drone.Spec.NodeRef.Name,
			}, node); err != nil {
				if apierrors.IsNotFound(err) {
					logger.Info("node not found", "name", drone.Spec.NodeRef.Name)
					latestDrone.Status.ComputingDrone = false
					latestDrone.Status.OperationalStatus = ptr.To(fanetv1alpha1.OperationalStatusNotAvailable)
					latestDrone.Status.NodeReady = false
				} else {
					return err
				}
			} else {
				// Node exists
				latestDrone.Status.ComputingDrone = true
				latestDrone.Status.NodeReady = true

				// get node capacity
				if node.Status.Allocatable != nil {
					cpuQuantity := convertCPU(*node.Status.Allocatable.Cpu())
					memQuantity := convertMemory(*node.Status.Allocatable.Memory())
					latestDrone.Status.CEMaxCapacity = &fanetv1alpha1.ResourceCapacity{
						CPU:    strconv.Itoa(cpuQuantity),
						Memory: strconv.Itoa(memQuantity),
					}
					latestDrone.Status.CPUAvailable = strconv.Itoa(cpuQuantity)
					latestDrone.Status.MemoryAvailable = strconv.Itoa(memQuantity)
				}
			}
		} else {
			latestDrone.Status.ComputingDrone = false
			latestDrone.Status.OperationalStatus = ptr.To(fanetv1alpha1.OperationalStatusNotAvailable)
			latestDrone.Status.NodeReady = false
		}
		// Update operational status based on battery level
		if latestDrone.Status.BatteryLevel > 20 {
			if latestDrone.Status.NodeReady {
				latestDrone.Status.OperationalStatus = ptr.To(fanetv1alpha1.OperationalStatusFlying)
			}
		} else if latestDrone.Status.BatteryLevel >= 15 && latestDrone.Status.BatteryLevel <= 20 {
			latestDrone.Status.OperationalStatus = ptr.To(fanetv1alpha1.OperationalStatusLeaving)
		} else {
			latestDrone.Status.OperationalStatus = ptr.To(fanetv1alpha1.OperationalStatusNotAvailable)
		}

		// Update last updated timestamp
		latestDrone.Status.LastUpdated = time.Now().Format(time.RFC3339)

		// Update the status subresource
		if err := r.Status().Update(ctx, latestDrone); err != nil {
			if apierrors.IsConflict(err) {
				logger.Info("Conflict updating drone status, retrying", "drone", drone.Name, "attempt", i+1)
				if i < maxRetries-1 {
					time.Sleep(time.Millisecond * 100 * time.Duration(i+1)) // Exponential backoff
					continue                                                // Retry
				}
			}
			logger.Error(err, "Failed to update drone status", "drone", drone.Name)
			return err
		}
		logger.Info("Updated drone status", "name", drone.Name)
		return nil // Success
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *DroneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fanetv1alpha1.Drone{}).
		Named("drone").
		Complete(r)
}

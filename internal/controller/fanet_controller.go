package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fanetv1alpha1 "github.com/maczg/fanet-operator/api/v1alpha1"
)

// FanetReconciler reconciles a Fanet object
type FanetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fanet.restart.io,resources=fanets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fanet.restart.io,resources=fanets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fanet.restart.io,resources=fanets/finalizers,verbs=update
// +kubebuilder:rbac:groups=fanet.restart.io,resources=drones,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=fanet.restart.io,resources=drones/status,verbs=get;update;patch

func (r *FanetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("reconciling fanet", "fanet", req.NamespacedName)

	// Fetch the Fanet instance
	var fanet fanetv1alpha1.Fanet
	if err := r.Get(ctx, req.NamespacedName, &fanet); err != nil {
		logger.Error(err, "unable to fetch Fanet object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// handle deletion
	if !fanet.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Fanet is being deleted", "name", fanet.Name)
		return ctrl.Result{}, nil
	}

	// Update drone associations
	if err := r.updateDroneAssociations(ctx, &fanet); err != nil {
		logger.Error(err, "failed to update drone associations")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateFanetStatus(ctx, &fanet); err != nil {
		logger.Error(err, "Failed to update Fanet status")
		return ctrl.Result{}, err
	}

	// Requeue after 30 seconds for periodic updates
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *FanetReconciler) updateDroneAssociations(ctx context.Context, fanet *fanetv1alpha1.Fanet) error {
	logger := logf.FromContext(ctx)
	for _, droneSpec := range fanet.Spec.Drones {
		// Retry logic for updating drone status
		const maxRetries = 3
		for i := 0; i < maxRetries; i++ {
			drone := &fanetv1alpha1.Drone{}
			err := r.Get(ctx, types.NamespacedName{Name: droneSpec.Name, Namespace: fanet.Namespace}, drone)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Drone not found, skipping", "drone", droneSpec.Name)
					break // Move to next drone
				}
				return err
			}
			// Update drone's status with FANET reference
			drone.Status.FanetRef = &fanetv1alpha1.FanetReference{Name: fanet.Name}
			if err := r.Status().Update(ctx, drone); err != nil {
				if errors.IsConflict(err) {
					logger.Info("Conflict updating drone status, retrying", "drone", drone.Name, "attempt", i+1)
					if i < maxRetries-1 {
						time.Sleep(time.Millisecond * 100 * time.Duration(i+1)) // Exponential backoff
						continue                                                // Retry
					}
				}
				logger.Error(err, "failed to update status", "drone", drone.Name)
				return err
			}
			logger.Info("Drone Assigned to FANET", "drone", drone.Name, "fanet", fanet.Name)
			break // Success, move to next drone
		}
	}
	return nil
}

func (r *FanetReconciler) updateFanetStatus(ctx context.Context, fanet *fanetv1alpha1.Fanet) error {
	logger := logf.FromContext(ctx)

	var inFlightCount, leavingCount, notAvailableCount, vfTotal int
	var droneStatuses []fanetv1alpha1.DroneStatusInfo

	for _, droneSpec := range fanet.Spec.Drones {
		drone := &fanetv1alpha1.Drone{}
		err := r.Get(ctx, types.NamespacedName{Name: droneSpec.Name, Namespace: fanet.Namespace}, drone)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Drone not found for status update", "drone", droneSpec.Name)
				continue
			}
			return err
		}
		switch *drone.Status.OperationalStatus {
		case fanetv1alpha1.OperationalStatusFlying:
			inFlightCount++
		case fanetv1alpha1.OperationalStatusLeaving:
			leavingCount++
		case fanetv1alpha1.OperationalStatusNotAvailable:
			notAvailableCount++
		default:
			notAvailableCount++
		}

		vfTotal += drone.Status.VFCount
		logger.Info("Drone VFCount", "drone", drone.Name, "vfCount", drone.Status.VFCount, "runningTotal", vfTotal)

		droneStatuses = append(droneStatuses, fanetv1alpha1.DroneStatusInfo{
			DroneName:         drone.Name,
			OperationalStatus: *drone.Status.OperationalStatus,
			BatteryLevel:      drone.Status.BatteryLevel,
			ComputingDrone:    drone.Status.ComputingDrone,
			VFCount:           drone.Status.VFCount,
			NodeReady:         drone.Status.NodeReady,
			CPUAvailable:      drone.Status.CPUAvailable,
			MemoryAvailable:   drone.Status.MemoryAvailable,
		})
	}

	// Retry logic for updating FANET status
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		// Re-fetch the FANET object to get the latest version
		latestFanet := &fanetv1alpha1.Fanet{}
		if err := r.Get(ctx, types.NamespacedName{Name: fanet.Name, Namespace: fanet.Namespace}, latestFanet); err != nil {
			return err
		}

		// Update FANET status fields
		latestFanet.Status.TotalDroneCount = len(droneStatuses)
		latestFanet.Status.InFlightCount = inFlightCount
		latestFanet.Status.LeavingCount = leavingCount
		latestFanet.Status.NotAvailableCount = notAvailableCount
		latestFanet.Status.VFTotal = vfTotal
		latestFanet.Status.DroneStatuses = droneStatuses
		logger.Info("Updating FANET VFTotal", "fanet", fanet.Name, "vfTotal", vfTotal)

		if err := r.Status().Update(ctx, latestFanet); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Conflict updating FANET status, retrying", "fanet", fanet.Name, "attempt", i+1)
				if i < maxRetries-1 {
					time.Sleep(time.Millisecond * 100 * time.Duration(i+1)) // Exponential backoff
					continue                                                // Retry
				}
			}
			logger.Error(err, "Failed to update Fanet status")
			return err
		}
		logger.Info("Fanet status updated", "name", fanet.Name)
		return nil // Success
	}
	return nil
}

// findFanetForDrone returns a reconcile request for a Fanet when its Drone changes
func (r *FanetReconciler) findFanetForDrone(ctx context.Context, obj client.Object) []ctrl.Request {
	drone, ok := obj.(*fanetv1alpha1.Drone)
	if !ok {
		return []ctrl.Request{}
	}

	// If the drone has a FanetRef in its status, reconcile that Fanet
	if drone.Status.FanetRef != nil && drone.Status.FanetRef.Name != "" && drone.Status.FanetRef.Name != FanetNotAssigned {
		return []ctrl.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      drone.Status.FanetRef.Name,
					Namespace: drone.Namespace,
				},
			},
		}
	}

	// Otherwise, check all Fanets to see if any reference this drone
	fanetList := &fanetv1alpha1.FanetList{}
	if err := r.List(ctx, fanetList, client.InNamespace(drone.Namespace)); err != nil {
		return []ctrl.Request{}
	}

	for _, fanet := range fanetList.Items {
		for _, droneRef := range fanet.Spec.Drones {
			if droneRef.Name == drone.Name {
				return []ctrl.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      fanet.Name,
							Namespace: fanet.Namespace,
						},
					},
				}
			}
		}
	}

	return []ctrl.Request{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *FanetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fanetv1alpha1.Fanet{}).
		Watches(&fanetv1alpha1.Drone{},
			handler.EnqueueRequestsFromMapFunc(r.findFanetForDrone)).
		Named("fanet").
		Complete(r)
}

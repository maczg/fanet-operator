package controller

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fanetv1alpha1 "github.com/maczg/fanet-operator/api/v1alpha1"
)

const (
	VFfinalizerName = "virtualfunction.finalizer"
)

// VirtualFunctionReconciler reconciles a VirtualFunction object
type VirtualFunctionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fanet.restart.io,resources=virtualfunctions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fanet.restart.io,resources=virtualfunctions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fanet.restart.io,resources=virtualfunctions/finalizers,verbs=update
// +kubebuilder:rbac:groups=fanet.restart.io,resources=drones,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=fanet.restart.io,resources=drones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

func (r *VirtualFunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	// Fetch the VirtualFunction instance
	vf := &fanetv1alpha1.VirtualFunction{}
	if err := r.Get(ctx, req.NamespacedName, vf); err != nil {
		logger.Error(err, "Failed to get VirtualFunction")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !vf.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleVFDeletion(ctx, vf)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(vf, VFfinalizerName) {
		controllerutil.AddFinalizer(vf, VFfinalizerName)
		if err := r.Update(ctx, vf); err != nil {
			return ctrl.Result{}, err
		}
	}

	// check if migration is needed
	oldDroneRef := vf.Status.LastDroneRef
	newDroneRef := vf.Spec.DroneRef.Name

	if oldDroneRef != "" && oldDroneRef != newDroneRef {
		logger.Info("Migration needed", "from", oldDroneRef, "to", newDroneRef)
		if err := r.handleVFMigration(ctx, vf, oldDroneRef, newDroneRef); err != nil {
			logger.Error(err, "Failed to migrate VF")
			return ctrl.Result{}, err
		}
	}

	// Create or update the pod
	if err := r.ensurePod(ctx, vf); err != nil {
		logger.Error(err, "Failed to ensure VF Pod", "vf", vf.Name)
		return ctrl.Result{}, err
	}

	// Update VF count on drone
	if err := r.updateDroneVFCount(ctx, vf.Spec.DroneRef.Name, vf.Namespace, 1); err != nil {
		logger.Error(err, "Failed to update Drone VF count", "drone", vf.Spec.DroneRef.Name)
		return ctrl.Result{}, err
	}

	// update Status
	if err := r.updateVFStatus(ctx, vf); err != nil {
		logger.Error(err, "Failed to update VF status", "vf", vf.Name)
		return ctrl.Result{}, err
	}

	// Store current drone ref for migration detection
	vf.Status.LastDroneRef = vf.Spec.DroneRef.Name
	if err := r.Status().Update(ctx, vf); err != nil {
		logger.Error(err, "Failed to update VF status with LastDroneRef", "vf", vf.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30}, nil
}

func (r *VirtualFunctionReconciler) handleVFDeletion(ctx context.Context, vf *fanetv1alpha1.VirtualFunction) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if controllerutil.ContainsFinalizer(vf, VFfinalizerName) {
		// Delete the pod
		pod := &corev1.Pod{}
		podName := fmt.Sprintf("vf-%s", vf.Name)
		err := r.Get(ctx, types.NamespacedName{Name: podName, Namespace: vf.Namespace}, pod)
		if err == nil {
			if err := r.Delete(ctx, pod); err != nil && !errors.IsNotFound(err) {
				log.Error(err, "Failed to delete pod")
				return ctrl.Result{}, err
			}
		}

		// Update drone VF count
		if vf.Spec.DroneRef != nil {
			if err := r.updateDroneVFCount(ctx, vf.Spec.DroneRef.Name, vf.Namespace, -1); err != nil {
				log.Error(err, "Failed to update drone VF count")
			}
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(vf, VFfinalizerName)
		if err := r.Update(ctx, vf); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *VirtualFunctionReconciler) updateDroneVFCount(ctx context.Context, droneName, namespace string, delta int) error {
	drone := &fanetv1alpha1.Drone{}
	err := r.Get(ctx, types.NamespacedName{Name: droneName, Namespace: namespace}, drone)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	drone.Status.VFCount = drone.Status.VFCount + delta
	if drone.Status.VFCount < 0 {
		drone.Status.VFCount = 0
	}

	return r.Status().Update(ctx, drone)
}

func (r *VirtualFunctionReconciler) handleVFMigration(ctx context.Context, vf *fanetv1alpha1.VirtualFunction, oldDrone, newDrone string) error {
	log := logf.FromContext(ctx)

	// Delete the existing pod
	podName := fmt.Sprintf("vf-%s", vf.Name)
	pod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: podName, Namespace: vf.Namespace}, pod)
	if err == nil {
		if err := r.Delete(ctx, pod); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete pod during migration: %w", err)
		}

		// Wait for pod deletion
		// TODO: Implement proper wait with backoff instead of sleep
		time.Sleep(2 * time.Second)
	}

	// Update VF counts
	if err := r.updateDroneVFCount(ctx, oldDrone, vf.Namespace, -1); err != nil {
		log.Error(err, "Failed to decrement old drone VF count")
	}

	log.Info("VF migration completed", "vf", vf.Name)
	return nil
}

func (r *VirtualFunctionReconciler) ensurePod(ctx context.Context, vf *fanetv1alpha1.VirtualFunction) error {
	log := logf.FromContext(ctx)

	// Get the drone to find the node
	drone := &fanetv1alpha1.Drone{}
	err := r.Get(ctx, types.NamespacedName{Name: vf.Spec.DroneRef.Name, Namespace: vf.Namespace}, drone)
	if err != nil {
		return fmt.Errorf("failed to get drone: %w", err)
	}

	if drone.Spec.NodeRef == nil {
		return fmt.Errorf("drone %s has no node reference", drone.Name)
	}

	// Check if pod exists
	podName := fmt.Sprintf("vf-%s", vf.Name)
	pod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: podName, Namespace: vf.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create pod
			pod = r.createPodForVF(vf, drone.Spec.NodeRef.Name)
			if err := r.Create(ctx, pod); err != nil {
				return fmt.Errorf("failed to create pod: %w", err)
			}
			log.Info("Created pod for VF", "vf", vf.Name, "pod", podName, "node", drone.Spec.NodeRef.Name)
		} else {
			return err
		}
	}
	return nil
}

func (r *VirtualFunctionReconciler) createPodForVF(vf *fanetv1alpha1.VirtualFunction, nodeName string) *corev1.Pod {
	podName := fmt.Sprintf("vf-%s", vf.Name)
	labels := map[string]string{
		"app": "virtualfunction",
		"vf":  vf.Name,
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: vf.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			Containers: []corev1.Container{
				{
					Name:      "vf-container",
					Image:     vf.Spec.Image,
					Args:      vf.Spec.Args,
					Resources: vf.Spec.Resources,
				},
			},
		},
	}

	// Set VF as owner
	_ = controllerutil.SetControllerReference(vf, pod, r.Scheme)
	return pod
}

func (r *VirtualFunctionReconciler) updateVFStatus(ctx context.Context, vf *fanetv1alpha1.VirtualFunction) error {
	log := logf.FromContext(ctx)

	// Check pod status
	podName := fmt.Sprintf("vf-%s", vf.Name)
	pod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: podName, Namespace: vf.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			if vf.Status.StatusPODVF == "Running" {
				vf.Status.StatusPODVF = "Missing"
			} else if vf.Status.StatusPODVF == "Missing" {
				// Delete VF if pod is missing for too long
				log.Info("Pod still missing, deleting VF", "vf", vf.Name)
				return r.Delete(ctx, vf)
			} else {
				vf.Status.StatusPODVF = "Missing"
			}
		} else {
			return err
		}
	} else {
		vf.Status.Phase = string(pod.Status.Phase)
		vf.Status.StatusPODVF = "Running"
	}

	vf.Status.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	if err := r.Status().Update(ctx, vf); err != nil {
		log.Error(err, "Failed to update VF status")
		return err
	}

	log.Info("Updated VF status", "vf", vf.Name, "phase", vf.Status.Phase)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualFunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fanetv1alpha1.VirtualFunction{}).
		Owns(&corev1.Pod{}).
		Named("virtualfunction").
		Complete(r)
}

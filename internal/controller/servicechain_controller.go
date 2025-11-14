package controller

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fanetv1alpha1 "github.com/maczg/fanet-operator/api/v1alpha1"
)

const (
	finalizerName = "servicechain.finalizer"
)

// ServiceChainReconciler reconciles a ServiceChain object
type ServiceChainReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fanet.restart.io,resources=servicechains,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fanet.restart.io,resources=servicechains/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fanet.restart.io,resources=servicechains/finalizers,verbs=update
// +kubebuilder:rbac:groups=fanet.restart.io,resources=virtualfunctions,verbs=get;list;watch;create;update;patch;delete

func (r *ServiceChainReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// Fetch the ServiceChain instance
	sc := &fanetv1alpha1.ServiceChain{}
	if err := r.Get(ctx, req.NamespacedName, sc); err != nil {
		logger.Error(err, "Failed to get ServiceChain")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !sc.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleSCDeletion(ctx, sc)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(sc, finalizerName) {
		controllerutil.AddFinalizer(sc, finalizerName)
		if err := r.Update(ctx, sc); err != nil {
			return ctrl.Result{}, err
		}
	}
	// Create VirtualFunctions for each function in the chain
	if err := r.ensureVirtualFunctions(ctx, sc); err != nil {
		logger.Error(err, "Failed to ensure VirtualFunctions for ServiceChain", "servicechain", sc.Name)
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateSCStatus(ctx, sc); err != nil {
		logger.Error(err, "Failed to update ServiceChain status", "servicechain", sc.Name)
		return ctrl.Result{}, err
	}

	// requeue after 30 seconds
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *ServiceChainReconciler) handleSCDeletion(ctx context.Context, sc *fanetv1alpha1.ServiceChain) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	if controllerutil.ContainsFinalizer(sc, finalizerName) {
		// Delete associated VirtualFunctions
		vfList := &fanetv1alpha1.VirtualFunctionList{}
		if err := r.List(ctx, vfList, client.InNamespace(sc.Namespace)); err != nil {
			logger.Error(err, "Failed to list VirtualFunctions")
			return ctrl.Result{}, err
		}
		for _, vf := range vfList.Items {
			if vf.Spec.ServiceChainName == sc.Name {
				if err := r.Delete(ctx, &vf); err != nil {
					logger.Error(err, "Failed to delete VirtualFunction", "virtualfunction", vf.Name)
					return ctrl.Result{}, err
				}
				logger.Info("VirtualFunction deleted", "virtualfunction", vf.Name)
			}
		}
		// Remove finalizer
		controllerutil.RemoveFinalizer(sc, finalizerName)
		if err := r.Update(ctx, sc); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ServiceChainReconciler) ensureVirtualFunctions(ctx context.Context, sc *fanetv1alpha1.ServiceChain) error {
	logger := logf.FromContext(ctx)
	for _, function := range sc.Spec.Functions {
		vfName := fmt.Sprintf("vf-%s-%s", sc.Name, function.Name)
		// Check if VirtualFunction already exists
		vf := &fanetv1alpha1.VirtualFunction{}
		err := r.Get(ctx, client.ObjectKey{Namespace: sc.Namespace, Name: vfName}, vf)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				logger.Error(err, "Failed to get VirtualFunction", "virtualfunction", vfName)
				return err
			}
			// Create new VirtualFunction
			vf = &fanetv1alpha1.VirtualFunction{
				ObjectMeta: ctrl.ObjectMeta{
					Name:      vfName,
					Namespace: sc.Namespace,
				},
				Spec: fanetv1alpha1.VirtualFunctionSpec{
					Image:                     function.Image,
					Args:                      function.Args,
					DroneRef:                  function.DroneRef,
					ServiceChainName:          sc.Name,
					Resources:                 function.Resources,
					IdleEnergyConsumptionWh:   function.IdleEnergyConsumptionWh,
					ActiveEnergyConsumptionWh: function.ActiveEnergyConsumptionWh,
					MaxLatencyMs:              function.MaxLatencyMs,
				},
			}

			// Set owner reference
			if err := controllerutil.SetControllerReference(sc, vf, r.Scheme); err != nil {
				logger.Error(err, "Failed to set owner reference for VirtualFunction", "virtualfunction", vfName)
				return fmt.Errorf("failed to set controller reference: %w", err)
			}
			if err := r.Create(ctx, vf); err != nil {
				return fmt.Errorf("failed to create VirtualFunction %s: %w", vfName, err)
			}
			logger.Info("Created VirtualFunction", "vf", vfName, "servicechain", sc.Name)
		}
	}
	return nil
}

func (r *ServiceChainReconciler) updateSCStatus(ctx context.Context, sc *fanetv1alpha1.ServiceChain) error {
	logger := logf.FromContext(ctx)
	// List all VirtualFunctions associated with this ServiceChain
	vfList := &fanetv1alpha1.VirtualFunctionList{}
	if err := r.List(ctx, vfList, client.InNamespace(sc.Namespace)); err != nil {
		logger.Error(err, "Failed to list VirtualFunctions for status update", "servicechain", sc.Name)
		return err
	}

	vfCount := 0
	totalLatencyMs := 0
	for _, vf := range vfList.Items {
		if vf.Spec.ServiceChainName == sc.Name {
			vfCount++
			totalLatencyMs += vf.Spec.MaxLatencyMs
		}
	}
	sc.Status.VFCount = vfCount
	sc.Status.TotalVFLatency = totalLatencyMs
	if err := r.Status().Update(ctx, sc); err != nil {
		logger.Error(err, "Failed to update ServiceChain status", "servicechain", sc.Name)
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceChainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fanetv1alpha1.ServiceChain{}).
		Owns(&fanetv1alpha1.VirtualFunction{}).
		Named("servicechain").
		Complete(r)
}

package controller

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fanetv1alpha1 "github.com/maczg/fanet-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

// ServiceChainReconciler reconciles a ServiceChain object
type ServiceChainReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fanet.restart.eu,resources=servicechains,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fanet.restart.eu,resources=servicechains/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fanet.restart.eu,resources=servicechains/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ServiceChainReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Reconciling ServiceChain", "name", req.Name, "namespace", req.Namespace)

	var sc fanetv1alpha1.ServiceChain
	if err := r.Get(ctx, req.NamespacedName, &sc); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ServiceChain resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ServiceChain")
		return ctrl.Result{}, err
	}

	// Calculate chain metrics
	totalPropagationDelay := r.calculateTotalPropagationDelay(&sc)
	totalEnergyConsumption := r.calculateTotalEnergyConsumption(&sc)
	chainHealthy, healthMessage := r.checkChainHealth(&sc)

	// Update chain status
	sc.Status.TotalPropagationDelay = totalPropagationDelay
	sc.Status.TotalEnergyConsumption = totalEnergyConsumption
	sc.Status.LastUpdate = metav1.Now()

	// Determine chain state
	if !chainHealthy {
		sc.Status.State = "error"
		sc.Status.StatusMessage = healthMessage
	} else if totalPropagationDelay > sc.Spec.PerformanceMetrics.MaxTotalPropagationDelay {
		sc.Status.State = "warning"
		sc.Status.StatusMessage = fmt.Sprintf("Total propagation delay (%dms) exceeds maximum allowed (%dms)",
			totalPropagationDelay, sc.Spec.PerformanceMetrics.MaxTotalPropagationDelay)
	} else {
		sc.Status.State = "active"
		sc.Status.StatusMessage = "Service chain is healthy"
	}

	// Update function statuses
	var functionStatuses []fanetv1alpha1.FunctionStatus
	for _, function := range sc.Spec.Functions {
		var vf fanetv1alpha1.VirtualFunction
		if err := r.Get(context.Background(), types.NamespacedName{Name: function.Name}, &vf); err == nil {
			functionStatuses = append(functionStatuses, fanetv1alpha1.FunctionStatus{
				Name:                     function.Name,
				Status:                   vf.Status.State,
				Node:                     vf.Spec.UAVName,
				CurrentPropagationDelay:  vf.Status.CurrentPropagationDelay,
				CurrentEnergyConsumption: vf.Status.CurrentEnergyConsumption,
			})
		}
	}
	sc.Status.Functions = functionStatuses

	if err := r.Status().Update(ctx, &sc); err != nil {
		logger.Error(err, "Failed to update ServiceChain status")
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("[%s] Total Delay=%dms | Total Energy=%sW | State=%s",
		sc.Name, totalPropagationDelay, totalEnergyConsumption, sc.Status.State))

	return ctrl.Result{}, nil
}

// calculateTotalPropagationDelay calculates the total propagation delay of the service chain
func (r *ServiceChainReconciler) calculateTotalPropagationDelay(sc *fanetv1alpha1.ServiceChain) int32 {
	var totalDelay int32

	// Get all VFs in the chain
	for _, function := range sc.Spec.Functions {
		var vf fanetv1alpha1.VirtualFunction
		if err := r.Get(context.Background(), types.NamespacedName{Name: function.Name}, &vf); err == nil {
			totalDelay += vf.Status.CurrentPropagationDelay
		}
	}

	return totalDelay
}

// calculateTotalEnergyConsumption calculates the total energy consumption of the service chain
func (r *ServiceChainReconciler) calculateTotalEnergyConsumption(sc *fanetv1alpha1.ServiceChain) string {
	var totalEnergy float64

	// Get all VFs in the chain
	for _, function := range sc.Spec.Functions {
		var vf fanetv1alpha1.VirtualFunction
		if err := r.Get(context.Background(), types.NamespacedName{Name: function.Name}, &vf); err == nil {
			// Parse the energy consumption string to float
			if energy, err := strconv.ParseFloat(vf.Status.CurrentEnergyConsumption, 64); err == nil {
				totalEnergy += energy
			}
		}
	}

	// Format the total energy consumption as a string with 2 decimal places
	return fmt.Sprintf("%.2f", totalEnergy)
}

// checkChainHealth checks if all functions in the chain are healthy
func (r *ServiceChainReconciler) checkChainHealth(sc *fanetv1alpha1.ServiceChain) (bool, string) {
	for _, function := range sc.Spec.Functions {
		var vf fanetv1alpha1.VirtualFunction
		if err := r.Get(context.Background(), types.NamespacedName{Name: function.Name}, &vf); err != nil {
			return false, fmt.Sprintf("Function %s not found", function.Name)
		}

		if vf.Status.State != "active" {
			return false, fmt.Sprintf("Function %s is not active (state: %s)", function.Name, vf.Status.State)
		}
	}

	return true, ""
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceChainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fanetv1alpha1.ServiceChain{}).
		Named("servicechain").
		Complete(r)
}

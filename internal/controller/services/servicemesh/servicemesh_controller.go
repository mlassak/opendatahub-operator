package servicemesh

import (
	"context"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	serviceApi "github.com/opendatahub-io/opendatahub-operator/v2/api/services/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
)

type ServiceMeshReconciler struct {
	Client   client.Client
	Recorder record.EventRecorder
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceMeshReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	logf.FromContext(ctx).Info("Adding controller for ServiceMesh.")

	return ctrl.NewControllerManagedBy(mgr).
		Named("servicemesh-controller").
		For(&serviceApi.ServiceMesh{}).
		Complete(r)
}

func (r *ServiceMeshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling ServiceMesh controller")

	sm := &serviceApi.ServiceMesh{}
	if err := r.Client.Get(ctx, req.NamespacedName, sm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dsci, err := cluster.GetDSCI(ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !sm.DeletionTimestamp.IsZero() {
		// ServiceMesh instance is being deleted
		if err := r.removeServiceMesh(ctx, dsci); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	// Apply Service Mesh configurations
	if err := r.configureServiceMesh(ctx, dsci); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

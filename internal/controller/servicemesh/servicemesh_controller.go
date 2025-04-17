package servicemesh

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v1"
)

type ServiceMeshReconciler struct {
	client.Client
	Recorder record.EventRecorder
}

var dsciServiceMeshPredicate = predicate.Funcs{
	CreateFunc: func(e event.CreateEvent) bool {
		dsci, ok := e.Object.(*dsciv1.DSCInitialization)
		if !ok {
			return false
		}

		return dsci.Spec.ServiceMesh != nil
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		dsci, ok := e.Object.(*dsciv1.DSCInitialization)
		if !ok {
			return false
		}

		return dsci.Spec.ServiceMesh != nil
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldDsci, ok := e.ObjectOld.(*dsciv1.DSCInitialization)
		if !ok {
			return false
		}
		newDsci, ok := e.ObjectNew.(*dsciv1.DSCInitialization)
		if !ok {
			return false
		}

		return !reflect.DeepEqual(oldDsci.Spec.ServiceMesh, newDsci.Spec.ServiceMesh)
	},
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceMeshReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	logf.FromContext(ctx).Info("Adding controller for ServiceMesh.")

	return ctrl.NewControllerManagedBy(mgr).
		Named("servicemesh-controller").
		Watches(
			&dsciv1.DSCInitialization{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				return []ctrl.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.GetName(),
							Namespace: obj.GetNamespace(),
						},
					},
				}
			}),
			builder.WithPredicates(dsciServiceMeshPredicate),
		).
		Complete(r)
}

func (r *ServiceMeshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling ServiceMesh controller")

	dsciInstance := &dsciv1.DSCInitialization{}
	if err := r.Client.Get(ctx, req.NamespacedName, dsciInstance); err != nil {
		return ctrl.Result{}, err
	}

	if !dsciInstance.ObjectMeta.DeletionTimestamp.IsZero() {
		// DSCI is being deleted, remove ServiceMesh
		if err := r.removeServiceMesh(ctx, dsciInstance); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Apply Service Mesh configurations
	if errServiceMesh := r.configureServiceMesh(ctx, dsciInstance); errServiceMesh != nil {
		return ctrl.Result{}, errServiceMesh
	}

	return ctrl.Result{}, nil
}

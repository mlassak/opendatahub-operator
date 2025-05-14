package dscinitialization

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v1"
	serviceApi "github.com/opendatahub-io/opendatahub-operator/v2/api/services/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/resources"
)

func (r *DSCInitializationReconciler) handleServiceMesh(ctx context.Context, dscInit *dsciv1.DSCInitialization) error {
	log := logf.FromContext(ctx)

	if dscInit.Spec.ServiceMesh != nil && dscInit.Spec.ServiceMesh.ManagementState == operatorv1.Managed {
		desiredServiceMesh := &serviceApi.ServiceMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceApi.ServiceMeshInstanceName,
			},
			Spec: serviceApi.ServiceMeshSpec{
				ManagementState: dscInit.Spec.ServiceMesh.ManagementState, // TODO do not copy this over, do it as with components
				ControlPlane:    serviceApi.ServiceMeshControlPlaneSpec(*dscInit.Spec.ServiceMesh.ControlPlane.DeepCopy()),
				Auth:            serviceApi.ServiceMeshAuthSpec(*dscInit.Spec.ServiceMesh.Auth.DeepCopy()),
			},
		}

		// TODO update DSCI with ServiceMesh conditions

		if err := controllerutil.SetControllerReference(dscInit, desiredServiceMesh, r.Client.Scheme()); err != nil {
			return err
		}

		err := resources.Apply(
			ctx,
			r.Client,
			desiredServiceMesh,
			client.FieldOwner(fieldManager),
			client.ForceOwnership,
		)
		if err != nil && !k8serr.IsAlreadyExists(err) {
			return err
		}
	} else if dscInit.Spec.ServiceMesh.ManagementState == operatorv1.Removed { // TODO check and possibly handle Unmanaged state
		// ServiceMesh not defined in DSCI -> remove ServiceMesh instance if it exists
		sm := &serviceApi.ServiceMesh{}
		err := r.Client.Get(ctx, types.NamespacedName{Name: serviceApi.ServiceMeshInstanceName}, sm)

		if err == nil {
			log.Info("deleting ServiceMesh instance")
			propagationPolicy := metav1.DeletePropagationForeground
			if err := r.Client.Delete(
				ctx,
				sm,
				&client.DeleteOptions{
					PropagationPolicy: &propagationPolicy,
				},
			); err != nil {
				if !k8serr.IsNotFound(err) {
					log.Error(err, "failed to delete ServiceMesh instance")
					return err
				}
			} else {
				log.Info("ServiceMesh instance deleted successfully")
			}
		} else if !k8serr.IsNotFound(err) {
			log.Error(err, "failed to get ServiceMesh instance during deletion scenario")
			return err
		}
	}

	return nil
}

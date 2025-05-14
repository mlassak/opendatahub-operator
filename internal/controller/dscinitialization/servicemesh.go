package dscinitialization

import (
	"context"
	"reflect"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	dsciv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/dscinitialization/v1"
	serviceApi "github.com/opendatahub-io/opendatahub-operator/v2/api/services/v1alpha1"
)

func (r *DSCInitializationReconciler) handleServiceMesh(ctx context.Context, dscInit *dsciv1.DSCInitialization) error {
	log := logf.FromContext(ctx)

	sm := &serviceApi.ServiceMesh{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: serviceApi.ServiceMeshInstanceName}, sm)

	// ServiceMesh configuration
	if dscInit.Spec.ServiceMesh != nil {
		desiredServiceMeshSpec := &serviceApi.ServiceMeshSpec{
			ManagementState: dscInit.Spec.ServiceMesh.ManagementState,
			ControlPlane:    serviceApi.ServiceMeshControlPlaneSpec(*dscInit.Spec.ServiceMesh.ControlPlane.DeepCopy()),
			Auth:            serviceApi.ServiceMeshAuthSpec(*dscInit.Spec.ServiceMesh.Auth.DeepCopy()),
		}

		switch {
		case err == nil:
			// ServiceMesh instance already exists -> update if changes happened
			smSpecChanged := false
			if !reflect.DeepEqual(sm.Spec, desiredServiceMeshSpec) {
				smSpecChanged = true
				sm.Spec = *desiredServiceMeshSpec
			}

			if err := controllerutil.SetControllerReference(dscInit, sm, r.Client.Scheme()); err != nil {
				isOwnedByDSCI, ownerRefCheckError := controllerutil.HasOwnerReference(sm.OwnerReferences, dscInit, r.Client.Scheme())
				if !isOwnedByDSCI {
					log.Error(err, "failed to ensure DSCI owner reference on ServiceMesh")
					return err
				}
				if ownerRefCheckError != nil {
					log.Error(ownerRefCheckError, "error ensuring owner reference on ServiceMesh")
					return ownerRefCheckError
				}
			}

			if smSpecChanged {
				log.Info("Updating ServiceMesh instance", "name", serviceApi.ServiceMeshInstanceName)
				if err := r.Client.Update(ctx, sm); err != nil {
					log.Error(err, "failed to update ServiceMesh instance")
					return err
				}

				log.Info("ServiceMesh instance updated successfully")
			}
		case err != nil && k8serr.IsNotFound(err):
			// ServiceMesh instance does not exist -> create
			log.Info("creating ServiceMesh instance", "name", serviceApi.ServiceMeshInstanceName)

			newServiceMesh := &serviceApi.ServiceMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name: serviceApi.ServiceMeshInstanceName,
				},
				Spec: *desiredServiceMeshSpec.DeepCopy(),
			}

			if err := controllerutil.SetControllerReference(dscInit, newServiceMesh, r.Client.Scheme()); err != nil {
				log.Error(err, "failed to set owner reference on ServiceMesh")
				return err
			}

			if err := r.Client.Create(ctx, newServiceMesh); err != nil {
				log.Error(err, "failed to create ServiceMesh instance")
				return err
			}

			log.Info("Successfully created ServiceMesh instance")
		default:
			log.Error(err, "failed to get ServiceMesh instance")
			return err
		}
	} else {
		// ServiceMesh not defined in DSCI -> remove ServiceMesh instance if it exists
		if err == nil {
			log.Info("deleting ServiceMesh instance")
			if err := r.Client.Delete(ctx, sm); err != nil {
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

package servicemesh

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	serviceApi "github.com/opendatahub-io/opendatahub-operator/v2/api/services/v1alpha1"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/status"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	odhtypes "github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/types"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/labels"
)

// TODO add owner references to ServiceMesh CR or DSCI for the created resources (to replace FeatureTracker owner refs).
func checkPreconditions(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	log := logf.FromContext(ctx)

	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	// ensure ServiceMesh v2 operator is installed as pre-requisite
	if err := checkServiceMeshOperator(ctx, rr); err != nil {
		conditions := sm.Status.Conditions
		status.SetCondition(
			&conditions,
			status.CapabilityServiceMesh,
			status.MissingOperatorReason,
			"OpenShift ServiceMesh v2 operator not found / not setup properly on the cluster, cannot setup ServiceMesh",
			metav1.ConditionFalse,
		)
		status.SetCondition(
			&conditions,
			status.CapabilityServiceMeshAuthorization,
			status.MissingOperatorReason,
			"OpenShift ServiceMesh v2 operator not found / not setup properly on the cluster, cannot setup ServiceMesh Authorization",
			metav1.ConditionFalse,
		)
		sm.Status.SetConditions(conditions)
		if err := rr.Client.Status().Update(ctx, sm); err != nil {
			log.Error(err, "failed to update ServiceMesh status conditions")
			return err
		}

		return errors.New("OpenShift ServiceMesh v2 operator not found / not setup properly on the cluster, failed to setup ServiceMesh v2 resources")
	}

	return nil
}

func checkServiceMeshOperator(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	if smOperatorFound, err := cluster.SubscriptionExists(ctx, rr.Client, serviceMeshOperatorName); !smOperatorFound || err != nil {
		return fmt.Errorf(
			"failed to find the pre-requisite operator subscription %q, please ensure operator is installed. %w",
			serviceMeshOperatorName,
			fmt.Errorf("missing operator %q", serviceMeshOperatorName),
		)
	}

	if err := cluster.CustomResourceDefinitionExists(ctx, rr.Client, gvk.ServiceMeshControlPlane.GroupKind()); err != nil {
		return fmt.Errorf("failed to find the Service Mesh Control Plane CRD, please ensure Service Mesh Operator is installed. %w", err)
	}

	// Extra check smcp validation service is running.
	validationService := &corev1.Service{}
	if err := rr.Client.Get(ctx, client.ObjectKey{
		Name:      "istio-operator-service",
		Namespace: "openshift-operators",
	}, validationService); err != nil {
		if k8serr.IsNotFound(err) {
			return fmt.Errorf("failed to find the Service Mesh VWC service, please ensure Service Mesh Operator is running. %w", err)
		}
		return fmt.Errorf("failed to find the Service Mesh VWC service. %w", err)
	}

	return nil
}

func createControlPlaneNamespace(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	// ensure SMCP namespace exists
	if _, err := cluster.CreateNamespace(ctx, rr.Client, sm.Spec.ControlPlane.Namespace); err != nil {
		return errors.New("error creating SMCP namespace")
	}

	return nil
}

func initializeServiceMesh(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	rr.Templates = append(
		rr.Templates,
		odhtypes.TemplateInfo{
			FS:   resourcesFS,
			Path: serviceMeshControlPlaneTemplate,
		},
	)

	return nil
}

func initializeServiceMeshMetricsCollection(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	log := logf.FromContext(ctx)

	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	if sm.Spec.ControlPlane.MetricsCollection != "Istio" {
		log.Info("MetricsCollection not set to Istion, skipping ServiceMesh metrics collection configuration")
		return nil
	}

	rr.Templates = append(
		rr.Templates,
		odhtypes.TemplateInfo{
			FS:   resourcesFS,
			Path: podMonitorTemplate,
		},
		odhtypes.TemplateInfo{
			FS:   resourcesFS,
			Path: serviceMonitorTemplate,
		},
	)

	return nil
}

func initializeAuthorino(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	log := logf.FromContext(ctx)

	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	// ensure Authorino operator is installed as pre-requisite
	authorinoOperatorFound, err := cluster.SubscriptionExists(ctx, rr.Client, authorinoOperatorName)
	if err != nil {
		return err
	}
	if !authorinoOperatorFound {
		log.Info("Authorino operator not found on the cluster, skipping authorization capability")

		conditions := sm.Status.Conditions
		status.SetCondition(
			&conditions,
			status.CapabilityServiceMeshAuthorization,
			status.MissingOperatorReason,
			"Authorino operator is not installed on the cluster, skipping authorization capability",
			metav1.ConditionFalse,
		)
		sm.Status.SetConditions(conditions)
		if err := rr.Client.Status().Update(ctx, sm); err != nil {
			log.Error(err, "failed to update ServiceMesh status conditions")
			return err
		}

		return nil
	}

	// create authorino namespace if it does not exist
	if _, err := cluster.CreateNamespace(
		ctx,
		rr.Client,
		sm.Spec.Auth.Namespace,
		cluster.OwnedBy(rr.DSCI, rr.Client.Scheme()), // TODO check if this ownership is desired.
		cluster.WithLabels(labels.ODH.OwnedNamespace, "true"),
	); err != nil {
		return errors.New("error creating Authorino namespace")
	}

	rr.Templates = append(
		rr.Templates,
		odhtypes.TemplateInfo{
			FS:   resourcesFS,
			Path: authorinoTemplate,
		},
		odhtypes.TemplateInfo{
			FS:   resourcesFS,
			Path: authorinoServiceMeshMemberTemplate,
		},
		odhtypes.TemplateInfo{
			FS:   resourcesFS,
			Path: authorinoDeploymentInjectionTemplate,
		},
		odhtypes.TemplateInfo{
			FS:   resourcesFS,
			Path: authorinoServiceMeshControlPlaneTemplate,
		},
	)

	return nil
}

func getTemplateData(ctx context.Context, rr *odhtypes.ReconciliationRequest) (map[string]any, error) {
	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return nil, fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	return map[string]any{
		"AuthExtensionName": sm.Spec.Auth.Namespace,
		"AuthNamespace":     sm.Spec.Auth.Namespace,
		"AuthProviderName":  authProviderName,
		"ControlPlane":      sm.Spec.ControlPlane,
	}, nil
}

// TODO maybe use rr.AddResources.
func updateMeshRefsConfigMap(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	data := map[string]string{
		"CONTROL_PLANE_NAME": sm.Spec.ControlPlane.Name,
		"MESH_NAMESPACE":     sm.Spec.ControlPlane.Namespace,
	}

	if err := cluster.CreateOrUpdateConfigMap(
		ctx,
		rr.Client,
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapMeshRefName,
				Namespace: rr.DSCI.Spec.ApplicationsNamespace,
			},
			Data: data,
		},
		cluster.OwnedBy(rr.DSCI, rr.Client.Scheme()), // TODO check who should own this, used to be FeatureTracker
	); err != nil {
		return fmt.Errorf("failed to update %s ConfigMap", configMapMeshRefName)
	}

	return nil
}

// TODO maybe use rr.AddResources.
func updateAuthRefsConfigMap(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	audiences := sm.Spec.Auth.Audiences
	audiencesList := ""
	if audiences != nil && len(*audiences) > 0 {
		audiencesList = strings.Join(*audiences, ",")
	}

	data := map[string]string{
		"AUTH_AUDIENCE":   audiencesList,
		"AUTH_PROVIDER":   authProviderName,
		"AUTH_NAMESPACE":  sm.Spec.Auth.Namespace,
		"AUTHORINO_LABEL": authorinoLabel,
	}

	if err := cluster.CreateOrUpdateConfigMap(
		ctx,
		rr.Client,
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapAuthRefName,
				Namespace: rr.DSCI.Spec.ApplicationsNamespace,
			},
			Data: data,
		},
		cluster.OwnedBy(rr.DSCI, rr.Client.Scheme()), // TODO check who should own this, used to be FeatureTracker
	); err != nil {
		return fmt.Errorf("failed to update %s ConfigMap", configMapAuthRefName)
	}

	return nil
}

func updateStatus(ctx context.Context, rr *odhtypes.ReconciliationRequest) error {
	log := logf.FromContext(ctx)

	sm, ok := rr.Instance.(*serviceApi.ServiceMesh)
	if !ok {
		return fmt.Errorf("resource instance %v is not a serviceApi.ServiceMesh)", rr.Instance)
	}

	log.Info("ServiceMesh CR is configured and managed by the operator")

	conditions := sm.Status.Conditions
	status.SetCondition(
		&conditions,
		status.CapabilityServiceMesh,
		status.ConfiguredReason,
		"ServiceMesh configured",
		metav1.ConditionTrue,
	)

	authorinoOperatorFound, err := cluster.SubscriptionExists(ctx, rr.Client, authorinoOperatorName)
	if err != nil {
		return err
	}
	if authorinoOperatorFound {
		status.SetCondition(
			&conditions,
			status.CapabilityServiceMeshAuthorization,
			status.ConfiguredReason,
			"ServiceMesh authorization configured",
			metav1.ConditionTrue,
		)
	}

	sm.Status.SetConditions(conditions)
	if err := rr.Client.Status().Update(ctx, sm); err != nil {
		log.Error(err, "failed to update ServiceMesh status conditions")
		return err
	}

	return nil
}

package workbenches

import (
	"context"
	"errors"
	"fmt"

	operatorv1 "github.com/openshift/api/operator/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/opendatahub-io/opendatahub-operator/v2/api/common"
	componentApi "github.com/opendatahub-io/opendatahub-operator/v2/api/components/v1alpha1"
	dscv1 "github.com/opendatahub-io/opendatahub-operator/v2/api/datasciencecluster/v1"
	cr "github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/components/registry"
	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/status"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/conditions"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/types"
	odhdeploy "github.com/opendatahub-io/opendatahub-operator/v2/pkg/deploy"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"
)

type componentHandler struct{}

func init() { //nolint:gochecknoinits
	cr.Add(&componentHandler{})
}

func (s *componentHandler) GetName() string {
	return componentApi.WorkbenchesComponentName
}

func (s *componentHandler) NewCRObject(dsc *dscv1.DataScienceCluster) common.PlatformObject {
	return &componentApi.Workbenches{
		TypeMeta: metav1.TypeMeta{
			Kind:       componentApi.WorkbenchesKind,
			APIVersion: componentApi.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentApi.WorkbenchesInstanceName,
			Annotations: map[string]string{
				annotations.ManagementStateAnnotation: string(dsc.Spec.Components.Workbenches.ManagementState),
			},
		},
		Spec: componentApi.WorkbenchesSpec{
			WorkbenchesCommonSpec: dsc.Spec.Components.Workbenches.WorkbenchesCommonSpec,
		},
	}
}

func (s *componentHandler) Init(platform common.Platform) error {
	nbcManifestInfo := notebookControllerManifestInfo(notebookControllerManifestSourcePath)
	if err := odhdeploy.ApplyParams(nbcManifestInfo.String(), map[string]string{
		"odh-notebook-controller-image": "RELATED_IMAGE_ODH_NOTEBOOK_CONTROLLER_IMAGE",
	}); err != nil {
		return fmt.Errorf("failed to update params.env from %s : %w", nbcManifestInfo.String(), err)
	}

	kfNbcManifestInfo := kfNotebookControllerManifestInfo(kfNotebookControllerManifestSourcePath)
	if err := odhdeploy.ApplyParams(kfNbcManifestInfo.String(), map[string]string{
		"odh-kf-notebook-controller-image": "RELATED_IMAGE_ODH_KF_NOTEBOOK_CONTROLLER_IMAGE",
	}); err != nil {
		return fmt.Errorf("failed to update params.env from %s : %w", kfNbcManifestInfo.String(), err)
	}

	return nil
}

func (s *componentHandler) IsEnabled(dsc *dscv1.DataScienceCluster) bool {
	return dsc.Spec.Components.Workbenches.ManagementState == operatorv1.Managed
}

func (s *componentHandler) UpdateDSCStatus(ctx context.Context, rr *types.ReconciliationRequest) (metav1.ConditionStatus, error) {
	cs := metav1.ConditionUnknown

	c := componentApi.Workbenches{}
	c.Name = componentApi.WorkbenchesInstanceName

	if err := rr.Client.Get(ctx, client.ObjectKeyFromObject(&c), &c); err != nil && !k8serr.IsNotFound(err) {
		return cs, nil
	}

	dsc, ok := rr.Instance.(*dscv1.DataScienceCluster)
	if !ok {
		return cs, errors.New("failed to convert to DataScienceCluster")
	}

	dsc.Status.InstalledComponents[LegacyComponentName] = false
	dsc.Status.Components.Workbenches.ManagementState = dsc.Spec.Components.Workbenches.ManagementState
	dsc.Status.Components.Workbenches.WorkbenchesCommonStatus = nil

	rr.Conditions.MarkFalse(ReadyConditionType)

	if s.IsEnabled(dsc) {
		dsc.Status.InstalledComponents[LegacyComponentName] = true
		dsc.Status.Components.Workbenches.WorkbenchesCommonStatus = c.Status.WorkbenchesCommonStatus.DeepCopy()

		if rc := conditions.FindStatusCondition(c.GetStatus(), status.ConditionTypeReady); rc != nil {
			rr.Conditions.MarkFrom(ReadyConditionType, *rc)
			cs = rc.Status
		} else {
			cs = metav1.ConditionFalse
		}
	} else {
		rr.Conditions.MarkFalse(
			ReadyConditionType,
			conditions.WithReason(string(dsc.Spec.Components.Workbenches.ManagementState)),
			conditions.WithMessage("Component ManagementState is set to %s", dsc.Spec.Components.Workbenches.ManagementState),
			conditions.WithSeverity(common.ConditionSeverityInfo),
		)
	}

	return cs, nil
}

/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trainer

import (
	"context"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	"github.com/opendatahub-io/opendatahub-operator/v2/internal/controller/status"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/cluster/gvk"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/precondition"
	odhtypes "github.com/opendatahub-io/opendatahub-operator/v2/pkg/controller/types"
)

func checkPreConditions(ctx context.Context, rr *odhtypes.ReconciliationRequest) (precondition.CheckResult, error) {
	jobSetInfo, err := cluster.OperatorExists(ctx, rr.Client, jobSetOperator)
	if err != nil {
		return precondition.CheckResult{}, err
	}

	if jobSetInfo == nil {
		return precondition.CheckResult{Pass: false, Message: status.JobSetOperatorNotInstalledMessage}, nil
	}

	jobSetOperatorCR := &unstructured.Unstructured{}
	jobSetOperatorCR.SetGroupVersionKind(gvk.JobSetOperatorV1)
	if err := rr.Client.Get(ctx, types.NamespacedName{Name: jobSetOperatorCRName}, jobSetOperatorCR); err != nil {
		if k8serr.IsNotFound(err) {
			return precondition.CheckResult{Pass: false, Message: status.JobSetOperatorCRNotFoundMessage}, nil
		}
		return precondition.CheckResult{}, err
	}

	return precondition.CheckResult{Pass: true}, nil
}

func initialize(_ context.Context, rr *odhtypes.ReconciliationRequest) error { //nolint:unparam
	rr.Manifests = append(rr.Manifests, manifestPath(rr.ManifestsBasePath))
	return nil
}

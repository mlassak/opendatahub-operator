# RHAII RBAC Investigation Methodology

**Ticket:** RHOAIENG-54569
**Date:** 2026-04-15

Methodology used to determine the minimum required RBAC permissions for the RHAII operator on xKS (Azure).

---

## Phase 1: Static Analysis

### 1.1 Identify active controllers

Starting from the RHAII manager patch (`config/rhaii/odh-operator/manager_patch.yaml`), catalogued
which controllers, components, and services are enabled vs disabled via `RHAI_DISABLE_*` environment
variables. Result: only KServe is active; all other components and all services are disabled.

### 1.2 Map kubebuilder RBAC markers to controllers

Read `internal/controller/datasciencecluster/kubebuilder_rbac.go` and mapped each `+kubebuilder:rbac`
marker to the controller/component that requires it. Separated rules into three categories:
- Rules required by KServe
- Rules required by the operator framework (shared infrastructure)
- Rules required only by disabled components (candidates for removal)

### 1.3 Identify additional RBAC sources

Checked files beyond the main kubebuilder markers:
- `internal/controller/cloudmanager/common/kubebuilder_rbac.go` (CCM shared RBAC)
- `internal/controller/components/kserve/kserve_controller.go` (dynamic watches via `OwnsGVK`, `WatchesGVK`)
- `pkg/controller/actions/gc/action_gc.go` (GC runtime API calls)
- `pkg/rules/rules.go` (SelfSubjectRulesReview)
- `cmd/main.go` (startup code, cache setup, upgrade tasks)
- `internal/webhook/webhook.go` (webhook registration gated on component flags)
- `pkg/upgrade/upgrade.go` (unconditional startup cleanup)

### 1.4 Produce initial proposed ClusterRole

Combined the KServe-specific and operator framework rules into a proposed minimal ClusterRole
(`rhaii-rbac-analysis.md` Section 3).

### 1.5 Gap analysis

Cross-referenced the proposed ClusterRole against the actual codebase to identify missing
permissions, producing `rhaii-rbac-gap-analysis.md` with 10 identified gaps.

---

## Phase 2: Runtime Validation on KinD

### 2.1 Environment setup

1. Created a KinD cluster
2. Built the operator image (`make image-build`) and loaded it into KinD (`make image-kind-load`)
3. Deployed the RHAII operator with the full ClusterRole (`make deploy-rhaii-local`)
4. Created a placeholder webhook TLS secret (workaround for cert-manager chicken-and-egg)
5. Deployed the Azure Cloud Manager (`make deploy-ccm-local-azure`)
6. Set up pull secrets for `registry.redhat.io` (`make kind-setup-pull-secrets`)
7. Created the `AzureKubernetesEngine` CR to bootstrap dependencies (cert-manager, sail-operator, LWS, Gateway API)
8. Created the `Kserve` CR to trigger KServe reconciliation

See `rhaii-rbac-testing-guide.md` for full step-by-step instructions.

### 2.2 Baseline validation

Verified the full deployment is healthy under the default (full) ClusterRole:
- All operator pods running
- KServe CR `Ready=True`
- AzureKubernetesEngine CR `Ready=True`
- No RBAC errors in logs

### 2.3 Apply minimal ClusterRole

Replaced the operator's ClusterRole with the proposed minimal version from the static
analysis. The ClusterRole name must match the existing ClusterRoleBinding (kustomize
prefixes the namespace, so it was `opendatahub-operator-controller-manager-role`, not
the bare `controller-manager-role`).

### 2.4 Iterative error detection and correction

Restarted the operator after each ClusterRole change and checked:

```bash
# RBAC errors in all operator pods
for pod in $(kubectl -n opendatahub-operator-system get pods -o name); do
  kubectl -n opendatahub-operator-system logs $pod | \
    grep -i -E 'forbidden|cannot create|cannot get|cannot list|cannot watch|cannot update|cannot patch|cannot delete|unauthorized'
done

# KServe CR health
kubectl get kserve default-kserve -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Found and fixed: `monitoring.coreos.com/servicemonitors` missing (Gap 4 confirmed). The KServe
controller owns ServiceMonitors via `OwnsGVK`, and the cache starts watching them when the CRD
exists (installed by CCM's sail-operator).

### 2.5 Gap validation

Tested each gap from the static analysis against the live cluster:

- **Gap 1 (selfsubjectrulesreviews):** FALSE POSITIVE. Kubernetes implicitly allows any
  authenticated user to create self-subject reviews. Confirmed via
  `kubectl auth can-i create selfsubjectrulesreviews.authorization.k8s.io --as=system:serviceaccount:...`
  returning `yes` despite no explicit RBAC rule.

- **Gap 2 (hardwareprofiles):** NOT APPLICABLE on xKS. The HardwareProfile CRD is not deployed
  in RHAII, the MutatingWebhookConfiguration doesn't route to the handler, and the upgrade
  migration code skips gracefully when the CRD is absent.

- **Gap 4 (servicemonitors):** CONFIRMED. Added to the minimal ClusterRole.

### 2.6 Rule trimming

After the minimal ClusterRole was stable, systematically reviewed each remaining rule against:
1. KServe controller code (`internal/controller/components/kserve/`)
2. KServe XKS manifests (`opt/manifests/kserve/overlays/odh-xks/`)
3. Operator framework code (`cmd/main.go`, `pkg/`)
4. Actual resources deployed on the cluster

Removed rules with no code path or manifest reference on xKS, verifying each removal with
a restart + error check cycle:

| Rule removed | Reason |
|---|---|
| `authentication.k8s.io/tokenreviews` | Auth proxy sidecar not used in RHAII |
| `authorization.k8s.io/subjectaccessreviews` | Same as above |
| `controller-runtime.sigs.k8s.io/controllermanagerconfigs` | API group does not exist; kubebuilder marker artifact |
| `config.openshift.io/ingresses` | `GetDomain()` only called by disabled controllers; API absent on xKS |
| Core `endpoints` | No KServe code or manifests reference Endpoints |
| `apps/statefulsets` | No KServe code or manifests reference StatefulSets |
| `kubeflow.org/notebooks` | No KServe controller code references notebooks; marker mis-attributed |
| `kuadrant.io/kuadrants` | Only used by ModelsAsService controller (disabled); GVK defined but never referenced |
| `operator.openshift.io/jobsetoperators` | Only used by Trainer controller (disabled) |

Each removal was validated: apply ClusterRole, restart operator, confirm zero RBAC errors
and KServe `Ready=True`.

---

## Phase 3: Findings

### 3.1 Minimal ClusterRole

The validated minimal ClusterRole is in `xks-rhaii-minimal-clusterrole.yaml`. It was tested
on a live KinD cluster with:
- Full KServe reconciliation (create, delete, recreate cycle)
- CCM-managed dependencies running (cert-manager, sail-operator, LWS)
- GC action executing successfully
- Webhook server operational

### 3.2 Scope

The minimal ClusterRole covers the **RHAII operator only**. The CCM has a separate ClusterRole
(`opendatahub-azure-cloud-manager-role`) bound to a separate service account in a separate
namespace. Both are deployed independently and both are required for a functioning xKS deployment.

### 3.3 Static vs dynamic generation recommendation

**Static separate ClusterRole is recommended.** The kubebuilder markers are unreliable as a
source of truth for RHAII RBAC:
- Multiple markers under the `// Kserve` section are not used by KServe at runtime (notebooks,
  kuadrant, KEDA, metrics, jobsetoperators, endpoints, statefulsets, controllermanagerconfigs)
- Permissions needed at runtime are missing from markers (servicemonitors, added via `OwnsGVK`
  in controller code, not kubebuilder markers)
- Markers are not cleanly separated per component — the DSC controller file contains all
  component markers in one file

Dynamic generation from markers would produce a ClusterRole that is both too broad (includes
dead rules) and too narrow (misses runtime dependencies). The iteratively tested static
ClusterRole reflects actual runtime requirements.

### 3.4 Key observations

- The RHAII operator and CCM are **separate operators** with separate RBAC. Scoping down one
  does not affect the other.
- Many kubebuilder markers under `// Kserve` are mis-attributed — they belong to other
  components or are historical artifacts.
- Several permissions exist only because they are in the shared operator framework markers,
  not because KServe needs them.
- The `operator.openshift.io` API group is present on xKS (CCM installs the LWS operator
  CRD under this group), so rules for that group are not automatically irrelevant on non-OpenShift.
- `SelfSubjectRulesReview` does not require an explicit RBAC grant — Kubernetes allows it
  implicitly for any authenticated identity.
- The HardwareProfile webhook handler is registered in-process but has no webhook configuration
  routing to it on RHAII, making it inert.

---

## Artifacts

| File | Description |
|---|---|
| `rhaii-rbac-analysis.md` | Initial static analysis and proposed ClusterRole |
| `rhaii-rbac-gap-analysis.md` | Gap analysis cross-referencing proposal against codebase |
| `xks-rhaii-minimal-clusterrole.yaml` | Runtime-validated minimal ClusterRole for RHAII on xKS |
| `rhaii-rbac-testing-guide.md` | Step-by-step guide for reproducing the validation on KinD |

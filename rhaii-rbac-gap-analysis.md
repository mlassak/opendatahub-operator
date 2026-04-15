# RHAII RBAC Gap Analysis

**Ticket:** RHOAIENG-54569
**Date:** 2026-04-15
**Source:** Cross-referencing `rhaii-rbac-analysis.md` against the actual codebase

---

## Overview

After cross-referencing the RBAC analysis document against the actual codebase, the following gaps and issues were identified in the proposed minimal ClusterRole.

---

## ~~GAP 1: Missing `authorization.k8s.io/selfsubjectrulesreviews`~~ (FALSE POSITIVE)

~~The GC action creates a `SelfSubjectRulesReview` at runtime.~~

**Disproved by runtime validation on KinD.** `SelfSubjectRulesReview` is a
self-referential Kubernetes API -- any authenticated user can create a review of
their *own* permissions without an explicit RBAC grant. This is the same behavior
as `SelfSubjectAccessReview`. Kubernetes allows this implicitly, so no ClusterRole
rule is needed.

---

## GAP 2: Missing `infrastructure.opendatahub.io/hardwareprofiles` (NOT APPLICABLE on XKS)

The HardwareProfile webhook handler is registered in-process when KServe is enabled
(`internal/webhook/webhook.go:42-44`), but on RHAII/XKS there is **no
`MutatingWebhookConfiguration` routing to it** -- the RHAII webhook manifests
(`config/rhaii/webhook/manifests.yaml`) only include the two KServe connection
webhooks (`connection-isvc`, `connection-llmisvc`). The webhook path exists on the
server but Kubernetes never invokes it.

The `upgrade.CleanupExistingResource()` code path that creates HardwareProfile
resources (`pkg/upgrade/upgrade.go:97-113`) is also a non-issue on XKS: it checks
for the HardwareProfile CRD existence first (`cluster.HasCRD`), and the CRD is not
deployed as part of RHAII (not in `config/rhaii/crd/`). The migration is skipped
entirely.

**Validated on KinD:** upgrade task ran without errors; no HardwareProfile-related
API calls were made.

**Note:** This gap would be relevant on **OpenShift/RHOAI** where the full operator
deploys the HardwareProfile CRD and webhook configuration. If RHAII is ever extended
to include HardwareProfiles, this rule must be added:

```yaml
- apiGroups: [infrastructure.opendatahub.io]
  resources: [hardwareprofiles, hardwareprofiles/status, hardwareprofiles/finalizers]
  verbs: [get, list, watch, create, update, patch, delete]
```

---

## GAP 3: Missing Cloud Manager RBAC (`cloudmanager/common/kubebuilder_rbac.go`) (HIGH)

The analysis mentions this file in the Appendix but **does not include its rules in the proposed ClusterRole**. Critical rules missing:

| Resource | Why needed |
|---|---|
| `cert-manager.io/clusterissuers` | KServe dependency monitoring checks for ClusterIssuer CRD (`kserve_controller_actions.go:276`) |
| `sailoperator.io/istios` | Cloud manager manages sail-operator for KServe's Istio dependency |
| `operator.openshift.io/certmanagers` | Cloud manager manages cert-manager operator |
| `rbac.authorization.k8s.io/roles` with `resourceNames` + `bind;escalate` | Cloud manager needs escalation for cert-manager and LWS operator roles |

The cloud manager RBAC has **scoped escalation permissions** (lines 25-32) that are particularly important and entirely absent from the proposed ClusterRole.

---

## GAP 4: Missing `monitoring.coreos.com/servicemonitors` (MEDIUM)

KServe **owns** ServiceMonitor resources (`kserve_controller.go:76`):
```go
OwnsGVK(gvk.CoreosServiceMonitor, reconciler.Dynamic(reconciler.CrdExists(gvk.CoreosServiceMonitor)))
```

This means the KServe controller watches, creates, and deletes ServiceMonitors. The proposed ClusterRole omits this entirely. While it's dynamically gated on CRD existence, when the CRD exists the operator needs:

```yaml
- apiGroups: [monitoring.coreos.com]
  resources: [servicemonitors]
  verbs: [get, create, delete, update, watch, list, patch]
```

---

## GAP 5: Missing `template.openshift.io/templates` for OpenShift (MEDIUM)

The analysis in Section 3 (the proposed ClusterRole) **intentionally excludes** `template.openshift.io/templates` for XKS. However, the KServe controller explicitly **owns** OpenShift templates (`kserve_controller.go:75`):
```go
OwnsGVK(gvk.OpenshiftTemplate, reconciler.WithPredicates(hash.Updated()), reconciler.Dynamic(reconciler.ClusterIsOpenShift()))
```

The analysis document's note in Section 3 acknowledges this but only in passing. If RHAII is **also deployed on RHOAI/OpenShift** (the RHOAI path in `config/rhaii/rhoai/` confirms this is a real deployment target), these rules are mandatory. The proposed ClusterRole should either include them or there should be **two variants** -- the document doesn't make this clear.

---

## GAP 6: `CleanupExistingResource` runs unconditionally and needs broad RBAC (HIGH)

`cmd/main.go:498-506` shows that `upgrade.CleanupExistingResource()` runs at startup **regardless of component flags**. This code:

1. Gets application namespace via DSCI (needs `dscinitialization.opendatahub.io` GET/LIST even though DSCI controller is disabled)
2. Lists and deletes RoleBindings (`rbac.authorization.k8s.io/rolebindings`)
3. Checks for HardwareProfile CRD existence and runs migration creating HardwareProfile resources
4. Checks for AcceleratorProfile CRD (`dashboard.opendatahub.io/acceleratorprofiles`) and reads them
5. Lists Notebooks (`kubeflow.org/notebooks`) and patches them
6. Lists InferenceServices (`serving.kserve.io/inferenceservices`) and patches them
7. Gets/Lists ServingRuntimes (`serving.kserve.io/servingruntimes`)
8. Creates Events (`events` core)
9. Gets/Patches GatewayConfig (`services.platform.opendatahub.io/gatewayconfigs`)
10. Gets Services (core)
11. Checks `ValidatingAdmissionPolicyBinding` (`admissionregistration.k8s.io`)

The proposed ClusterRole covers some of these but misses several (e.g., `dashboard.opendatahub.io/acceleratorprofiles`, `services.platform.opendatahub.io/gatewayconfigs`). The analysis categorizes these as "disabled component only" but the cleanup code doesn't respect component flags.

---

## GAP 7: Missing `operators.coreos.com/subscriptions` for OpenShift (MEDIUM)

KServe's `checkSubscriptionDependencies()` (`kserve_controller_actions.go:287-313`) and its `Watches` setup (`kserve_controller.go:113-124`) explicitly watch `operators.coreos.com/Subscriptions` for cert-manager, LWS, and RHCL operator subscriptions on OpenShift clusters. The proposed ClusterRole omits `operators.coreos.com` entirely.

---

## GAP 8: Missing `networking.k8s.io/ingresses` (LOW)

The current kubebuilder markers include `networking.k8s.io/ingresses` (line 42 of `kubebuilder_rbac.go`), and on XKS, Ingress resources may be used instead of OpenShift Routes. The proposed ClusterRole omits this.

---

## GAP 9: Missing `config.openshift.io/clusterversions` and `config.openshift.io/infrastructures` (MEDIUM)

`cluster.Init()` at startup unconditionally queries these for platform detection. Even though they are OpenShift-specific APIs, on OpenShift the operator will fail to start without these permissions. The analysis categorizes these as "disabled only" but they run before any component suppression takes effect.

---

## GAP 10: Structural Issue -- Analysis assumes XKS-only, but RHAII has RHOAI deployment path

The analysis says the proposed ClusterRole is for XKS and lists OpenShift extras as a footnote. However, `config/rhaii/rhoai/` is a **full RHOAI deployment path** with its own kustomization, CRDs, patches, and namespace. The deliverable should be **two ClusterRoles** (or a kustomize overlay for OpenShift extras), but the recommendation in Section 5 only mentions creating one file at `config/rhaii/rbac/role.yaml`.

---

## Summary of Gaps by Severity

| Severity | Gap | Impact |
|---|---|---|
| ~~CRITICAL~~ FALSE POSITIVE | ~~`selfsubjectrulesreviews` missing~~ | Kubernetes implicitly allows self-subject reviews; no RBAC rule needed |
| ~~CRITICAL~~ N/A on XKS | ~~`hardwareprofiles` wrongly excluded~~ | No webhook config routes to it on RHAII; CRD not deployed; upgrade code skips gracefully. Relevant only on OpenShift/RHOAI. |
| HIGH | Cloud manager RBAC missing | Dependency management for cert-manager, sail-operator, LWS fails |
| HIGH | `CleanupExistingResource` needs broader RBAC | Operator startup cleanup fails |
| MEDIUM | `servicemonitors` missing | KServe owned ServiceMonitors can't be managed |
| MEDIUM | OpenShift template rules for RHOAI path | OpenShift deployments fail to manage templates |
| MEDIUM | `subscriptions` for OpenShift | Dependency checking fails on OpenShift |
| MEDIUM | Platform detection RBAC (`clusterversions`, `infrastructures`) | Startup failure on OpenShift |
| LOW | `ingresses` missing | Ingress management may fail on XKS |

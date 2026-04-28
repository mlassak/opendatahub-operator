# RHAII RBAC Static Analysis: Generated Role vs Actual Code Usage

**Ticket:** RHOAIENG-54569

Static analysis of every `(apiGroup, resource, verb)` in the generated RHAII
ClusterRole (`config/rhaii/rbac/role.yaml`) cross-referenced against the KServe
controller code, shared framework actions, and operator startup code.

For each rule the analysis traces the code path that requires it, or marks it
as a candidate for removal if no code path was found.

---

## Methodology

1. Read every `.go` file under `internal/controller/components/kserve/`.
2. Read the shared action code called by KServe: `deploy`, `gc`, `render`,
   `status/deployments`, `status/releases`, `dependency`.
3. Read framework code: `pkg/controller/reconciler/`, `pkg/deploy/`,
   `pkg/resources/`, `pkg/cluster/`, `cmd/main.go`.
4. For each rule in the generated role, identify which code path uses it.

---

## Analysis: Shared Operator Framework Rules (`internal/controller/rbac/`)

These rules come from the shared framework package and apply to all deployment
modes.

### Core resources (`""` apiGroup)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `configmaps` | CRUD+watch | Deploy action (SSA), KServe `customizeKserveConfigMap` reads CM from rendered resources, GC deletes | **NEEDED** |
| `configmaps/status` | get,update,patch,delete | Deploy action SSA on ConfigMaps can touch status subresource | **NEEDED** |
| `secrets`, `secrets/finalizers` | CRUD+watch | Deploy action creates/patches Secrets, KServe owns Secrets, GC deletes | **NEEDED** |
| `services`, `services/finalizers` | CRUD+watch | Deploy action creates/patches Services, KServe owns Services, GC deletes | **NEEDED** |
| `serviceaccounts` | CRUD+watch | Deploy action creates ServiceAccounts, KServe owns them | **NEEDED** |
| `namespaces`, `namespaces/finalizers` | CRUD+watch | Framework creates application namespace (`pkg/cluster/resources.go`), GC may list | **NEEDED** |
| `events` | CRUD+watch | controller-runtime event recorder | **NEEDED** |
| `pods` | CRUD+watch | Framework Owns pods indirectly via Deployments; GC may list/delete | **NEEDED** |
| `pods/log` | get | Debug/troubleshooting (`pkg/cluster/`) | **REVIEW** — may not be exercised in RHAII |
| `pods/exec` | create | Operational tasks (`pkg/cluster/`) | **REVIEW** — may not be exercised in RHAII |
| `endpoints` | get,list,watch,create,update,delete | Deploy action SSA on Services touches Endpoints | **NEEDED** |
| `nodes` | get,list,watch | `pkg/cluster/cluster_config.go:IsSingleNodeCluster` lists nodes | **NEEDED** |
| `deployments` (core group) | CRUD+watch | Legacy; `core` group deployments marker — duplicate of `apps` group | **REDUNDANT** but harmless |

### Apps group

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `deployments`, `deployments/finalizers` | CRUD+watch | KServe Owns Deployments, deploy action has special Deployment merging, status action lists them | **NEEDED** |
| `replicasets` | get,list,watch | Deploy action `RevertManagedDeploymentDrift` reads ReplicaSets to detect drift | **NEEDED** |

### Admission registration

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `mutatingwebhookconfigurations` | CRUD+watch | KServe owns MutatingWebhookConfiguration (connection-isvc, connection-llmisvc webhooks) | **NEEDED** |
| `validatingwebhookconfigurations` | CRUD+watch | KServe owns ValidatingWebhookConfiguration | **NEEDED** |
| `validatingadmissionpolicies` | CRUD+watch | KServe owns ValidatingAdmissionPolicy | **NEEDED** |
| `validatingadmissionpolicybindings` | CRUD+watch | KServe owns ValidatingAdmissionPolicyBinding | **NEEDED** |

### API extensions

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `customresourcedefinitions` | CRUD+watch | Reconciler framework checks CRD existence (`cluster.HasCRD`), deploy action has special CRD handling, KServe watches CRDs | **NEEDED** |

### RBAC

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `roles`, `rolebindings`, `clusterroles`, `clusterrolebindings` | `*` | KServe Owns all four types; deploy action creates them from manifests; GC deletes them | **NEEDED** |

### Coordination

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `leases` | CRUD+watch | controller-runtime leader election | **NEEDED** |

### Cert-manager

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `issuers` | create,patch | KServe dependency monitoring checks cert-manager CRDs; webhook TLS certs | **NEEDED** |
| `certificates` | CRUD+watch | Webhook certificate management on XKS | **NEEDED** |

### Networking

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `networkpolicies` | CRUD+watch | KServe owns NetworkPolicies | **NEEDED** |
| `ingresses` | CRUD+watch | Shared marker; ingresses may be created by KServe manifests on XKS (non-OpenShift) | **NEEDED** |

### Monitoring

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `servicemonitors` | CRUD+watch+deletecollection | KServe dynamically owns ServiceMonitors (`OwnsGVK(gvk.CoreosServiceMonitor)`) | **NEEDED** |

### Extensions (deprecated)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `deployments` (extensions group) | CRUD+watch | Legacy API group alias for `apps/deployments` | **REDUNDANT** but harmless |
| `replicasets` (extensions group) | get,list,watch | Legacy API group alias for `apps/replicasets` | **REDUNDANT** but harmless |

### Wildcard apiGroup (`*`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `deployments` | CRUD+watch | Catches all API groups for deployments (framework generic handling) | **REDUNDANT** with `apps` — but used by framework |
| `services` | CRUD+watch | Catches all API groups for services | **REDUNDANT** with `core` — but used by framework |
| `replicasets` | get,list,watch | Catches all API groups for replicasets | **REDUNDANT** with `apps` — but used by framework |

---

## Analysis: KServe Component Rules (`internal/controller/components/kserve/`)

### KServe CR management (`components.platform.opendatahub.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `kserves` | CRUD+watch | Reconciler primary resource — the Kserve CR itself | **NEEDED** |
| `kserves/status` | get,update,patch | Status updates by reconciler | **NEEDED** |
| `kserves/finalizers` | update | Finalizer management for GC | **NEEDED** |

### KServe serving resources (`serving.kserve.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `clusterservingruntimes[/*]` | CRUD+watch | Deploy action applies from manifests; KServe watches via CRD dependency | **NEEDED** |
| `clusterstoragecontainers[/*]` | CRUD+watch | Deploy action applies from manifests | **NEEDED** |
| `inferencegraphs[/*]` | CRUD+watch | Deploy action applies from manifests | **NEEDED** |
| `inferenceservices[/*]` | CRUD+watch | Core KServe resource; webhooks mutate these | **NEEDED** |
| `llminferenceserviceconfigs[/*]` | CRUD+watch | LLM-d: KServe owns these (`OwnsGVK`), deploy action applies them with special ordering | **NEEDED** |
| `llminferenceservices[/*]` | get,list,watch | LLM-d: connection webhooks read these | **NEEDED** |
| `predictors[/*]` | CRUD+watch | KServe predictors | **NEEDED** |
| `servingruntimes[/*]` | CRUD+watch | KServe serving runtimes | **NEEDED** |
| `trainedmodels[/*]` | CRUD+watch | KServe trained models | **NEEDED** |

### Inference networking (`inference.networking.x-k8s.io`, `inference.networking.k8s.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `inferencepools` (x-k8s.io) | get,list,watch | KServe owns InferencePoolV1alpha2 (`OwnsGVK`) | **NEEDED** |
| `inferencemodels` (x-k8s.io) | get,list,watch | KServe owns InferenceModelV1alpha2 (`OwnsGVK`) | **NEEDED** |
| `inferencepools` (k8s.io) | get,list,watch | KServe owns InferencePoolV1 (`OwnsGVK`) | **NEEDED** |

### OpenShift templates (`template.openshift.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `templates` | CRUD+watch | KServe owns OpenshiftTemplate (`OwnsGVK`, dynamic, ClusterIsOpenShift) | **REVIEW** — only used on OpenShift, not on XKS where RHAII runs |

### OpenShift config (`config.openshift.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `ingresses` | get | `pkg/cluster/cluster_config.go:GetDomain` reads cluster ingress domain | **REVIEW** — only on OpenShift; on XKS falls back gracefully (IsNoMatchError) |

### KEDA (`keda.sh`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `kedacontrollers` | get,list,watch | KServe manifest kustomization may reference KEDA | **NEEDED** — manifests contain KEDA resources |
| `triggerauthentications` | CRUD+watch | KServe manifests include TriggerAuthentication for autoscaling | **NEEDED** |

### Metrics (`metrics.k8s.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `pods`, `nodes` | get,list,watch | KServe autoscaling reads metrics | **NEEDED** |

### Kuadrant (`kuadrant.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `kuadrants` | get,list,watch | KServe dependency check: `kserve_controller_actions.go` checks for Kuadrant | **NEEDED** |

### Kubeflow (`kubeflow.org`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `notebooks` | CRUD+watch | **No KServe code references this.** Listed in the `// Kserve` section of the original `datasciencecluster/kubebuilder_rbac.go` but never used by any KServe controller code. Likely misplaced — belongs to Workbenches. | **OVER-GRANT** |

### LeaderWorkerSet operator (`operator.openshift.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `leaderworkersetoperators` | get,list,watch | KServe watches LWS operator status (`WatchesGVK(gvk.LeaderWorkerSetOperatorV1)`) and dependency monitoring (`dependency.MonitorOperator`) | **NEEDED** |

### Events (`events.k8s.io`)

| Resource | Verbs | Used By | Verdict |
|----------|-------|---------|---------|
| `events` | create,delete,get,list,patch,watch | controller-runtime event recorder; shared framework | **NEEDED** |

---

## Summary

### Over-grants (candidates for removal)

| apiGroup | Resource | Reason |
|----------|----------|--------|
| `kubeflow.org` | `notebooks` | No KServe code references it. Misplaced marker — belongs to Workbenches component. |
| `template.openshift.io` | `templates` | Only used on OpenShift clusters. RHAII targets XKS. On XKS the `OwnsGVK` is gated by `reconciler.Dynamic(reconciler.ClusterIsOpenShift())` — the watch is never registered, so the permission is never used. However, leaving it is harmless and preserves the ability to run RHAII on OpenShift in the future. |
| `config.openshift.io` | `ingresses` | Only used on OpenShift. On XKS, `GetDomain` falls back on `IsNoMatchError`. Same consideration as templates. |
| `""` (core) | `pods/log` | Framework marker for debugging. Not exercised by KServe reconciliation. |
| `""` (core) | `pods/exec` | Framework marker for operational tasks. Not exercised by KServe reconciliation. |
| `""` (core), `*`, `extensions` | `deployments` (redundant groups) | The `apps` group rules are sufficient. Wildcard/core/extensions are legacy duplicates. |
| `*`, `extensions` | `replicasets` | The `apps` group rule is sufficient. |
| `*` | `services` | The `core` group rule is sufficient. |

### Definite over-grant

Only **`kubeflow.org/notebooks`** is a definite over-grant — no KServe code
path requires it. All other items marked REVIEW are either harmless redundancy
or OpenShift-fallback rules that the dynamic reconciler gracefully handles.

### Under-grants

None identified. The generated role covers all API calls traced in the KServe
controller, shared actions, and framework code.

### Wildcard and redundant rules

The framework markers include wildcard `apiGroup: "*"` rules for `deployments`,
`services`, and `replicasets`. These exist because the framework's deploy action
and GC work with Unstructured resources and can theoretically touch any API group.
They are technically redundant with the specific `apps`/`core` group rules but
removing them could break if a manifest ever introduces a resource from an
unexpected API group. **Recommendation: keep them.**

---

## Comparison with runtime audit

This static analysis should be cross-referenced with the runtime audit
(`hack/rhaii-rbac-audit.sh`) for validation. The runtime audit will:
- Confirm that `kubeflow.org/notebooks` is never called (over-grant)
- Confirm that OpenShift-specific resources (`templates`, `config.openshift.io/ingresses`) are never called on XKS
- Potentially reveal API calls not visible in static analysis (e.g., from kustomize-rendered manifests that contain resources the static analysis cannot trace)

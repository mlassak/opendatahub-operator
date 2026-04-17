# RHAII RBAC: Active Controllers & Rule-to-Component Mapping

**Ticket:** RHOAIENG-54569
**Date:** 2026-04-17

---

## 1. Active Controllers and Reconcilers in RHAII Mode

RHAII mode runs the operator on non-OpenShift Kubernetes (platform type `XKS`) with a drastically reduced scope. The suppression mechanism uses `RHAI_DISABLE_*` environment variables, defined in the manager patch (`config/rhaii/odh-operator/manager_patch.yaml`).

### 1.1 Active

| Controller/Reconciler | Source | Notes |
|---|---|---|
| **KServe component controller** | `internal/controller/components/kserve/kserve_controller.go` | The only component reconciler that runs. Reconciles `Kserve` CRs. |
| **Connection webhooks** (in-process) | `internal/webhook/serving/mutating_isvc.go`, `mutating_llmisvc.go` | Two MutatingWebhookConfigurations: `connection-isvc.opendatahub.io` (InferenceServices v1beta1) and `connection-llmisvc.opendatahub.io` (LLMInferenceServices v1alpha1/v1alpha2). These run in the operator process and intercept CREATE/UPDATE on serving resources. Defined in `config/rhaii/webhook/manifests.yaml`. |

### 1.2 Suppressed

All of the following are explicitly disabled via `RHAI_DISABLE_*` env vars:

**High-level controllers (completely suppressed):**

| Controller | Disable Flag |
|---|---|
| DSCInitialization | `RHAI_DISABLE_DSCI_RESOURCE=true` |
| DataScienceCluster | `RHAI_DISABLE_DSC_RESOURCE=true` |

**Component controllers (all disabled except KServe):**

| Component | Disable Flag |
|---|---|
| Dashboard | `RHAI_DISABLE_DASHBOARD_COMPONENT` |
| DataSciencePipelines | `RHAI_DISABLE_DATASCIENCEPIPELINES_COMPONENT` |
| FeastOperator | `RHAI_DISABLE_FEASTOPERATOR_COMPONENT` |
| Kueue | `RHAI_DISABLE_KUEUE_COMPONENT` |
| LlamaStackOperator | `RHAI_DISABLE_LLAMASTACKOPERATOR_COMPONENT` |
| MLflowOperator | `RHAI_DISABLE_MLFLOWOPERATOR_COMPONENT` |
| ModelController | `RHAI_DISABLE_MODELCONTROLLER_COMPONENT` |
| ModelRegistry | `RHAI_DISABLE_MODELREGISTRY_COMPONENT` |
| ModelsAsService | `RHAI_DISABLE_MODELSASSERVICE_COMPONENT` |
| Ray | `RHAI_DISABLE_RAY_COMPONENT` |
| SparkOperator | `RHAI_DISABLE_SPARKOPERATOR_COMPONENT` |
| Trainer | `RHAI_DISABLE_TRAINER_COMPONENT` |
| TrainingOperator | `RHAI_DISABLE_TRAININGOPERATOR_COMPONENT` |
| TrustyAI | `RHAI_DISABLE_TRUSTYAI_COMPONENT` |
| Workbenches | `RHAI_DISABLE_WORKBENCHES_COMPONENT` |

**Service controllers (all disabled):**

| Service | Disable Flag |
|---|---|
| Auth | `RHAI_DISABLE_AUTH_SERVICE` |
| CertConfigMapGenerator | `RHAI_DISABLE_CERTCONFIGMAPGENERATOR_SERVICE` |
| Gateway | `RHAI_DISABLE_GATEWAY_SERVICE` |
| Monitoring | `RHAI_DISABLE_MONITORING_SERVICE` |
| SetupController | `RHAI_DISABLE_SETUPCONTROLLER_SERVICE` |

### 1.3 Registration and Suppression Mechanism

The flow in `cmd/main.go`:

1. All 16 components and 5 services are registered in `existingComponents` / `existingServices` maps (lines 118-143).
2. `registerComponents()` and `registerServices()` (lines 192-208) iterate these maps. For each, `flags.IsComponentEnabled(name)` / `flags.IsServiceEnabled(name)` is checked. If the `RHAI_DISABLE_<NAME>_COMPONENT` env var is set to `true`, the handler is disabled via `cr.Disable(name)` / `sr.Disable(name)`.
3. `flags.IsDSCEnabled()` and `flags.IsDSCIEnabled()` check `RHAI_DISABLE_DSC_RESOURCE` and `RHAI_DISABLE_DSCI_RESOURCE` respectively (lines 432-453). When disabled, `SetupWithManager` is never called for those controllers.
4. `DISABLE_DSC_CONFIG=true` additionally prevents auto-creation of default DSCI/DSC CRs.

### 1.4 What the KServe Controller Watches

From `kserve_controller.go` (lines 55-169), the KServe controller:

**Owns (triggers reconciliation on changes to owned resources):**

| Resource Type | API Group | Notes |
|---|---|---|
| Secret | core | |
| Service | core | |
| ConfigMap | core | |
| ServiceAccount | core | |
| Role | rbac.authorization.k8s.io | |
| RoleBinding | rbac.authorization.k8s.io | |
| ClusterRole | rbac.authorization.k8s.io | |
| ClusterRoleBinding | rbac.authorization.k8s.io | |
| NetworkPolicy | networking.k8s.io | |
| MutatingWebhookConfiguration | admissionregistration.k8s.io | |
| ValidatingWebhookConfiguration | admissionregistration.k8s.io | |
| ValidatingAdmissionPolicy | admissionregistration.k8s.io | |
| ValidatingAdmissionPolicyBinding | admissionregistration.k8s.io | |
| Deployment | apps | With deployment predicate |
| Template | template.openshift.io | Dynamic: only on OpenShift |
| ServiceMonitor | monitoring.coreos.com | Dynamic: only if CRD exists |
| InferencePool | inference.networking.x-k8s.io (v1alpha2) | Dynamic: only if CRD exists |
| InferencePool | inference.networking.k8s.io (v1) | Dynamic: only if CRD exists |
| InferenceModel | inference.networking.x-k8s.io (v1alpha2) | Dynamic: only if CRD exists |
| LLMInferenceServiceConfig | serving.kserve.io (v1alpha1, v1alpha2) | Dynamic: only if CRD exists |
| LLMInferenceService | serving.kserve.io (v1alpha1, v1alpha2) | Dynamic: only if CRD exists |

**Watches (external triggers that enqueue reconciliation):**

| Resource | API Group | Trigger Condition |
|---|---|---|
| CustomResourceDefinition | apiextensions.k8s.io | Name suffixed with `.networking.istio.io`, `.security.istio.io`, `.telemetry.istio.io`, `.extensions.istio.io`, `.cert-manager.io`, `.leaderworkerset.x-k8s.io`; or named `leaderworkersetoperators.operator.openshift.io`, `subscriptions.operators.coreos.com`; or labeled `app.opendatahub.io/kserve=true` |
| Subscription | operators.coreos.com | Named `rhcl-operator`, `leader-worker-set`, or `cert-manager-operator`. Dynamic: only if Subscription CRD exists. |
| LeaderWorkerSetOperator | operator.openshift.io (v1) | Dynamic: only if CRD exists. Watches status changes. |

---

## 2. RBAC Rule to Controller/Component Mapping

This section maps every RBAC rule in the full `config/rbac/role.yaml` (generated from kubebuilder markers in `datasciencecluster/kubebuilder_rbac.go` and `dscinitialization/kubebuilder_rbac.go`) to the controller(s) that require it.

Rules are categorized as:
- **RHAII-Required**: Needed for the KServe controller, operator framework, or webhooks
- **Not Required in RHAII**: Only needed by disabled controllers/services

### 2.1 Operator Framework / Shared Infrastructure (RHAII-Required)

These rules are needed regardless of which component is active because they support the controller-runtime framework, leader election, webhook serving, and manifest deployment.

| API Group | Resources | Verbs | Required By |
|---|---|---|---|
| `""` (core) | configmaps, configmaps/status | get, create, watch, patch, delete, list, update | Operator framework (leader election, config storage), KServe controller (owns ConfigMaps) |
| `""` (core) | secrets, secrets/finalizers | create, delete, get, list, update, watch, patch | KServe controller (owns Secrets), webhook TLS |
| `""` (core) | services, services/finalizers | create, delete, get, list, update, watch, patch | KServe controller (owns Services), webhook service |
| `""` (core) | serviceaccounts | get, list, watch, create, update, patch, delete | KServe controller (owns ServiceAccounts) |
| `""` (core) | namespaces, namespaces/finalizers | get, create, patch, delete, watch, update, list | KServe controller (manages application namespace) |
| `""` (core) | events | get, create, watch, update, list, patch, delete | Event recording (controller-runtime Recorder) |
| `""` (core) | pods | get, list, watch, create, update, patch, delete | KServe deploys workloads; may need pod management |
| events.k8s.io | events | list, watch, patch, delete, get | Event recording |
| apps | deployments, deployments/finalizers | get, list, watch, create, update, patch, delete | KServe controller (owns Deployments) |
| apps | replicasets | get, list, watch | KServe controller (watches ReplicaSets of owned Deployments) |
| admissionregistration.k8s.io | mutatingwebhookconfigurations | create, delete, get, list, patch, update, watch | KServe controller (owns MutatingWebhookConfiguration) |
| admissionregistration.k8s.io | validatingwebhookconfigurations | get, list, watch, create, update, delete, patch | KServe controller (owns ValidatingWebhookConfiguration) |
| admissionregistration.k8s.io | validatingadmissionpolicies | get, create, delete, update, watch, list, patch | KServe controller (owns VAP) |
| admissionregistration.k8s.io | validatingadmissionpolicybindings | get, create, delete, update, watch, list, patch | KServe controller (owns VAPB) |
| apiextensions.k8s.io | customresourcedefinitions | get, list, watch, create, patch, delete, update | KServe controller (watches CRDs for dependency detection) |
| rbac.authorization.k8s.io | roles, rolebindings, clusterroles, clusterrolebindings | * | KServe controller (owns Roles, RoleBindings, ClusterRoles, ClusterRoleBindings for KServe operands) |
| coordination.k8s.io | leases | get, list, watch, create, update, patch, delete | Leader election |
| cert-manager.io | issuers | create, patch | KServe dependency: cert-manager integration |
| cert-manager.io | certificates | get, list, watch, create, update, patch, delete | KServe dependency: cert-manager integration |
| networking.k8s.io | networkpolicies | get, create, list, watch, delete, update, patch | KServe controller (owns NetworkPolicies) |

### 2.2 KServe Component (RHAII-Required)

| API Group | Resources | Verbs | Required By |
|---|---|---|---|
| components.platform.opendatahub.io | kserves | get, list, watch, create, update, patch, delete | KServe controller (primary CR) |
| components.platform.opendatahub.io | kserves/status | get, update, patch | KServe controller (status updates) |
| components.platform.opendatahub.io | kserves/finalizers | update | KServe controller (finalizer management) |
| serving.kserve.io | clusterservingruntimes, /status, /finalizers | CRUD | KServe operand resources |
| serving.kserve.io | clusterstoragecontainers, /status, /finalizers | CRUD | KServe operand resources |
| serving.kserve.io | inferencegraphs, /status | CRUD | KServe operand resources |
| serving.kserve.io | inferenceservices, /status, /finalizers | CRUD | KServe operand resources, webhook |
| serving.kserve.io | predictors, /status, /finalizers | CRUD | KServe operand resources |
| serving.kserve.io | servingruntimes, /status, /finalizers | CRUD | KServe operand resources |
| serving.kserve.io | trainedmodels, /status | CRUD | KServe operand resources |
| serving.kserve.io | llminferenceserviceconfigs, /status | CRUD | KServe LLM-d support |
| serving.kserve.io | llminferenceservices, /status | get, list, watch | KServe LLM-d support, webhook |
| inference.networking.x-k8s.io | inferencepools, inferencemodels | get, list, watch | KServe LLM-d: Gateway API inference extension |
| inference.networking.k8s.io | inferencepools | get, list, watch | KServe LLM-d: Gateway API inference extension (v1) |
| keda.sh | kedacontrollers | get, list, watch | KServe: KEDA autoscaling support |
| keda.sh | triggerauthentications | create, delete, get, list, patch, update, watch | KServe: KEDA trigger auth for autoscaling |
| metrics.k8s.io | pods, nodes | get, list, watch | KServe: metrics-based autoscaling |
| operator.openshift.io | leaderworkersetoperators | get, list, watch | KServe: LWS dependency detection |
| monitoring.coreos.com | servicemonitors | get, create, delete, update, watch, list, patch | KServe: owns ServiceMonitor for operand monitoring |

### 2.3 DSCInitialization Controller (NOT Required in RHAII)

Source: `internal/controller/dscinitialization/kubebuilder_rbac.go`

| API Group | Resources | Required By |
|---|---|---|
| dscinitialization.opendatahub.io | dscinitializations, /status, /finalizers | DSCI controller |
| features.opendatahub.io | featuretrackers, /status, /finalizers | DSCI controller |
| config.openshift.io | authentications, infrastructures | DSCI Auth service |
| services.platform.opendatahub.io | auths, /status, /finalizers | Auth service |
| services.platform.opendatahub.io | monitorings, /status, /finalizers | Monitoring service |
| route.openshift.io | routers/metrics, routers/federate | Monitoring service |
| image.openshift.io | registry/metrics | Monitoring service |
| monitoring.coreos.com | podmonitors | Monitoring/observability |
| monitoring.coreos.com | prometheusrules | Monitoring/observability |
| monitoring.coreos.com | prometheuses, /status, /finalizers | Monitoring/observability |
| monitoring.coreos.com | alertmanagers, /status, /finalizers, alertmanagerconfigs | Monitoring/observability |
| monitoring.coreos.com | thanosrulers, /status, /finalizers | Monitoring/observability |
| monitoring.coreos.com | probes | Monitoring/observability |
| tempo.grafana.com | tempostacks, tempomonolithics | Observability |
| perses.dev | perses, persesdashboards, persesdatasources, /status, /finalizers | Observability |
| monitoring.rhobs | servicemonitors, monitoringstacks, prometheusrules, thanosqueriers, /status, /finalizers | RHOBS monitoring |
| opentelemetry.io | opentelemetrycollectors, instrumentations, /status, /finalizers | Observability |

### 2.4 DataScienceCluster Controller -- Component-Specific (NOT Required in RHAII)

Source: `internal/controller/datasciencecluster/kubebuilder_rbac.go`

**DSC Core:**

| API Group | Resources | Required By |
|---|---|---|
| datasciencecluster.opendatahub.io | datascienceclusters, /status, /finalizers | DSC controller |
| authentication.k8s.io | tokenreviews | DSC controller |
| authorization.k8s.io | subjectaccessreviews | DSC controller |
| operators.coreos.com | clusterserviceversions, subscriptions, operatorconditions, catalogsources | DSC controller (OLM integration) |
| operator.openshift.io | consoles, ingresscontrollers | DSC controller |

**Dashboard:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | dashboards, /status, /finalizers | Dashboard component |
| opendatahub.io | odhdashboardconfigs | Dashboard component |
| console.openshift.io | odhquickstarts, consolelinks | Dashboard component |
| dashboard.opendatahub.io | odhdocuments, odhapplications, acceleratorprofiles, hardwareprofiles | Dashboard component |

**Ray:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | rays, /status, /finalizers | Ray component |
| ray.io | rayservices, rayjobs, rayclusters | Ray component |
| autoscaling | horizontalpodautoscalers | Ray component |
| autoscaling.openshift.io | machinesets, machineautoscalers | Ray component |

**ModelRegistry:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | modelregistries, /status, /finalizers | ModelRegistry component |
| modelregistry.opendatahub.io | modelregistries, /status, /finalizers | ModelRegistry component |

**Kueue:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | kueues, /status, /finalizers | Kueue component |
| kueue.x-k8s.io | clusterqueues, localqueues, resourceflavors, /status | Kueue component |
| kueue.openshift.io | kueues, /status | Kueue component |

**Workbenches:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | workbenches, /status, /finalizers | Workbenches component |
| image.openshift.io | imagestreamtags, imagestreams | Workbenches component |
| build.openshift.io | builds, buildconfigs, buildconfigs/instantiate | Workbenches component |

**DataSciencePipelines:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | datasciencepipelines, /status, /finalizers | DSP component |
| datasciencepipelinesapplications.opendatahub.io | datasciencepipelinesapplications, /status, /finalizers | DSP component |
| argoproj.io | workflows | DSP component |

**TrainingOperator:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | trainingoperators, /status, /finalizers | Training component |

**TrustyAI:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | trustyais, /status, /finalizers | TrustyAI component |

**ModelController:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | modelcontrollers, /status, /finalizers | ModelController component |

**ModelMeshServing:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | modelmeshservings, /status, /finalizers | ModelMesh component |

**FeastOperator:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | feastoperators, /status, /finalizers | Feast component |

**LlamaStackOperator:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | llamastackoperators, /status, /finalizers | LlamaStack component |

**Trainer:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | trainers, /status, /finalizers | Trainer component |
| trainer.kubeflow.org | clustertrainingruntimes | Trainer component |

**MLflowOperator:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | mlflowoperators, /status, /finalizers | MLflow component |
| mlflow.opendatahub.io | mlflows, /status, /finalizers | MLflow component |

**SparkOperator:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | sparkoperators, /status, /finalizers | Spark component |

**Models-as-a-Service:**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | modelsasservices, /status, /finalizers | MaaS component |
| kuadrant.io | authpolicies, tokenratelimitpolicies, ratelimitpolicies, telemetrypolicies | MaaS component |
| extensions.kuadrant.io | telemetrypolicies | MaaS component |
| operator.authorino.kuadrant.io | authorinos | MaaS component |
| telemetry.istio.io | telemetries | MaaS component |

**CodeFlare (read-only legacy):**

| API Group | Resources | Required By |
|---|---|---|
| components.platform.opendatahub.io | codeflares, /status | DSC controller (legacy read-only) |

**HardwareProfile:**

| API Group | Resources | Required By |
|---|---|---|
| infrastructure.opendatahub.io | hardwareprofiles, /status, /finalizers | HardwareProfile webhook/controller |

**Other DSC-only resources:**

| API Group | Resources | Required By |
|---|---|---|
| snapshot.storage.k8s.io | volumesnapshots | DSC controller (storage) |
| security.openshift.io | securitycontextconstraints | DSC controller (OpenShift SCC) |
| route.openshift.io | routes | DSC controller (OpenShift routes) |
| oauth.openshift.io | oauthclients | DSC controller (OpenShift OAuth) |
| apiregistration.k8s.io | apiservices | DSC controller |
| user.openshift.io | users, groups | DSC controller (RHOAI user management) |
| machine.openshift.io | machinesets, machineautoscalers | DSC controller (OpenShift machines) |
| integreatly.org | rhmis | DSC controller (RHMI integration) |
| config.openshift.io | ingresses, clusterversions | DSC controller (OpenShift config) |
| controller-runtime.sigs.k8s.io | controllermanagerconfigs | DSC controller |
| template.openshift.io | templates | DSC controller (OpenShift templates) |
| kubeflow.org | notebooks | DSC controller (Workbenches) |
| policy | poddisruptionbudgets | DSC controller |
| batch | jobs, jobs/status, cronjobs | DSC controller (batch workloads) |
| networking.k8s.io | ingresses | DSC controller (ingress management) |
| `""` (core) | persistentvolumes, persistentvolumeclaims | DSC controller (storage) |
| `""` (core) | endpoints | DSC controller |
| `""` (core) | pods/log, pods/exec | DSC controller (operational debugging) |
| `""` (core) | nodes | DSC controller (node info) |
| apps | statefulsets | DSC controller (stateful workloads) |
| `*` (wildcard) | deployments, replicasets, statefulsets, services | DSC controller (broad access via `extensions` group) |
| machinelearning.seldon.io | seldondeployments | DSC controller (legacy Seldon) |
| kuadrant.io | kuadrants | DSC controller (Kuadrant detection) |
| operator.openshift.io | jobsetoperators | DSC controller |

### 2.5 Existing RHAII ClusterRole vs. Full Role

The existing RHAII ClusterRole at `config/rhaii/rbac/role.yaml` (141 lines) already reflects a scoped-down version matching the analysis above. Comparing against the full `config/rbac/role.yaml` (1369 lines):

**Correctly included** (matches Sections 2.1 + 2.2):
- All operator framework resources (configmaps, secrets, services, serviceaccounts, namespaces, events, pods)
- All apps resources (deployments, replicasets)
- Admission registration (webhooks, VAP, VAPB)
- CRDs, RBAC, leases, cert-manager
- Network policies
- KServe component CR (kserves, /status, /finalizers)
- All serving.kserve.io resources
- Inference pool/model resources
- KEDA resources
- Metrics resources
- LeaderWorkerSetOperators
- ServiceMonitors

**Correctly excluded**:
- All disabled component CRs (dashboards, rays, modelregistries, etc.)
- All disabled service CRs (auths, monitorings, etc.)
- OpenShift-specific resources (routes, SCCs, OAuth, imagestreams, builds, templates, consolelinks)
- Full observability stack (Prometheus, Alertmanager, Thanos, Perses, OpenTelemetry, RHOBS)
- DSC/DSCI resources
- Storage (PVs, PVCs, VolumeSnapshots)
- Batch (jobs, cronjobs)
- StatefulSets
- Legacy resources (Seldon, RHMI, CodeFlare)
- All other component-specific APIs

This represents an approximately **90% reduction** in RBAC scope (1369 lines -> 141 lines).

---

## Appendix: File References

| File | Purpose |
|---|---|
| `config/rbac/role.yaml` | Full ClusterRole (1369 lines, all components) |
| `config/rhaii/rbac/role.yaml` | RHAII-scoped ClusterRole (141 lines, KServe-only) |
| `config/rhaii/odh-operator/manager_patch.yaml` | RHAII env vars disabling components/services |
| `config/rhaii/odh-operator/kustomization.yaml` | RHAII kustomization (references `../rbac`) |
| `config/rhaii/webhook/manifests.yaml` | RHAII webhook configuration (2 mutating webhooks) |
| `internal/controller/datasciencecluster/kubebuilder_rbac.go` | All component RBAC markers (344 lines) |
| `internal/controller/dscinitialization/kubebuilder_rbac.go` | DSCI/monitoring/observability RBAC markers (79 lines) |
| `internal/controller/components/kserve/kserve_controller.go` | KServe controller watches/owns definitions |
| `pkg/utils/flags/suppression.go` | Component/service suppression mechanism |
| `cmd/main.go` | Controller registration and conditional setup |

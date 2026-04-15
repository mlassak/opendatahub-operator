# RHAII RBAC Analysis: Minimum Required Permissions

**Ticket:** RHOAIENG-54569
**Date:** 2026-04-13

## Executive Summary

The RHAII (Red Hat AI Inference) operator currently reuses the full `controller-manager-role` ClusterRole
(`config/rbac/role.yaml`) which contains **130+ RBAC rules** covering all 17 components and 5 services.
RHAII runs in a reduced mode with **only KServe enabled** and all other components/services disabled via
environment flags. This analysis identifies the minimum required RBAC permissions and proposes a
scoped-down ClusterRole.

The existing `config/rhaii/odh-operator/kustomization.yaml` already contains a TODO acknowledging this:
```yaml
# TODO: here we should use specific rhaii only RBAC rules
- ../../rbac
```

**Related PR:** [opendatahub-io/opendatahub-operator#3409](https://github.com/opendatahub-io/opendatahub-operator/pull/3409)
expands `cert-manager.io/certificates` permissions from `create;patch` to full CRUD
(`get;list;watch;create;update;patch;delete`) because the RHAII operator needs to manage
cert-manager Certificate resources for webhook TLS on XKS. This PR validates the finding
that cert-manager permissions are in the critical path for RHAII and must be included in
the minimal ClusterRole. See Section 2.1 and the proposed ClusterRole in Section 3 for
the updated cert-manager rules reflecting this change.

---

## 1. Active Controllers in RHAII Mode

Based on `config/rhaii/odh-operator/manager_patch.yaml`, RHAII disables all controllers except KServe:

| Controller | Status | Source |
|---|---|---|
| **KServe** | **ENABLED** | Only active component |
| DSC (DataScienceCluster) | DISABLED | `RHAI_DISABLE_DSC_RESOURCE=true` |
| DSCI (DSCInitialization) | DISABLED | `RHAI_DISABLE_DSCI_RESOURCE=true` |
| Dashboard | DISABLED | `RHAI_DISABLE_DASHBOARD_COMPONENT=true` |
| DataSciencePipelines | DISABLED | `RHAI_DISABLE_DATASCIENCEPIPELINES_COMPONENT=true` |
| FeastOperator | DISABLED | `RHAI_DISABLE_FEASTOPERATOR_COMPONENT=true` |
| Kueue | DISABLED | `RHAI_DISABLE_KUEUE_COMPONENT=true` |
| LlamaStackOperator | DISABLED | `RHAI_DISABLE_LLAMASTACKOPERATOR_COMPONENT=true` |
| MLflowOperator | DISABLED | `RHAI_DISABLE_MLFLOWOPERATOR_COMPONENT=true` |
| ModelController | DISABLED | `RHAI_DISABLE_MODELCONTROLLER_COMPONENT=true` |
| ModelRegistry | DISABLED | `RHAI_DISABLE_MODELREGISTRY_COMPONENT=true` |
| ModelsAsService | DISABLED | `RHAI_DISABLE_MODELSASSERVICE_COMPONENT=true` |
| Ray | DISABLED | `RHAI_DISABLE_RAY_COMPONENT=true` |
| SparkOperator | DISABLED | `RHAI_DISABLE_SPARKOPERATOR_COMPONENT=true` |
| Trainer | DISABLED | `RHAI_DISABLE_TRAINER_COMPONENT=true` |
| TrainingOperator | DISABLED | `RHAI_DISABLE_TRAININGOPERATOR_COMPONENT=true` |
| TrustyAI | DISABLED | `RHAI_DISABLE_TRUSTYAI_COMPONENT=true` |
| Workbenches | DISABLED | `RHAI_DISABLE_WORKBENCHES_COMPONENT=true` |
| Auth (service) | DISABLED | `RHAI_DISABLE_AUTH_SERVICE=true` |
| CertConfigMapGenerator (service) | DISABLED | `RHAI_DISABLE_CERTCONFIGMAPGENERATOR_SERVICE=true` |
| Gateway (service) | DISABLED | `RHAI_DISABLE_GATEWAY_SERVICE=true` |
| Monitoring (service) | DISABLED | `RHAI_DISABLE_MONITORING_SERVICE=true` |
| SetupController (service) | DISABLED | `RHAI_DISABLE_SETUPCONTROLLER_SERVICE=true` |

The suppression mechanism is implemented in `pkg/utils/flags/suppression.go` — each `RHAI_DISABLE_*`
environment variable maps to a CLI flag that prevents the corresponding reconciler and webhooks from
being registered with the controller manager.

---

## 2. RBAC Rule Mapping by Controller/Component

The table below maps each RBAC rule group from the current `controller-manager-role` to the
controller that requires it. Rules are sourced from the `+kubebuilder:rbac` markers in the
respective `kubebuilder_rbac.go` files.

### 2.1 Rules Required by KServe (NEEDED for RHAII)

These are declared in the **KServe section** of `internal/controller/datasciencecluster/kubebuilder_rbac.go`
(lines 213-255) and represent resources the KServe controller directly manages:

| API Group | Resources | Verbs | Purpose |
|---|---|---|---|
| `components.platform.opendatahub.io` | kserves | get, list, watch, create, update, patch, delete | KServe CR management |
| `components.platform.opendatahub.io` | kserves/status | get, update, patch | Status updates |
| `components.platform.opendatahub.io` | kserves/finalizers | update | Finalizer management |
| `serving.kserve.io` | inferenceservices, inferenceservices/finalizers | create, delete, get, list, update, watch, patch | Core KServe resources |
| `serving.kserve.io` | inferenceservices/status | update, patch, delete, get | |
| `serving.kserve.io` | inferencegraphs | create, delete, get, list, update, watch, patch | Inference graphs |
| `serving.kserve.io` | inferencegraphs/status | update, patch, delete, get | |
| `serving.kserve.io` | servingruntimes, servingruntimes/finalizers | create, delete, get, list, update, watch, patch | Serving runtimes |
| `serving.kserve.io` | servingruntimes/status | update, patch, get | |
| `serving.kserve.io` | clusterservingruntimes, clusterservingruntimes/finalizers | create, delete, get, list, update, watch, patch | Cluster-wide runtimes |
| `serving.kserve.io` | clusterservingruntimes/status | update, patch, delete, get | |
| `serving.kserve.io` | clusterstoragecontainers, clusterstoragecontainers/finalizers | create, delete, get, list, update, watch, patch | Storage containers |
| `serving.kserve.io` | clusterstoragecontainers/status | update, patch, delete, get | |
| `serving.kserve.io` | trainedmodels | create, delete, get, list, update, watch, patch | Trained models |
| `serving.kserve.io` | trainedmodels/status | update, patch, delete, get | |
| `serving.kserve.io` | predictors, predictors/finalizers | create, delete, get, list, update, watch, patch | Predictors |
| `serving.kserve.io` | predictors/status | update, patch, delete, get | |
| `serving.kserve.io` | llminferenceserviceconfigs | get, list, watch, create, update, patch, delete | LLM-d configs |
| `serving.kserve.io` | llminferenceserviceconfigs/status | get, update, patch | |
| `serving.kserve.io` | llminferenceservices, llminferenceservices/status | get, list, watch | LLM inference (read-only) |
| `inference.networking.x-k8s.io` | inferencepools, inferencemodels | get, list, watch | LLM-d networking |
| `inference.networking.k8s.io` | inferencepools | get, list, watch | LLM-d networking (GA) |
| `keda.sh` | kedacontrollers | get, list, watch | KEDA autoscaling check |
| `keda.sh` | triggerauthentications | get, list, watch, create, update, patch, delete | KEDA trigger auth |
| `metrics.k8s.io` | pods, nodes | get, list, watch | Metrics for autoscaling |
| `kuadrant.io` | kuadrants | get, list, watch | Kuadrant availability check |
| `operator.openshift.io` | leaderworkersetoperators | get, list, watch | LWS operator dependency |
| `operator.openshift.io` | jobsetoperators | get, list, watch | JobSet operator dependency |
| `config.openshift.io` | ingresses | get | Ingress domain discovery |
| `kubeflow.org` | notebooks | create, delete, get, list, update, watch, patch | KServe notebook integration |
| `template.openshift.io` | templates | get, list, watch, create, update, patch, delete | OpenShift templates |

### 2.2 Rules Required by the Operator Framework (NEEDED for RHAII)

These are shared infrastructure permissions needed regardless of which components are active:

| API Group | Resources | Verbs | Purpose |
|---|---|---|---|
| `""` (core) | configmaps, configmaps/status | get, create, watch, patch, delete, list, update | Manifest deployment, config |
| `""` (core) | secrets, secrets/finalizers | create, delete, get, list, update, watch, patch | Secret management |
| `""` (core) | services, services/finalizers | create, delete, get, list, update, watch, patch | Service deployment |
| `""` (core) | serviceaccounts | get, list, watch, create, update, patch, delete | SA deployment |
| `""` (core) | namespaces, namespaces/finalizers | get, create, patch, delete, watch, update, list | Namespace management |
| `""` (core) | events | get, create, watch, update, list, patch, delete | Event recording |
| `""` (core) | pods | get, list, watch, create, update, patch, delete | Pod management |
| `""` (core) | endpoints | watch, list, get, create, update, delete | Endpoint management |
| `events.k8s.io` | events | list, watch, patch, delete, get | Event recording |
| `apps` | deployments, deployments/finalizers | get, list, watch, create, update, patch, delete | Deployment management |
| `apps` | replicasets | get, list, watch | ReplicaSet status monitoring |
| `apps` | statefulsets | get, list, watch, create, update, patch, delete | StatefulSet management |
| `admissionregistration.k8s.io` | mutatingwebhookconfigurations | create, delete, get, list, patch, update, watch | Webhook management |
| `admissionregistration.k8s.io` | validatingwebhookconfigurations | create, delete, get, list, patch, update, watch | Webhook management |
| `admissionregistration.k8s.io` | validatingadmissionpolicies | get, create, delete, update, watch, list, patch | Admission policies |
| `admissionregistration.k8s.io` | validatingadmissionpolicybindings | get, create, delete, update, watch, list, patch | Admission policy bindings |
| `apiextensions.k8s.io` | customresourcedefinitions | get, list, watch, create, patch, delete, update | CRD management |
| `rbac.authorization.k8s.io` | roles, rolebindings, clusterroles, clusterrolebindings | * | RBAC management for deployed components |
| `coordination.k8s.io` | leases | get, list, watch, create, update, patch, delete | Leader election |
| `authentication.k8s.io` | tokenreviews | create, get | Auth proxy |
| `authorization.k8s.io` | subjectaccessreviews | create, get | Auth proxy |
| `cert-manager.io` | issuers | create, patch | TLS issuer management |
| `cert-manager.io` | certificates | get, list, watch, create, update, patch, delete | TLS certificate management (expanded per [PR #3409](https://github.com/opendatahub-io/opendatahub-operator/pull/3409)) |
| `networking.k8s.io` | networkpolicies | get, create, list, watch, delete, update, patch | Network policy deployment |
| `controller-runtime.sigs.k8s.io` | controllermanagerconfigs | get, create, patch, delete | Manager configuration |

### 2.3 Rules Required ONLY by Disabled Components (CAN BE REMOVED for RHAII)

#### Dashboard-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | dashboards, dashboards/status, dashboards/finalizers | Dashboard controller |
| `opendatahub.io` | odhdashboardconfigs | Dashboard controller |
| `console.openshift.io` | consolelinks, odhquickstarts | Dashboard controller |
| `dashboard.opendatahub.io` | odhdocuments, odhapplications, acceleratorprofiles, hardwareprofiles | Dashboard controller |

#### DataSciencePipelines-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | datasciencepipelines, datasciencepipelines/status, datasciencepipelines/finalizers | DSP controller |
| `datasciencepipelinesapplications.opendatahub.io` | datasciencepipelinesapplications, /finalizers, /status | DSP controller |
| `argoproj.io` | workflows | DSP controller |

#### Ray-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | rays, rays/status, rays/finalizers | Ray controller |
| `ray.io` | rayservices, rayjobs, rayclusters | Ray controller |
| `autoscaling` | horizontalpodautoscalers | Ray controller |
| `autoscaling.openshift.io` | machinesets, machineautoscalers | Ray controller |

#### Kueue-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | kueues, kueues/status, kueues/finalizers | Kueue controller |
| `kueue.x-k8s.io` | clusterqueues, localqueues, resourceflavors, /status | Kueue controller |
| `kueue.openshift.io` | kueues, kueues/status | Kueue controller |

#### ModelRegistry-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | modelregistries, modelregistries/status, modelregistries/finalizers | ModelRegistry controller |
| `modelregistry.opendatahub.io` | modelregistries, /status, /finalizers | ModelRegistry controller |

#### Workbenches-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | workbenches, workbenches/status, workbenches/finalizers | Workbenches controller |
| `image.openshift.io` | imagestreams, imagestreamtags | Workbenches controller |
| `build.openshift.io` | buildconfigs, buildconfigs/instantiate, builds | Workbenches controller |

#### TrainingOperator-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | trainingoperators, trainingoperators/status, trainingoperators/finalizers | TrainingOperator controller |

#### Trainer-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | trainers, trainers/status, trainers/finalizers | Trainer controller |
| `trainer.kubeflow.org` | clustertrainingruntimes | Trainer controller |

#### TrustyAI-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | trustyais, trustyais/status, trustyais/finalizers | TrustyAI controller |

#### ModelController-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | modelcontrollers, modelcontrollers/status, modelcontrollers/finalizers | ModelController |

#### ModelMeshServing-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | modelmeshservings, modelmeshservings/status, modelmeshservings/finalizers | ModelMeshServing |

#### FeastOperator-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | feastoperators, feastoperators/status, feastoperators/finalizers | FeastOperator |

#### LlamaStackOperator-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | llamastackoperators, llamastackoperators/status, llamastackoperators/finalizers | LlamaStackOperator |

#### MLflowOperator-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | mlflowoperators, mlflowoperators/status, mlflowoperators/finalizers | MLflowOperator |
| `mlflow.opendatahub.io` | mlflows, mlflows/status, mlflows/finalizers | MLflowOperator |

#### SparkOperator-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | sparkoperators, sparkoperators/status, sparkoperators/finalizers | SparkOperator |

#### ModelsAsService-only
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | modelsasservices, modelsasservices/status, modelsasservices/finalizers | MaaS controller |
| `kuadrant.io` | authpolicies, tokenratelimitpolicies, ratelimitpolicies, telemetrypolicies | MaaS controller |
| `extensions.kuadrant.io` | telemetrypolicies | MaaS controller |
| `operator.authorino.kuadrant.io` | authorinos | MaaS controller |

#### CodeFlare-only (legacy, read-only)
| API Group | Resources | Source |
|---|---|---|
| `components.platform.opendatahub.io` | codeflares, codeflares/status | Legacy read-only |

#### HardwareProfile (infrastructure, not component-specific but not KServe)
| API Group | Resources | Source |
|---|---|---|
| `infrastructure.opendatahub.io` | hardwareprofiles, /status, /finalizers | HardwareProfile controller |

#### DSC/DSCI-only (disabled in RHAII)
| API Group | Resources | Source |
|---|---|---|
| `datasciencecluster.opendatahub.io` | datascienceclusters, /status, /finalizers | DSC controller |
| `dscinitialization.opendatahub.io` | dscinitializations, /status, /finalizers | DSCI controller |
| `features.opendatahub.io` | featuretrackers, /status, /finalizers | DSCI controller |

#### Gateway service-only (disabled in RHAII)
| API Group | Resources | Source |
|---|---|---|
| `services.platform.opendatahub.io` | gatewayconfigs, /status, /finalizers | Gateway controller |
| `gateway.networking.k8s.io` | gateways, gatewayclasses, httproutes | Gateway controller |
| `networking.istio.io` | destinationrules, envoyfilters | Gateway controller |

#### Auth service-only (disabled in RHAII)
| API Group | Resources | Source |
|---|---|---|
| `services.platform.opendatahub.io` | auths, /status, /finalizers | Auth controller |
| `config.openshift.io` | authentications, infrastructures | Auth/DSCI controller |

#### Monitoring service-only (disabled in RHAII)
| API Group | Resources | Source |
|---|---|---|
| `services.platform.opendatahub.io` | monitorings, /status, /finalizers | Monitoring controller |
| `monitoring.coreos.com` | (all resources) | Monitoring/DSCI controller |
| `monitoring.rhobs` | (all resources) | Monitoring/DSCI controller |
| `tempo.grafana.com` | tempostacks, tempomonolithics | DSCI controller |
| `perses.dev` | perses, persesdashboards, persesdatasources, /status, /finalizers | DSCI controller |
| `opentelemetry.io` | instrumentations, opentelemetrycollectors, /status, /finalizers | DSCI controller |
| `route.openshift.io` | routers/metrics, routers/federate | Monitoring/DSCI controller |
| `image.openshift.io` | registry/metrics | DSCI controller |

#### Other disabled-only rules
| API Group | Resources | Source |
|---|---|---|
| `""` (core) | clusterversions, rhmis, nodes | DSC shared (not needed on XKS) |
| `""` (core) | persistentvolumes, persistentvolumeclaims | Workbenches/Pipelines storage |
| `""` (core) | pods/exec | Operational tasks (not needed for KServe-only) |
| `""` (core) | pods/log | Debugging (can be removed for minimal RBAC) |
| `batch` | jobs, jobs/status, cronjobs | Monitoring/Ray controllers |
| `policy` | poddisruptionbudgets | LlamaStack controller |
| `snapshot.storage.k8s.io` | volumesnapshots | DSC controller |
| `machinelearning.seldon.io` | seldondeployments | Legacy Seldon check |
| `integreatly.org` | rhmis | RHMI integration |
| `machine.openshift.io` | machinesets, machineautoscalers | Ray controller |
| `oauth.openshift.io` | oauthclients | Dashboard/Auth |
| `security.openshift.io` | securitycontextconstraints | DSC controller (OpenShift) |
| `operator.openshift.io` | consoles, ingresscontrollers | DSC/Dashboard controller |
| `operators.coreos.com` | clusterserviceversions, subscriptions, catalogsources, operatorconditions, customresourcedefinitions | OLM integration (not on XKS) |
| `user.openshift.io` | users, groups | RHOAI-only (Dashboard) |
| `route.openshift.io` | routes, routes/custom-host | OpenShift routes (not on XKS) |
| `apiregistration.k8s.io` | apiservices | DSC controller |
| `*` (wildcard) | deployments, services, statefulsets, replicasets | Overly broad (replaceable with specific groups) |
| `extensions` | deployments, ingresses, replicasets | Deprecated API group |

---

## 3. Proposed Minimal ClusterRole for RHAII

The following ClusterRole contains only the permissions required for KServe and the operator
framework on an XKS (non-OpenShift) cluster:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: rhaii-controller-manager-role
rules:

# ============================================================
# Operator Framework (controller-runtime, reconciler, webhooks)
# ============================================================

# Core resources for manifest deployment
- apiGroups: [""]
  resources: [configmaps, configmaps/status]
  verbs: [get, create, watch, patch, delete, list, update]
- apiGroups: [""]
  resources: [secrets, secrets/finalizers]
  verbs: [create, delete, get, list, update, watch, patch]
- apiGroups: [""]
  resources: [services, services/finalizers]
  verbs: [create, delete, get, list, update, watch, patch]
- apiGroups: [""]
  resources: [serviceaccounts]
  verbs: [get, list, watch, create, update, patch, delete]
- apiGroups: [""]
  resources: [namespaces, namespaces/finalizers]
  verbs: [get, create, patch, delete, watch, update, list]
- apiGroups: [""]
  resources: [events]
  verbs: [get, create, watch, update, list, patch, delete]
- apiGroups: [""]
  resources: [pods]
  verbs: [get, list, watch, create, update, patch, delete]
- apiGroups: [""]
  resources: [endpoints]
  verbs: [watch, list, get, create, update, delete]

# Events API
- apiGroups: [events.k8s.io]
  resources: [events]
  verbs: [list, watch, patch, delete, get]

# Deployments and workloads
- apiGroups: [apps]
  resources: [deployments, deployments/finalizers]
  verbs: [get, list, watch, create, update, patch, delete]
- apiGroups: [apps]
  resources: [replicasets]
  verbs: [get, list, watch]
- apiGroups: [apps]
  resources: [statefulsets]
  verbs: [get, list, watch, create, update, patch, delete]

# Webhooks
- apiGroups: [admissionregistration.k8s.io]
  resources: [mutatingwebhookconfigurations, validatingwebhookconfigurations]
  verbs: [create, delete, get, list, patch, update, watch]
- apiGroups: [admissionregistration.k8s.io]
  resources: [validatingadmissionpolicies, validatingadmissionpolicybindings]
  verbs: [get, create, delete, update, watch, list, patch]

# CRDs
- apiGroups: [apiextensions.k8s.io]
  resources: [customresourcedefinitions]
  verbs: [get, list, watch, create, patch, delete, update]

# RBAC (needed to deploy component RBAC resources)
- apiGroups: [rbac.authorization.k8s.io]
  resources: [roles, rolebindings, clusterroles, clusterrolebindings]
  verbs: ["*"]

# Leader election
- apiGroups: [coordination.k8s.io]
  resources: [leases]
  verbs: [get, list, watch, create, update, patch, delete]

# Auth proxy
- apiGroups: [authentication.k8s.io]
  resources: [tokenreviews]
  verbs: [create, get]
- apiGroups: [authorization.k8s.io]
  resources: [subjectaccessreviews]
  verbs: [create, get]

# Cert-manager (TLS for webhooks)
# Issuers only need create/patch; certificates need full CRUD per PR #3409
- apiGroups: [cert-manager.io]
  resources: [issuers]
  verbs: [create, patch]
- apiGroups: [cert-manager.io]
  resources: [certificates]
  verbs: [get, list, watch, create, update, patch, delete]

# Network policies
- apiGroups: [networking.k8s.io]
  resources: [networkpolicies]
  verbs: [get, create, list, watch, delete, update, patch]

# Controller manager configuration
- apiGroups: [controller-runtime.sigs.k8s.io]
  resources: [controllermanagerconfigs]
  verbs: [get, create, patch, delete]

# ============================================================
# KServe Component
# ============================================================

# KServe CR
- apiGroups: [components.platform.opendatahub.io]
  resources: [kserves]
  verbs: [get, list, watch, create, update, patch, delete]
- apiGroups: [components.platform.opendatahub.io]
  resources: [kserves/status]
  verbs: [get, update, patch]
- apiGroups: [components.platform.opendatahub.io]
  resources: [kserves/finalizers]
  verbs: [update]

# KServe serving resources
- apiGroups: [serving.kserve.io]
  resources:
  - clusterservingruntimes
  - clusterservingruntimes/finalizers
  - clusterstoragecontainers
  - clusterstoragecontainers/finalizers
  - inferencegraphs
  - inferenceservices
  - inferenceservices/finalizers
  - llminferenceserviceconfigs
  - predictors
  - servingruntimes
  - servingruntimes/finalizers
  - trainedmodels
  verbs: [create, delete, get, list, patch, update, watch]
- apiGroups: [serving.kserve.io]
  resources:
  - clusterservingruntimes/status
  - clusterstoragecontainers/status
  - inferencegraphs/status
  - inferenceservices/status
  - predictors/status
  - trainedmodels/status
  verbs: [delete, get, patch, update]
- apiGroups: [serving.kserve.io]
  resources: [llminferenceserviceconfigs/status, predictors/finalizers, servingruntimes/status]
  verbs: [get, patch, update]
- apiGroups: [serving.kserve.io]
  resources: [llminferenceservices, llminferenceservices/status]
  verbs: [get, list, watch]

# LLM-d / Inference networking
- apiGroups: [inference.networking.x-k8s.io]
  resources: [inferencepools, inferencemodels]
  verbs: [get, list, watch]
- apiGroups: [inference.networking.k8s.io]
  resources: [inferencepools]
  verbs: [get, list, watch]

# KEDA autoscaling
- apiGroups: [keda.sh]
  resources: [kedacontrollers]
  verbs: [get, list, watch]
- apiGroups: [keda.sh]
  resources: [triggerauthentications]
  verbs: [create, delete, get, list, patch, update, watch]

# Metrics
- apiGroups: [metrics.k8s.io]
  resources: [pods, nodes]
  verbs: [get, list, watch]

# Operator dependencies check
- apiGroups: [operator.openshift.io]
  resources: [leaderworkersetoperators, jobsetoperators]
  verbs: [get, list, watch]

# Kuadrant check
- apiGroups: [kuadrant.io]
  resources: [kuadrants]
  verbs: [get, list, watch]

# Ingress domain discovery
- apiGroups: [config.openshift.io]
  resources: [ingresses]
  verbs: [get]

# Notebook integration
- apiGroups: [kubeflow.org]
  resources: [notebooks]
  verbs: [create, delete, get, list, patch, update, watch]
```

**Note:** The following rules from the full ClusterRole are intentionally excluded because the
platform type is `XKS` (non-OpenShift). If RHAII is also deployed on OpenShift (RHOAI variant),
additional rules would be needed for:
- `template.openshift.io/templates` (OpenShift templates used by KServe)
- `route.openshift.io/routes` (if applicable)
- `security.openshift.io/securitycontextconstraints`
- `config.openshift.io/authentications, clusterversions, infrastructures`
- `operators.coreos.com/*` (OLM resources)

---

## 4. Impact Analysis

### Rules removed: ~85 rules (out of ~130+)

| Category | Rules Removed | Examples |
|---|---|---|
| Disabled components (CR management) | ~45 | dashboards, rays, kueues, trainers, etc. |
| Disabled component backends | ~20 | ray.io, argoproj.io, kueue.x-k8s.io, modelregistry.opendatahub.io, etc. |
| Monitoring/Observability | ~25 | monitoring.coreos.com, monitoring.rhobs, perses.dev, tempo.grafana.com, opentelemetry.io |
| OpenShift-specific (XKS doesn't need) | ~15 | build.openshift.io, image.openshift.io, oauth.openshift.io, user.openshift.io |
| DSC/DSCI management | ~6 | datasciencecluster.opendatahub.io, dscinitialization.opendatahub.io, features.opendatahub.io |
| Gateway service | ~4 | gateway.networking.k8s.io, networking.istio.io, gatewayconfigs |
| Legacy/deprecated | ~5 | machinelearning.seldon.io, integreatly.org, extensions/* |
| Wildcard API groups | ~3 | `*` group rules (replaced by specific `apps` group) |

### Rules retained: ~45 rules

Covering: KServe serving resources, operator framework, RBAC management, webhooks, CRDs,
leader election, cert-manager, networking, auth proxy.

---

## 5. Recommendation: Separate ClusterRole vs Dynamic Generation

### Option A: Maintain a Separate Static ClusterRole (Recommended)

**Approach:** Create `config/rhaii/rbac/role.yaml` with the minimal ClusterRole and update
`config/rhaii/odh-operator/kustomization.yaml` to reference `../rbac` instead of `../../rbac`.

**Pros:**
- Simple, auditable, and reviewable in PRs
- No runtime complexity
- Consistent with existing Kustomize overlay pattern already used for RHAII
- Easy to validate with `kustomize build` and policy tools (e.g., OPA/Gatekeeper)
- The RHAII config directory already has its own CRD and webhook overrides -- RBAC fits naturally

**Cons:**
- Must be kept in sync when KServe RBAC requirements change
- Two ClusterRoles to maintain

### Option B: Generate ClusterRole Dynamically at Build Time

**Approach:** Parse `+kubebuilder:rbac` markers from only the KServe-related source files and
generate a role at build time (e.g., via a Makefile target).

**Pros:**
- Automatically stays in sync with code changes
- Single source of truth (the kubebuilder markers)

**Cons:**
- The current `kubebuilder_rbac.go` structure doesn't cleanly separate per-component rules --
  the DSC controller's file (`internal/controller/datasciencecluster/kubebuilder_rbac.go`)
  contains ALL component RBAC markers in a single file, making selective extraction fragile
- Framework-level rules (leader election, events, RBAC management) are implicitly required but
  not tagged to a specific component
- Would require custom tooling beyond `controller-gen`
- Higher maintenance burden on the build system

### Recommendation

**Option A (static separate ClusterRole)** is the better approach given the current codebase
structure. The RHAII deployment already uses dedicated Kustomize overlays for CRDs, webhooks,
and manager patches. Adding a scoped RBAC overlay follows the established pattern.

To mitigate the sync risk, a CI check could verify that any change to
`internal/controller/datasciencecluster/kubebuilder_rbac.go` lines tagged `// Kserve` triggers
a reminder to update the RHAII ClusterRole.

---

## 6. Runtime Validation Plan

To validate the proposed minimal ClusterRole:

1. **Deploy RHAII with the scoped ClusterRole** on an XKS cluster
2. **Create a KServe CR** (`config/rhaii/samples/kserve.yaml`)
3. **Deploy an InferenceService** and verify it becomes ready
4. **Check operator logs** for any `Forbidden` or `cannot` RBAC errors
5. **Verify LLM-d flow** (if LWS operator is present): create LLMInferenceServiceConfig
6. **Test cleanup**: delete the KServe CR and verify all owned resources are removed
7. **Run `kubectl auth can-i --list`** against the service account to confirm no excess permissions

---

## Appendix: File References

| File | Purpose |
|---|---|
| `config/rbac/role.yaml` | Current full ClusterRole (130+ rules) |
| `config/rhaii/odh-operator/kustomization.yaml` | RHAII kustomization (references full RBAC) |
| `config/rhaii/odh-operator/manager_patch.yaml` | RHAII env vars disabling components/services |
| `internal/controller/datasciencecluster/kubebuilder_rbac.go` | All component RBAC markers |
| `internal/controller/dscinitialization/kubebuilder_rbac.go` | DSCI/monitoring/observability RBAC markers |
| `internal/controller/services/gateway/kubebuilder_rbac.go` | Gateway service RBAC markers |
| `internal/controller/cloudmanager/common/kubebuilder_rbac.go` | Cloud manager shared RBAC |
| `internal/controller/components/kserve/kserve_controller.go` | KServe controller watches/owns |
| `internal/controller/components/kserve/kserve_controller_actions.go` | KServe runtime actions |
| `pkg/utils/flags/suppression.go` | Component/service suppression mechanism |

// Package rbac holds shared kubebuilder RBAC markers for the operator framework.
// These are the base permissions needed by any deployment mode (full ODH, RHAII, etc.)
// and are scanned by controller-gen to produce the generated ClusterRole.
package rbac

// Core resources
// +kubebuilder:rbac:groups="core",resources=configmaps/status,verbs=get;update;patch;delete
// +kubebuilder:rbac:groups="core",resources=configmaps,verbs=get;create;watch;patch;delete;list;update

// +kubebuilder:rbac:groups="core",resources=secrets,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="core",resources=secrets/finalizers,verbs=get;create;watch;update;patch;list;delete

// +kubebuilder:rbac:groups="core",resources=services/finalizers,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="core",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="*",resources=services,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="core",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="core",resources=namespaces/finalizers,verbs=update;list;watch;patch;delete;get
// +kubebuilder:rbac:groups="core",resources=namespaces,verbs=get;create;patch;delete;watch;update;list

// +kubebuilder:rbac:groups="core",resources=events,verbs=get;create;watch;update;list;patch;delete
// +kubebuilder:rbac:groups="events.k8s.io",resources=events,verbs=create;list;watch;patch;delete;get

// Log access for debugging and troubleshooting
// +kubebuilder:rbac:groups="core",resources=pods/log,verbs=get
// Exec access may be needed for certain operational tasks
// +kubebuilder:rbac:groups="core",resources=pods/exec,verbs=create
// Pod management for all workloads
// +kubebuilder:rbac:groups="core",resources=pods,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="core",resources=endpoints,verbs=watch;list;get;create;update;delete

// +kubebuilder:rbac:groups="core",resources=nodes,verbs=get;list;watch

// Deployment management for all components
// +kubebuilder:rbac:groups="apps",resources=deployments/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="*",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// Note: extensions API group is deprecated
// +kubebuilder:rbac:groups="extensions",resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Note: extensions API group is deprecated; ReplicaSets managed by Deployments
// +kubebuilder:rbac:groups="extensions",resources=replicasets,verbs=get;list;watch
// ReplicaSets managed by Deployments - controller needs to watch for status
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="*",resources=replicasets,verbs=get;list;watch

// Webhooks
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=validatingadmissionpolicybindings,verbs=get;create;delete;update;watch;list;patch
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=validatingadmissionpolicies,verbs=get;create;delete;update;watch;list;patch

// CRDs
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;patch;delete;update

// RBAC
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=*
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=*
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles,verbs=*
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=*

// Leader election
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;watch;create;update;patch;delete

// Cert-manager
// +kubebuilder:rbac:groups="cert-manager.io",resources=issuers,verbs=create;patch
// +kubebuilder:rbac:groups="cert-manager.io",resources=certificates,verbs=get;list;watch;create;update;patch;delete

// Networking
// +kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies,verbs=get;create;list;watch;delete;update;patch
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=create;delete;list;update;watch;patch;get

// Monitoring (ServiceMonitors are used across components)
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;create;delete;update;watch;list;patch;deletecollection

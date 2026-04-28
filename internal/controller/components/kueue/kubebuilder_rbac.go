package kueue

// Kueue
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=kueues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=kueues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=kueues/finalizers,verbs=update
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheusrules,verbs=get;create;patch;delete;deletecollection;list;watch;update
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=podmonitors,verbs=get;create;delete;update;watch;list;patch
// +kubebuilder:rbac:groups="kueue.x-k8s.io",resources=clusterqueues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="kueue.x-k8s.io",resources=clusterqueues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="kueue.x-k8s.io",resources=localqueues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="kueue.x-k8s.io",resources=localqueues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="kueue.x-k8s.io",resources=resourceflavors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="kueue.openshift.io",resources=kueues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="kueue.openshift.io",resources=kueues/status,verbs=get;update;patch

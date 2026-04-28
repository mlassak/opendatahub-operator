package ray

// Ray
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=rays,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=rays/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=rays/finalizers,verbs=update
// +kubebuilder:rbac:groups="ray.io",resources=rayservices,verbs=create;delete;list;watch;update;patch;get
// +kubebuilder:rbac:groups="ray.io",resources=rayjobs,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="ray.io",resources=rayclusters,verbs=create;delete;list;patch;get
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=watch;create;update;delete;list;patch;get
// +kubebuilder:rbac:groups="autoscaling.openshift.io",resources=machinesets,verbs=list;patch;delete;get
// +kubebuilder:rbac:groups="autoscaling.openshift.io",resources=machineautoscalers,verbs=list;patch;delete;get
// +kubebuilder:rbac:groups="batch",resources=jobs/status,verbs=get;list;watch;create;update;patch;delete
// Monitoring controller manages Jobs through Owns() relationship
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=get;list;watch;create;update;patch;delete;get

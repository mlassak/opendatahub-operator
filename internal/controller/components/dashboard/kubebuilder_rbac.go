package dashboard

// Dashboard
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=dashboards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=dashboards/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=dashboards/finalizers,verbs=create;get;list;patch;update;use;watch
// +kubebuilder:rbac:groups="opendatahub.io",resources=odhdashboardconfigs,verbs=create;get;patch;watch;update;delete;list
// +kubebuilder:rbac:groups="console.openshift.io",resources=odhquickstarts,verbs=create;get;patch;list;delete;watch;update
// +kubebuilder:rbac:groups="dashboard.opendatahub.io",resources=odhdocuments,verbs=create;get;patch;list;delete;watch;update
// +kubebuilder:rbac:groups="dashboard.opendatahub.io",resources=odhapplications,verbs=create;get;patch;list;delete;watch;update
// +kubebuilder:rbac:groups="dashboard.opendatahub.io",resources=acceleratorprofiles,verbs=create;get;patch;list;delete;watch;update
// +kubebuilder:rbac:groups="dashboard.opendatahub.io",resources=hardwareprofiles,verbs=get;list;watch;update

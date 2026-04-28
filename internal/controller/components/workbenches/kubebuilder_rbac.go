package workbenches

// WB
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=workbenches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=workbenches/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=workbenches/finalizers,verbs=update
// +kubebuilder:rbac:groups="image.openshift.io",resources=imagestreamtags,verbs=get
// +kubebuilder:rbac:groups="image.openshift.io",resources=imagestreams,verbs=patch;create;update;delete;get
// +kubebuilder:rbac:groups="image.openshift.io",resources=imagestreams,verbs=create;list;watch;patch;delete;get
// OpenVino still need buildconfig
// +kubebuilder:rbac:groups="build.openshift.io",resources=builds,verbs=create;patch;delete;list;watch;get
// +kubebuilder:rbac:groups="build.openshift.io",resources=buildconfigs/instantiate,verbs=create;patch;delete;get;list;watch
// +kubebuilder:rbac:groups="build.openshift.io",resources=buildconfigs,verbs=list;watch;create;patch;delete;get;update

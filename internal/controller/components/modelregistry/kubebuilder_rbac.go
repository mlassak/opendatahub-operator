package modelregistry

// ModelRegistry
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelregistries/finalizers,verbs=update
// +kubebuilder:rbac:groups=modelregistry.opendatahub.io,resources=modelregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=modelregistry.opendatahub.io,resources=modelregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=modelregistry.opendatahub.io,resources=modelregistries/finalizers,verbs=update;get

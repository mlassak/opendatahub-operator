package modelcontroller

// ModelController
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelcontrollers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelcontrollers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelcontrollers/finalizers,verbs=update

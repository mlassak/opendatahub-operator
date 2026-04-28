package llamastackoperator

// LlamaStackOperator
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=llamastackoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=llamastackoperators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=llamastackoperators/finalizers,verbs=update
// +kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

package trainingoperator

// TrainingOperator
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=trainingoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=trainingoperators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=trainingoperators/finalizers,verbs=update

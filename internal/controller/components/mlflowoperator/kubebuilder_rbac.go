package mlflowoperator

// MLflowOperator
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=mlflowoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=mlflowoperators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=mlflowoperators/finalizers,verbs=update
// +kubebuilder:rbac:groups=mlflow.opendatahub.io,resources=mlflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mlflow.opendatahub.io,resources=mlflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mlflow.opendatahub.io,resources=mlflows/finalizers,verbs=update

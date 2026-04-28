package feastoperator

// FeastOperator
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=feastoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=feastoperators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=feastoperators/finalizers,verbs=update

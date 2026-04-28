package trainer

// Trainer
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=trainers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=trainers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=trainers/finalizers,verbs=update
// +kubebuilder:rbac:groups=trainer.kubeflow.org,resources=clustertrainingruntimes,verbs=get;list;watch;create;update;patch;delete

package datasciencepipelines

// DataSciencePipelines
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=datasciencepipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=datasciencepipelines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=datasciencepipelines/finalizers,verbs=update
// +kubebuilder:rbac:groups="datasciencepipelinesapplications.opendatahub.io",resources=datasciencepipelinesapplications/status,verbs=update;patch;get
// +kubebuilder:rbac:groups="datasciencepipelinesapplications.opendatahub.io",resources=datasciencepipelinesapplications/finalizers,verbs=update;patch;get
// +kubebuilder:rbac:groups="datasciencepipelinesapplications.opendatahub.io",resources=datasciencepipelinesapplications,verbs=create;delete;list;update;watch;patch;get
// Argo workflows for data science pipelines
// +kubebuilder:rbac:groups="argoproj.io",resources=workflows,verbs=get;list;watch;create;update;patch;delete

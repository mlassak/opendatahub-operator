package kserve

// Kserve
// +kubebuilder:rbac:groups="kubeflow.org",resources=notebooks,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=kserves,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=kserves/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=kserves/finalizers,verbs=update
// +kubebuilder:rbac:groups="serving.kserve.io",resources=trainedmodels/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=trainedmodels,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=servingruntimes/status,verbs=update;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=servingruntimes/finalizers,verbs=create;delete;list;update;watch;patch;get
// KServe serving runtimes for model serving
// +kubebuilder:rbac:groups="serving.kserve.io",resources=servingruntimes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serving.kserve.io",resources=predictors/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=predictors/finalizers,verbs=update;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=predictors,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferenceservices/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferenceservices/finalizers,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferenceservices,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferencegraphs/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferencegraphs,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterservingruntimes/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterservingruntimes/finalizers,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterservingruntimes,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterstoragecontainers/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterstoragecontainers/finalizers,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterstoragecontainers,verbs=create;delete;list;update;watch;patch;get
// OpenShift templates for workbenches
// +kubebuilder:rbac:groups="template.openshift.io",resources=templates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="config.openshift.io",resources=ingresses,verbs=get
/* KEDA (CMA) InferenceService autoscaling */
// +kubebuilder:rbac:groups=keda.sh,resources=kedacontrollers,verbs=get;list;watch
// +kubebuilder:rbac:groups=keda.sh,resources=triggerauthentications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods;nodes,verbs=get;list;watch
/* LLM-d */
// +kubebuilder:rbac:groups="serving.kserve.io",resources=llminferenceserviceconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serving.kserve.io",resources=llminferenceserviceconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="serving.kserve.io",resources=llminferenceservices,verbs=get;list;watch
// +kubebuilder:rbac:groups="serving.kserve.io",resources=llminferenceservices/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="inference.networking.x-k8s.io",resources=inferencepools,verbs=get;list;watch
// +kubebuilder:rbac:groups="inference.networking.x-k8s.io",resources=inferencemodels,verbs=get;list;watch
// +kubebuilder:rbac:groups="inference.networking.k8s.io",resources=inferencepools,verbs=get;list;watch
// +kubebuilder:rbac:groups="kuadrant.io",resources=kuadrants,verbs=get;list;watch
// +kubebuilder:rbac:groups="operator.openshift.io",resources=leaderworkersetoperators,verbs=get;list;watch

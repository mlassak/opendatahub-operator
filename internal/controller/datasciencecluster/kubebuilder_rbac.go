package datasciencecluster

// +kubebuilder:rbac:groups="datasciencecluster.opendatahub.io",resources=datascienceclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="datasciencecluster.opendatahub.io",resources=datascienceclusters/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups="datasciencecluster.opendatahub.io",resources=datascienceclusters,verbs=get;list;watch;create;update;patch;delete;deletecollection

// +kubebuilder:rbac:groups="authentication.k8s.io",resources=tokenreviews,verbs=create;get
// +kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=create;get

// +kubebuilder:rbac:groups="operators.coreos.com",resources=clusterserviceversions,verbs=get;list;watch;delete;update
// +kubebuilder:rbac:groups="operators.coreos.com",resources=customresourcedefinitions,verbs=create;get;patch;delete
// +kubebuilder:rbac:groups="operators.coreos.com",resources=subscriptions,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups="operators.coreos.com",resources=operatorconditions,verbs=get;list;watch
// +kubebuilder:rbac:groups="operators.coreos.com",resources=catalogsources,verbs=get;list;watch

// +kubebuilder:rbac:groups="snapshot.storage.k8s.io",resources=volumesnapshots,verbs=create;delete;patch;get

// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,verbs=get;list;watch;create;patch;use,resourceNames=restricted
// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,verbs=get;list;watch;create;patch;use,resourceNames=anyuid
// +kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,verbs=get;list;watch;create;patch;use

// +kubebuilder:rbac:groups="route.openshift.io",resources=routes,verbs=get;list;watch;create;delete;update;patch

// +kubebuilder:rbac:groups="apiregistration.k8s.io",resources=apiservices,verbs=create;delete;list;watch;update;patch;get

// +kubebuilder:rbac:groups="operator.openshift.io",resources=consoles,verbs=get;list;watch;patch;delete
// +kubebuilder:rbac:groups="operator.openshift.io",resources=ingresscontrollers,verbs=get;list;watch;patch;delete

// +kubebuilder:rbac:groups="oauth.openshift.io",resources=oauthclients,verbs=create;delete;list;watch;update;patch;get

// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=podmonitors,verbs=get;create;delete;update;watch;list;patch
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheusrules,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheuses,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheuses/finalizers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheuses/status,verbs=get;create;patch;delete;deletecollection

// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagers/finalizers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagers/status,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagerconfigs,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=thanosrulers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=thanosrulers/finalizers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=thanosrulers/status,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=probes,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheusrules,verbs=get;create;patch;delete;deletecollection

// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheuses/finalizers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=prometheuses/status,verbs=get;create;patch;delete;deletecollection

// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagers/finalizers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagers/status,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=alertmanagerconfigs,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=thanosrulers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=thanosrulers/finalizers,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=thanosrulers/status,verbs=get;create;patch;delete;deletecollection
// +kubebuilder:rbac:groups="monitoring.coreos.com",resources=probes,verbs=get;create;patch;delete;deletecollection

// Controller may need to check for Seldon CRD presence for legacy support
// +kubebuilder:rbac:groups="machinelearning.seldon.io",resources=seldondeployments,verbs=get;list;watch

// +kubebuilder:rbac:groups="machine.openshift.io",resources=machinesets,verbs=list;patch;delete;get
// +kubebuilder:rbac:groups="machine.openshift.io",resources=machineautoscalers,verbs=list;patch;delete;get

// +kubebuilder:rbac:groups="integreatly.org",resources=rhmis,verbs=list;watch;patch;delete;get

// +kubebuilder:rbac:groups="extensions",resources=ingresses,verbs=list;watch;patch;delete;get

// Storage resources for workbenches, pipelines, and data persistence
// +kubebuilder:rbac:groups="core",resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="core",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups="core",resources=clusterversions,verbs=watch;list;get

// +kubebuilder:rbac:groups="config.openshift.io",resources=clusterversions,verbs=watch;list;get

// +kubebuilder:rbac:groups="core",resources=rhmis,verbs=watch;list;get

// +kubebuilder:rbac:groups="controller-runtime.sigs.k8s.io",resources=controllermanagerconfigs,verbs=get;create;patch;delete

// StatefulSet management for stateful components
// +kubebuilder:rbac:groups="*",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

/* Only for RHOAI */
// +kubebuilder:rbac:groups="user.openshift.io",resources=users,verbs=list;watch;patch;delete;get;update
// +kubebuilder:rbac:groups="user.openshift.io",resources=groups,verbs=get;create;list;watch;patch;delete
// +kubebuilder:rbac:groups="console.openshift.io",resources=consolelinks,verbs=create;get;patch;list;delete;watch;update

// CFO
//+kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=codeflares,verbs=get;list;watch
//+kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=codeflares/status,verbs=get

// ModelMeshServing
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelmeshservings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelmeshservings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelmeshservings/finalizers,verbs=update

// HardwareProfile
// +kubebuilder:rbac:groups=infrastructure.opendatahub.io,resources=hardwareprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.opendatahub.io,resources=hardwareprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.opendatahub.io,resources=hardwareprofiles/finalizers,verbs=update

// +kubebuilder:rbac:groups="operator.openshift.io",resources=jobsetoperators,verbs=get;list;watch

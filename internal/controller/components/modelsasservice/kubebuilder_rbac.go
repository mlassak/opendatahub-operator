package modelsasservice

// Tenant (read status for DSC mirroring; delete on disable; watch for OwnsGVK)
// +kubebuilder:rbac:groups=maas.opendatahub.io,resources=tenants,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups=maas.opendatahub.io,resources=tenants/status,verbs=get

// Models-as-a-Service
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelsasservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelsasservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=components.platform.opendatahub.io,resources=modelsasservices/finalizers,verbs=update
// +kubebuilder:rbac:groups=kuadrant.io,resources=authpolicies;tokenratelimitpolicies;ratelimitpolicies;telemetrypolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions.kuadrant.io,resources=telemetrypolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.authorino.kuadrant.io,resources=authorinos,verbs=get;list
// +kubebuilder:rbac:groups=telemetry.istio.io,resources=telemetries,verbs=get;list;watch;create;update;patch;delete

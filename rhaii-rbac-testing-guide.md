# Testing CCM (Azure) on KinD with RBAC Validation

**Ticket:** RHOAIENG-54569
**Date:** 2026-04-15

Guide for building, deploying, and testing the Azure Cloud Manager and RHAII operator on a KinD cluster to validate the RBAC subset.

Throughout this guide, `$CLUSTER_NAME` refers to the name of your KinD cluster (e.g. `rhaii-rbac-test`).

---

## Prerequisites

- A running KinD cluster
- `kubectl` configured to reach the cluster
- `podman` for building images
- `kind` CLI installed
- Valid `registry.redhat.io` credentials in your Podman auth file (typically `$XDG_RUNTIME_DIR/containers/auth.json`)

```bash
# Confirm cluster is ready
kubectl --context kind-$CLUSTER_NAME get nodes
```

---

## Step 1: Build the operator image

The single Dockerfile (`Dockerfiles/Dockerfile`) builds **both** the main operator binary (`cmd/main.go`) and the `cloudmanager` binary (`cmd/cloudmanager/`) into one image. Both RHAII and CCM use the same image.

```bash
make image-build IMG=localhost/odh-operator:rbac-test
```

---

## Step 2: Load the image into KinD

```bash
make image-kind-load IMG=localhost/odh-operator:rbac-test KIND_CLUSTER_NAME=$CLUSTER_NAME
```

---

## Step 3: Deploy the RHAII operator (KServe-only)

```bash
make deploy-rhaii-local IMG=localhost/odh-operator:rbac-test
```

This does the following:
- Runs `make prepare` (generates CRDs, RBAC, webhooks, sets the image in kustomization)
- Builds kustomize overlay from `config/rhaii/odh-local/` which layers on top of `config/rhaii/odh-operator/`
- Substitutes `REPLACE_RHAI_VERSION` with `3.3.0`
- Applies to namespace `opendatahub-operator-system`

The RHAII operator pod will start with:
- `ODH_PLATFORM_TYPE=XKS`
- `RHAI_APPLICATIONS_NAMESPACE=opendatahub`
- `RHAI_VERSION=3.3.0`
- All components/services disabled except KServe
- `imagePullPolicy: IfNotPresent` (the `odh-local` patch)

**Workaround: webhook certificate secret.** The operator pods mount a TLS secret
(`opendatahub-operator-controller-webhook-cert`) for the webhook server. On a fresh
cluster cert-manager is not yet installed (CCM does that later), so the secret does
not exist and the pods will stay in `ContainerCreating`. Create a placeholder
self-signed certificate to unblock them:

```bash
openssl req -x509 -newkey rsa:2048 -keyout /tmp/tls.key -out /tmp/tls.crt \
  -days 365 -nodes -subj "/CN=webhook-service.opendatahub-operator-system.svc"

kubectl -n opendatahub-operator-system create secret tls \
  opendatahub-operator-controller-webhook-cert \
  --cert=/tmp/tls.crt --key=/tmp/tls.key
```

Once CCM later bootstraps cert-manager and its PKI, the real certificate will be
issued and will replace this placeholder.

Verify:

```bash
kubectl -n opendatahub-operator-system get pods
kubectl -n opendatahub-operator-system logs deploy/opendatahub-operator-controller-manager -f
```

The operator will start but KServe reconciliation won't proceed until its dependencies (cert-manager, Istio CRDs) are available -- that's CCM's job.

---

## Step 4: Deploy the Azure Cloud Manager

```bash
make deploy-ccm-local-azure IMG=localhost/odh-operator:rbac-test
```

This:
- Runs `make manifests-ccm-azure` (generates `AzureKubernetesEngine` CRD + RBAC from kubebuilder markers)
- Sets the image in `config/cloudmanager/azure/manager/kustomization.yaml`
- Builds kustomize from `config/cloudmanager/azure/local/` (adds `imagePullPolicy: IfNotPresent`)
- Applies to namespace `opendatahub-cloudmanager-system`

Deploys:
- `AzureKubernetesEngine` CRD
- `opendatahub-azure-cloud-manager-role` ClusterRole (the CCM RBAC)
- ClusterRoleBinding + ServiceAccount
- Leader election Role/RoleBinding
- `azure-cloud-manager-operator` Deployment

Verify:

```bash
kubectl -n opendatahub-cloudmanager-system get pods
kubectl -n opendatahub-cloudmanager-system logs deploy/azure-cloud-manager-operator -f
```

The CCM pod will start but do nothing until you create the CR.

---

## Step 5: Set up pull secrets

The CCM-managed dependencies (cert-manager, sail-operator, LWS) pull images from
`registry.redhat.io`, which requires authentication. The Helm charts already
configure the service accounts in the dependency namespaces with
`imagePullSecrets: [{name: rhaii-pull-secret}]`, but the secret itself must be
created. **Without this, all dependency pods will fail with `ImagePullBackOff`.**

With Podman, the auth file is typically at `$XDG_RUNTIME_DIR/containers/auth.json`:

```bash
make kind-setup-pull-secrets PULL_SECRET=$XDG_RUNTIME_DIR/containers/auth.json
```

This creates (or updates) a `kubernetes.io/dockerconfigjson` secret named
`rhaii-pull-secret` in each of these namespaces (creating the namespace if needed):
`cert-manager`, `cert-manager-operator`, `openshift-lws-operator`, `istio-system`.

If you have already created the AzureKubernetesEngine CR (Step 6) before running
this, the dependency pods will be stuck in `ImagePullBackOff`. Delete them to
trigger a re-pull with the new credentials:

```bash
kubectl delete pod -n cert-manager-operator --all
kubectl delete pod -n istio-system --all
kubectl delete pod -n openshift-lws-operator --all
```

---

## Step 6: Create the AzureKubernetesEngine CR

```bash
kubectl apply -f config/cloudmanager/azure/samples/azurekubernetesengine_v1alpha1.yaml
```

This creates:

```yaml
apiVersion: infrastructure.opendatahub.io/v1alpha1
kind: AzureKubernetesEngine
metadata:
  name: default-azurekubernetesengine
spec:
  dependencies:
    gatewayAPI:
      managementPolicy: Managed
    certManager:
      managementPolicy: Managed
    lws:
      managementPolicy: Managed
    sailOperator:
      managementPolicy: Managed
```

The CCM reconciler kicks in and:

1. Renders Helm charts from `/opt/charts/` (gateway-api, cert-manager-operator, lws-operator, sail-operator)
2. Deploys them via Server-Side Apply
3. Bootstraps PKI: self-signed ClusterIssuer -> root CA Certificate -> CA-backed ClusterIssuer
4. Creates a webhook certificate for the RHAII operator
5. Runs GC to clean up stale resources

---

## Step 7: Monitor progress

```bash
# Watch CCM logs for RBAC errors
kubectl -n opendatahub-cloudmanager-system logs deploy/azure-cloud-manager-operator -f

# Check the CR status
kubectl get azurekubernetesengine default-azurekubernetesengine -o yaml

# Check the conditions -- you want Ready=True, DependenciesAvailable=True
kubectl get azurekubernetesengine default-azurekubernetesengine \
  -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\t"}{.message}{"\n"}{end}'

# Check that dependency namespaces and deployments were created
kubectl get ns cert-manager-operator istio-system openshift-lws-operator
kubectl get deploy -n cert-manager-operator
kubectl get deploy -n istio-system
kubectl get deploy -n openshift-lws-operator

# Check PKI resources
kubectl get clusterissuers
kubectl get certificates -n cert-manager
```

---

## Step 8: Validate RBAC -- the key test

Now that CCM is running with the **generated** ClusterRole (`opendatahub-azure-cloud-manager-role`), check for any RBAC errors:

```bash
# Check for Forbidden errors in CCM logs
kubectl -n opendatahub-cloudmanager-system logs deploy/azure-cloud-manager-operator \
  | grep -i -E 'forbidden|cannot|unauthorized|error'

# Check for Forbidden errors in RHAII operator logs
kubectl -n opendatahub-operator-system logs deploy/opendatahub-operator-controller-manager \
  | grep -i -E 'forbidden|cannot|unauthorized|error'

# Verify the actual permissions of the CCM service account
kubectl auth can-i --list \
  --as=system:serviceaccount:opendatahub-cloudmanager-system:azure-cloud-manager-operator

# Verify the actual permissions of the RHAII operator service account
kubectl auth can-i --list \
  --as=system:serviceaccount:opendatahub-operator-system:controller-manager
```

---

## Step 9: Test KServe reconciliation

Once CCM has bootstrapped dependencies (cert-manager CRDs available, Istio CRDs available), create a KServe CR to trigger the RHAII operator's reconciliation:

```bash
kubectl apply -f config/rhaii/samples/kserve.yaml
```

If that file doesn't exist, create it manually:

```bash
kubectl apply -f - <<'EOF'
apiVersion: components.platform.opendatahub.io/v1alpha1
kind: Kserve
metadata:
  name: default-kserve
EOF
```

Monitor:

```bash
# Watch RHAII operator reconciling KServe
kubectl -n opendatahub-operator-system logs deploy/opendatahub-operator-controller-manager -f

# Check KServe status
kubectl get kserve default-kserve -o yaml
```

---

## Step 10: Test with a scoped-down RBAC (your actual validation)

To test the **proposed minimal ClusterRole** from `rhaii-rbac-analysis.md` Section 3:

1. Export the current generated ClusterRoles for reference:

   ```bash
   kubectl get clusterrole opendatahub-operator-controller-manager-role -o yaml > ./original-role.yaml
   kubectl get clusterrole opendatahub-azure-cloud-manager-role -o yaml > ./original-ccm-role.yaml
   ```

2. Create your scoped-down ClusterRole YAML (from `rhaii-rbac-analysis.md` Section 3, plus fixes from `rhaii-rbac-gap-analysis.md`).

3. Replace the ClusterRole:

   ```bash
   kubectl apply -f your-minimal-role.yaml
   ```

4. Restart the operator to force a fresh reconciliation with the new RBAC:

   ```bash
   kubectl -n opendatahub-operator-system rollout restart deploy/opendatahub-operator-controller-manager
   ```

5. Re-check for RBAC errors (Step 8).

6. Delete and recreate the KServe CR to test the full lifecycle:

   ```bash
   kubectl delete kserve default-kserve
   # Wait for cleanup
   kubectl apply -f - <<'EOF'
   apiVersion: components.platform.opendatahub.io/v1alpha1
   kind: Kserve
   metadata:
     name: default-kserve
   EOF
   ```

---

## Cleanup

```bash
# Remove KServe CR
kubectl delete kserve default-kserve --ignore-not-found

# Remove AzureKubernetesEngine CR (triggers cascade deletion of managed dependencies)
kubectl delete azurekubernetesengine default-azurekubernetesengine --ignore-not-found

# Undeploy CCM
make undeploy-ccm-azure

# Undeploy RHAII operator
make undeploy-rhaii
```

---

## Important Notes

- **CCM has its own separate ClusterRole** (`opendatahub-azure-cloud-manager-role` at `config/cloudmanager/azure/rbac/role.yaml`). This is distinct from the main operator's `controller-manager-role`. When testing RBAC reduction, these are **two separate roles to scope down independently**.

- **On KinD, some dependencies may not fully start** -- e.g., the sail-operator and cert-manager images may not be pullable without pull secrets or may have platform-specific issues. This is expected. The key validation is whether the **RBAC allows the controller to perform its API calls** without `Forbidden` errors, not whether the deployed operators themselves become healthy.

- **The GC action uses `SelfSubjectRulesReview`** (Gap 1 from the gap analysis). If you don't include `authorization.k8s.io/selfsubjectrulesreviews` with `create` verb in your scoped-down role, you'll see errors on every reconciliation cycle after the first deployment.

- **The HardwareProfile webhook is active when KServe is enabled** (Gap 2). If you exclude `infrastructure.opendatahub.io/hardwareprofiles`, the webhook will fail when any InferenceService or Notebook with a hardware profile annotation is created/updated.

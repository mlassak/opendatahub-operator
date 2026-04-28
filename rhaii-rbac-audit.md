# RHAII RBAC Audit Validation

**Ticket:** RHOAIENG-54569

Validates the generated RHAII ClusterRole against actual runtime API usage using
Kubernetes audit logging. Deploys the RHAII operator on a KinD cluster with audit
logging enabled, exercises all reconciliation paths, then compares the audit log
against the generated role to find:

- **Over-grants:** permissions in the role that were never exercised (candidates for removal)
- **Under-grants:** API calls that would fail without the current role (missing permissions)

## Prerequisites

- `kind`, `kubectl`, `podman` (or `docker`), `jq`
- The operator image built and available locally

## Quick start

```bash
# 1. Build the operator
make image-build IMG=localhost/odh-operator:rbac-audit

# 2. Run the audit
#    This creates a KinD cluster with audit logging, deploys RHAII,
#    exercises KServe reconciliation, and produces the report.
bash hack/rhaii-rbac-audit.sh localhost/odh-operator:rbac-audit

# 3. Review the output
cat /tmp/rhaii-rbac-audit-report.txt
```

## What the script does

1. Creates a KinD cluster (`rhaii-rbac-audit`) with an audit policy that logs all
   API calls made by the RHAII operator's service account at `RequestResponse` level.
2. Loads the operator image into KinD.
3. Deploys the RHAII operator (`make deploy-rhaii-local`).
4. Creates a placeholder webhook cert (cert-manager is not available in this minimal setup).
5. Waits for the operator to start.
6. Creates a KServe CR to trigger reconciliation.
7. Waits for reconciliation to settle.
8. Exercises additional code paths: spec mutation, delete/recreate.
9. Extracts the audit log from the KinD control-plane node.
10. Parses the audit log to extract unique (apiGroup, resource, verb) tuples actually used.
11. Parses the generated RHAII ClusterRole (`config/rhaii/rbac/role.yaml`).
12. Produces a diff report showing over-grants and any 403 Forbidden calls.

## Interpreting the report

### Over-grants

Rules in the ClusterRole that were never exercised. Not all of these are safe to
remove -- some permissions are needed for rare code paths (error recovery, GC of
resources that didn't exist in this test, CRD watches that only fire when CRDs are
installed). Use judgment:

- **Safe to remove:** permissions for API groups that don't exist on the cluster and
  are not dynamically watched (e.g., `kubeflow.org/notebooks` if KServe doesn't
  actually use it)
- **Keep despite no usage:** permissions for resources the controller dynamically
  discovers via CRD watches (e.g., `serving.kserve.io/*` resources that only exist
  after KServe is fully deployed)

### Forbidden calls (under-grants)

API calls that returned 403. These indicate missing permissions in the generated
role. Should be zero -- if any appear, add the corresponding marker to the
appropriate `kubebuilder_rbac.go` file.

## Cleanup

```bash
kind delete cluster --name rhaii-rbac-audit
```

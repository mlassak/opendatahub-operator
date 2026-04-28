#!/usr/bin/env bash
# rhaii-rbac-audit.sh — Deploy RHAII on KinD with audit logging, exercise
# reconciliation, then diff actual API usage against the generated ClusterRole.
#
# Usage: bash hack/rhaii-rbac-audit.sh <operator-image>
# Example: bash hack/rhaii-rbac-audit.sh localhost/odh-operator:rbac-audit

set -euo pipefail

IMG="${1:?Usage: $0 <operator-image>}"
CLUSTER_NAME="rhaii-rbac-audit"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
AUDIT_POLICY="$SCRIPT_DIR/rhaii-audit-policy.yaml"
REPORT="/tmp/rhaii-rbac-audit-report.txt"
ROLE_FILE="$REPO_ROOT/config/rhaii/rbac/role.yaml"

# ---------------------------------------------------------------------------
# Colours (disabled when stdout is not a terminal)
# ---------------------------------------------------------------------------
if [ -t 1 ]; then
  BOLD='\033[1m' GREEN='\033[32m' YELLOW='\033[33m' RED='\033[31m' RESET='\033[0m'
else
  BOLD='' GREEN='' YELLOW='' RED='' RESET=''
fi
info()  { echo -e "${GREEN}[INFO]${RESET}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${RESET}  $*"; }
error() { echo -e "${RED}[ERROR]${RESET} $*"; }
step()  { echo -e "\n${BOLD}=== $* ===${RESET}"; }

# ---------------------------------------------------------------------------
# Step 1: Create KinD cluster with audit logging
# ---------------------------------------------------------------------------
step "Creating KinD cluster with audit logging"

if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  warn "Cluster '$CLUSTER_NAME' already exists, deleting it first"
  kind delete cluster --name "$CLUSTER_NAME"
fi

# KinD config with audit logging enabled
cat <<KINDEOF | kind create cluster --name "$CLUSTER_NAME" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: ClusterConfiguration
        apiServer:
          extraArgs:
            audit-policy-file: /etc/kubernetes/audit/audit-policy.yaml
            audit-log-path: /var/log/kubernetes/audit.log
            audit-log-maxsize: "100"
            audit-log-maxbackup: "1"
          extraVolumes:
            - name: audit-policy
              hostPath: /etc/kubernetes/audit/audit-policy.yaml
              mountPath: /etc/kubernetes/audit/audit-policy.yaml
              readOnly: true
              pathType: File
            - name: audit-log
              hostPath: /var/log/kubernetes/
              mountPath: /var/log/kubernetes/
              pathType: DirectoryOrCreate
    extraMounts:
      - hostPath: ${AUDIT_POLICY}
        containerPath: /etc/kubernetes/audit/audit-policy.yaml
        readOnly: true
KINDEOF

kubectl --context "kind-${CLUSTER_NAME}" cluster-info
info "Cluster created with audit logging enabled"

# ---------------------------------------------------------------------------
# Step 2: Load image into KinD
# ---------------------------------------------------------------------------
step "Loading operator image into KinD"
make -C "$REPO_ROOT" image-kind-load IMG="$IMG" KIND_CLUSTER_NAME="$CLUSTER_NAME"

# ---------------------------------------------------------------------------
# Step 3: Deploy RHAII operator
# ---------------------------------------------------------------------------
step "Deploying RHAII operator"
make -C "$REPO_ROOT" deploy-rhaii-local IMG="$IMG"

# Create placeholder webhook cert (no cert-manager in this minimal cluster)
info "Creating placeholder webhook certificate"
TMPDIR=$(mktemp -d)
openssl req -x509 -newkey rsa:2048 \
  -keyout "$TMPDIR/tls.key" -out "$TMPDIR/tls.crt" \
  -days 1 -nodes -subj "/CN=webhook.opendatahub-operator-system.svc" 2>/dev/null

kubectl -n opendatahub-operator-system create secret tls \
  opendatahub-operator-controller-webhook-cert \
  --cert="$TMPDIR/tls.crt" --key="$TMPDIR/tls.key" \
  --dry-run=client -o yaml | kubectl apply -f -
rm -rf "$TMPDIR"

# ---------------------------------------------------------------------------
# Step 4: Wait for operator pod
# ---------------------------------------------------------------------------
step "Waiting for operator pod to start"
kubectl -n opendatahub-operator-system wait --for=condition=Available \
  deployment/opendatahub-operator-controller-manager --timeout=120s || {
    warn "Deployment not Available yet, checking pod status..."
    kubectl -n opendatahub-operator-system get pods
    kubectl -n opendatahub-operator-system logs deploy/opendatahub-operator-controller-manager --tail=30 || true
    # Continue anyway — the pod may be running but not fully ready
}

info "Operator pod is running"
# Give it a few seconds to start its informers
sleep 10

# ---------------------------------------------------------------------------
# Step 5: Exercise KServe reconciliation
# ---------------------------------------------------------------------------
step "Creating KServe CR"
kubectl apply -f - <<'EOF'
apiVersion: components.platform.opendatahub.io/v1alpha1
kind: Kserve
metadata:
  name: default-kserve
spec: {}
EOF

info "Waiting for reconciliation to settle (60s)..."
sleep 60

# Check status
info "KServe CR status:"
kubectl get kserve default-kserve -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\t"}{.reason}{"\n"}{end}' 2>/dev/null || warn "Could not read KServe status"

# ---------------------------------------------------------------------------
# Step 6: Exercise additional code paths
# ---------------------------------------------------------------------------
step "Exercising additional code paths"

# Mutate spec to trigger re-reconciliation
info "Patching KServe spec..."
kubectl patch kserve default-kserve --type=merge \
  -p '{"spec":{"rawDeploymentServiceConfig":"Headed"}}' 2>/dev/null || warn "Patch failed (may be expected if CRD validation rejects)"
sleep 20

# Revert
kubectl patch kserve default-kserve --type=merge \
  -p '{"spec":{"rawDeploymentServiceConfig":"Headless"}}' 2>/dev/null || true
sleep 20

# Delete and recreate to exercise GC + finalizers
info "Delete/recreate KServe CR..."
kubectl delete kserve default-kserve --timeout=30s 2>/dev/null || true
sleep 15
kubectl apply -f - <<'EOF'
apiVersion: components.platform.opendatahub.io/v1alpha1
kind: Kserve
metadata:
  name: default-kserve
spec: {}
EOF
sleep 30

# ---------------------------------------------------------------------------
# Step 7: Extract and parse audit log
# ---------------------------------------------------------------------------
step "Extracting audit log from KinD node"

AUDIT_LOG="/tmp/rhaii-audit-raw.json"
docker exec "${CLUSTER_NAME}-control-plane" cat /var/log/kubernetes/audit.log > "$AUDIT_LOG" 2>/dev/null || \
  podman exec "${CLUSTER_NAME}-control-plane" cat /var/log/kubernetes/audit.log > "$AUDIT_LOG"

LINES=$(wc -l < "$AUDIT_LOG")
info "Extracted $LINES audit log entries"

# Parse unique (verb, apiGroup, resource) tuples from the audit log.
# Filter to only the RHAII operator's service account.
AUDIT_TUPLES="/tmp/rhaii-audit-tuples.txt"
jq -r '
  select(.user.username == "system:serviceaccount:opendatahub-operator-system:opendatahub-operator-controller-manager")
  | select(.objectRef != null)
  | [.verb, (.objectRef.apiGroup // ""), .objectRef.resource]
  | @tsv
' "$AUDIT_LOG" \
  | LC_ALL=C sort -u > "$AUDIT_TUPLES"

TUPLE_COUNT=$(wc -l < "$AUDIT_TUPLES")
info "Found $TUPLE_COUNT unique (verb, apiGroup, resource) tuples used at runtime"

# Extract 403 Forbidden responses
FORBIDDEN="/tmp/rhaii-audit-forbidden.txt"
jq -r '
  select(.user.username == "system:serviceaccount:opendatahub-operator-system:opendatahub-operator-controller-manager")
  | select(.responseStatus.code == 403)
  | select(.objectRef != null)
  | [.verb, (.objectRef.apiGroup // ""), .objectRef.resource, .responseStatus.message]
  | @tsv
' "$AUDIT_LOG" \
  | LC_ALL=C sort -u > "$FORBIDDEN"

# ---------------------------------------------------------------------------
# Step 8: Parse the generated ClusterRole
# ---------------------------------------------------------------------------
step "Parsing generated RHAII ClusterRole"

ROLE_TUPLES="/tmp/rhaii-role-tuples.txt"
# Extract (verb, apiGroup, resource) tuples from the YAML role.
# Each rule has apiGroups[], resources[], verbs[] — produce the cartesian product.
python3 -c "
import yaml, sys

with open('$ROLE_FILE') as f:
    role = yaml.safe_load(f)

tuples = set()
for rule in role.get('rules', []):
    groups = rule.get('apiGroups', [''])
    resources = rule.get('resources', [])
    verbs = rule.get('verbs', [])
    for g in groups:
        for r in resources:
            for v in verbs:
                # Normalize: '' and '*' for core API group
                g_norm = g if g else ''
                tuples.add((v, g_norm, r))

for v, g, r in sorted(tuples):
    print(f'{v}\t{g}\t{r}')
" | LC_ALL=C sort -u > "$ROLE_TUPLES"

ROLE_COUNT=$(wc -l < "$ROLE_TUPLES")
info "ClusterRole contains $ROLE_COUNT unique (verb, apiGroup, resource) tuples"

# ---------------------------------------------------------------------------
# Step 9: Produce the diff report
# ---------------------------------------------------------------------------
step "Generating audit report"

{
  echo "======================================================================"
  echo "RHAII RBAC Audit Report"
  echo "Generated: $(date -Iseconds)"
  echo "Cluster: $CLUSTER_NAME"
  echo "Image: $IMG"
  echo "======================================================================"
  echo ""

  # --- Forbidden calls (under-grants) ---
  FORBIDDEN_COUNT=$(wc -l < "$FORBIDDEN")
  if [ "$FORBIDDEN_COUNT" -gt 0 ]; then
    echo "FORBIDDEN CALLS ($FORBIDDEN_COUNT) — missing permissions!"
    echo "----------------------------------------------------------------------"
    echo "VERB	APIGROUP	RESOURCE	MESSAGE"
    cat "$FORBIDDEN"
  else
    echo "FORBIDDEN CALLS: none (all API calls were authorized)"
  fi
  echo ""

  # --- Actually used permissions ---
  echo "PERMISSIONS USED AT RUNTIME ($TUPLE_COUNT tuples)"
  echo "----------------------------------------------------------------------"
  echo "VERB	APIGROUP	RESOURCE"
  cat "$AUDIT_TUPLES"
  echo ""

  # --- Over-grants: in role but never used ---
  OVER="/tmp/rhaii-over-grants.txt"
  # For each role tuple, check if any audit tuple matches.
  # Wildcard '*' in verbs/apiGroups needs special handling.
  LC_ALL=C comm -23 "$ROLE_TUPLES" "$AUDIT_TUPLES" > "$OVER"
  OVER_COUNT=$(wc -l < "$OVER")

  echo "OVER-GRANTS ($OVER_COUNT tuples in role but never used at runtime)"
  echo "----------------------------------------------------------------------"
  echo "(Not all are safe to remove — see rhaii-rbac-audit.md for guidance)"
  echo "VERB	APIGROUP	RESOURCE"
  cat "$OVER"
  echo ""

  # --- Under-grants: used at runtime but not in role ---
  UNDER="/tmp/rhaii-under-grants.txt"
  LC_ALL=C comm -13 "$ROLE_TUPLES" "$AUDIT_TUPLES" > "$UNDER"
  UNDER_COUNT=$(wc -l < "$UNDER")

  echo "POTENTIAL UNDER-GRANTS ($UNDER_COUNT tuples used but not explicitly in role)"
  echo "----------------------------------------------------------------------"
  echo "(Some may be covered by wildcard rules like '*' apiGroup)"
  echo "VERB	APIGROUP	RESOURCE"
  cat "$UNDER"
  echo ""

  echo "======================================================================"
  echo "Summary"
  echo "  Role tuples:      $ROLE_COUNT"
  echo "  Runtime tuples:   $TUPLE_COUNT"
  echo "  Forbidden calls:  $FORBIDDEN_COUNT"
  echo "  Over-grants:      $OVER_COUNT"
  echo "  Under-grants:     $UNDER_COUNT"
  echo "======================================================================"
} > "$REPORT"

cat "$REPORT"

echo ""
info "Full report saved to: $REPORT"
info "Raw audit log: $AUDIT_LOG"
info "To clean up: kind delete cluster --name $CLUSTER_NAME"

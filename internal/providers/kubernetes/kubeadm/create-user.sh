#!/bin/bash

set -e

# Configuration
USERNAME=${1:-testuser}
NAMESPACE=${2:-default}
K8S_CONFIG_DIR="$HOME/.kube"
CERTS_DIR="$K8S_CONFIG_DIR/certs"

# Colors for output (using same style as your install script)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create certificates directory
mkdir -p "$CERTS_DIR"

print_status "Creating user certificate for: $USERNAME"

# Generate private key for user
openssl genrsa -out "$CERTS_DIR/$USERNAME.key" 2048

# Create certificate signing request
openssl req -new -key "$CERTS_DIR/$USERNAME.key" -out "$CERTS_DIR/$USERNAME.csr" -subj "/CN=$USERNAME/O=developers"

print_status "Creating Kubernetes CSR resource..."

# Create Kubernetes CSR resource
cat > "$CERTS_DIR/$USERNAME-csr.yaml" <<EOF
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: $USERNAME
spec:
  request: $(cat "$CERTS_DIR/$USERNAME.csr" | base64 | tr -d '\n')
  signerName: kubernetes.io/kube-apiserver-client
  expirationSeconds: 86400  # 1 day
  usages:
  - client auth
EOF

# Apply the CSR
kubectl apply -f "$CERTS_DIR/$USERNAME-csr.yaml"

# Approve the CSR (requires admin privileges)
print_status "Approving certificate signing request..."
kubectl certificate approve "$USERNAME"

# Get the signed certificate
print_status "Retrieving signed certificate..."
kubectl get csr "$USERNAME" -o jsonpath='{.status.certificate}' | base64 -d > "$CERTS_DIR/$USERNAME.crt"

# Get cluster info
CLUSTER_NAME=$(kubectl config current-context)
SERVER_URL=$(kubectl config view -o jsonpath='{.clusters[0].cluster.server}')
CA_DATA=$(kubectl config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')

# Create kubeconfig for the user
print_status "Creating kubeconfig for $USERNAME..."

cat > "$CERTS_DIR/$USERNAME-kubeconfig" <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: $CA_DATA
    server: $SERVER_URL
  name: $CLUSTER_NAME
contexts:
- context:
    cluster: $CLUSTER_NAME
    user: $USERNAME
    namespace: $NAMESPACE
  name: $USERNAME@$CLUSTER_NAME
current-context: $USERNAME@$CLUSTER_NAME
users:
- name: $USERNAME
  user:
    client-certificate: $CERTS_DIR/$USERNAME.crt
    client-key: $CERTS_DIR/$USERNAME.key
EOF

print_status "User certificate and kubeconfig created successfully!"
print_status "Kubeconfig location: $CERTS_DIR/$USERNAME-kubeconfig"

# Test the configuration (will fail without RBAC)
print_warning "Testing user authentication (this will fail without RBAC permissions):"
kubectl --kubeconfig="$CERTS_DIR/$USERNAME-kubeconfig" auth whoami || true

# Cleanup CSR
kubectl delete csr "$USERNAME" || true

print_status "Next steps:"
echo "1. Create RBAC permissions for $USERNAME"
echo "2. Check with: kubectl --kubeconfig=$CERTS_DIR/$USERNAME-kubeconfig auth whoami"
echo "3. Test with: kubectl --kubeconfig=$CERTS_DIR/$USERNAME-kubeconfig get pods"
echo "4. Or set as default: export KUBECONFIG=$CERTS_DIR/$USERNAME-kubeconfig"

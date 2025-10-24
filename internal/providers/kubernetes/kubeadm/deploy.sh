#!/bin/bash
# deploy-local.sh

set -e

echo "Building Docker image..."
docker build -t thand-dev/agent:latest ../../../../

# Check if we're using Docker Desktop Kubernetes
CURRENT_CONTEXT=$(kubectl config current-context)

if [[ "$CURRENT_CONTEXT" == "docker-desktop" ]]; then
    echo "Detected Docker Desktop Kubernetes - image automatically available to cluster"
    # No need to load images, Docker Desktop uses local Docker daemon
elif [[ "$CURRENT_CONTEXT" == *"minikube"* ]]; then
    echo "Detected minikube - using minikube docker-env"
    eval $(minikube docker-env)
    docker build -t thand-dev/agent:latest .
elif [[ "$CURRENT_CONTEXT" == *"kind"* ]]; then
    echo "Detected kind - loading image into cluster"
    kind load docker-image thand-dev/agent:latest
else
    echo "Unknown Kubernetes context: $CURRENT_CONTEXT"
    exit 0
fi

# Apply the Kubernetes manifests (fix the path)
echo "Applying Kubernetes deployment..."
kubectl apply -f ./kubernetes-deployment.yaml

echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/thand-agent -n thand-system

echo "Deployment complete!"
echo "Checking pod status..."
kubectl get pods -n thand-system -l app=thand-agent

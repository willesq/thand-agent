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
    echo "Remote cluster detected - loading image into nodes..."
    
    # Save the image to a tar file
    echo "Saving image to tar file..."
    docker save thand-dev/agent:latest -o thand-agent.tar
    
    # Get list of all nodes (control plane + workers)
    NODES=$(kubectl get nodes -o jsonpath='{.items[*].metadata.name}')
    
    for node in $NODES; do
        echo "Loading image into node: $node"
        
        # Try to copy and load the image
        if scp thand-agent.tar $node:/tmp/thand-agent.tar 2>/dev/null; then
            ssh $node "sudo docker load -i /tmp/thand-agent.tar && rm -f /tmp/thand-agent.tar"
        else
            echo "Warning: Could not copy image to node $node"
        fi
    done
    
    # Clean up local tar file
    rm -f thand-agent.tar
fi

# Apply the Kubernetes manifests (fix the path)
echo "Applying Kubernetes deployment..."
kubectl apply -f ./kubernetes-deployment.yaml

echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/thand-agent -n thand-system

echo "Deployment complete!"
echo "Checking pod status..."
kubectl get pods -n thand-system -l app=thand-agent

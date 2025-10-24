#!/bin/bash
NAMESPACE="thand-system"

# Loop through, provider, roles and workflows configs and create secrets

for dir in providers roles workflows; do

    echo "Creating/updating secret for ${dir} configurations..."

    kubectl create secret generic "thand-${dir}-config" \
            --namespace="$NAMESPACE" \
            --from-file="all.yaml=../../../../config/${dir}/all.yaml" \
            --dry-run=client -o yaml | kubectl apply -f -

done

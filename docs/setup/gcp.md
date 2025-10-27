---
layout: default
title: Google Cloud Platform
parent: Setup Guides
nav_order: 1
description: "Deploy Thand Agent on Google Cloud Platform with IAM integration"
---

# Google Cloud Platform Setup
{: .no_toc }

Complete guide to deploying Thand Agent on Google Cloud Platform with IAM integration.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

This guide walks you through setting up Thand Agent on Google Cloud Platform (GCP), including:

- Creating the necessary GCP resources
- Configuring IAM roles and service accounts
- Deploying the Thand server on GCP
- Setting up agent authentication

---

## Prerequisites

Before starting, ensure you have:

- GCP project with billing enabled
- `gcloud` CLI installed and configured
- Project owner or IAM admin permissions
- Kubernetes cluster (if deploying on GKE)

---

## Step 1: Project Setup

### Enable Required APIs

```bash
gcloud services enable \
  compute.googleapis.com \
  iam.googleapis.com \
  cloudresourcemanager.googleapis.com \
  container.googleapis.com \
  cloudbuild.googleapis.com
```

### Set Project Variables

```bash
export PROJECT_ID="your-project-id"
export REGION="us-central1"
export ZONE="us-central1-a"

gcloud config set project $PROJECT_ID
```

---

## Step 2: Create Service Accounts

### Thand Server Service Account

```bash
# Create service account for Thand server
gcloud iam service-accounts create thand-server \
  --display-name="Thand Server" \
  --description="Service account for Thand server operations"

# Grant necessary permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:thand-server@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountTokenCreator"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:thand-server@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"
```

### Workload Identity (for GKE)

If deploying on GKE, configure Workload Identity:

```bash
# Create Kubernetes service account
kubectl create serviceaccount thand-server -n thand-system

# Bind to GCP service account
gcloud iam service-accounts add-iam-policy-binding \
  thand-server@$PROJECT_ID.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:$PROJECT_ID.svc.id.goog[thand-system/thand-server]"

# Annotate Kubernetes service account
kubectl annotate serviceaccount thand-server \
  -n thand-system \
  iam.gke.io/gcp-service-account=thand-server@$PROJECT_ID.iam.gserviceaccount.com
```

---

## Step 3: Network Configuration

### Create VPC and Firewall Rules

```bash
# Create VPC network
gcloud compute networks create thand-network \
  --subnet-mode=custom

# Create subnet
gcloud compute networks subnets create thand-subnet \
  --network=thand-network \
  --range=10.1.0.0/16 \
  --region=$REGION

# Allow internal traffic
gcloud compute firewall-rules create thand-internal \
  --network=thand-network \
  --allow=tcp,udp,icmp \
  --source-ranges=10.1.0.0/16

# Allow HTTPS traffic
gcloud compute firewall-rules create thand-https \
  --network=thand-network \
  --allow=tcp:443 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=thand-server
```

---

## Step 4: Deploy Thand Server

### Option A: GKE Deployment

Create a GKE cluster:

```bash
gcloud container clusters create thand-cluster \
  --zone=$ZONE \
  --network=thand-network \
  --subnetwork=thand-subnet \
  --enable-workload-identity \
  --num-nodes=3 \
  --machine-type=e2-standard-2
```

Deploy Thand server:

```yaml
# thand-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: thand-server
  namespace: thand-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: thand-server
  template:
    metadata:
      labels:
        app: thand-server
    spec:
      serviceAccountName: thand-server
      containers:
      - name: thand-server
        image: thand/agent:latest
        args: ["server", "start"]
        env:
        - name: THAND_SERVER_PORT
          value: "8080"
        - name: GOOGLE_APPLICATION_CREDENTIALS
          value: "/var/secrets/google/key.json"
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: thand-server
  namespace: thand-system
spec:
  selector:
    app: thand-server
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

Apply the deployment:

```bash
kubectl create namespace thand-system
kubectl apply -f thand-deployment.yaml
```

### Option B: Compute Engine Deployment

Create a VM instance:

```bash
gcloud compute instances create thand-server \
  --zone=$ZONE \
  --machine-type=e2-standard-2 \
  --network-interface=subnet=thand-subnet \
  --tags=thand-server \
  --image-family=ubuntu-2004-lts \
  --image-project=ubuntu-os-cloud \
  --service-account=thand-server@$PROJECT_ID.iam.gserviceaccount.com \
  --scopes=https://www.googleapis.com/auth/cloud-platform \
  --metadata=startup-script='#!/bin/bash
    apt-get update
    apt-get install -y docker.io
    systemctl start docker
    systemctl enable docker
    docker run -d \
      --name thand-server \
      --restart unless-stopped \
      -p 80:8080 \
      -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/keys/key.json \
      thand/agent:latest server start'
```

---

## Step 5: Configure DNS and TLS

### Cloud DNS Setup

```bash
# Create DNS zone
gcloud dns managed-zones create thand-zone \
  --dns-name="thand.example.com." \
  --description="Thand server DNS zone"

# Get external IP
EXTERNAL_IP=$(gcloud compute instances describe thand-server \
  --zone=$ZONE \
  --format="get(networkInterfaces[0].accessConfigs[0].natIP)")

# Create A record
gcloud dns record-sets transaction start \
  --zone=thand-zone

gcloud dns record-sets transaction add $EXTERNAL_IP \
  --name="thand.example.com." \
  --ttl=300 \
  --type=A \
  --zone=thand-zone

gcloud dns record-sets transaction execute \
  --zone=thand-zone
```

### SSL Certificate

Use Google-managed SSL certificates:

```bash
gcloud compute ssl-certificates create thand-ssl \
  --domains=thand.example.com \
  --global
```

---

## Step 6: IAM Roles Configuration

Create custom roles for different access levels:

```bash
# Create read-only role
gcloud iam roles create thandReadOnly \
  --project=$PROJECT_ID \
  --title="Thand Read Only" \
  --description="Read-only access for Thand users" \
  --permissions="compute.instances.list,compute.zones.list,storage.buckets.list"

# Create operator role  
gcloud iam roles create thandOperator \
  --project=$PROJECT_ID \
  --title="Thand Operator" \
  --description="Operator access for Thand users" \
  --permissions="compute.instances.start,compute.instances.stop,compute.instances.reset"

# Create admin role
gcloud iam roles create thandAdmin \
  --project=$PROJECT_ID \
  --title="Thand Admin" \
  --description="Administrative access for Thand users" \
  --permissions="compute.instances.*,iam.serviceAccounts.actAs"
```

---

## Step 7: Agent Configuration

Configure the agent to connect to your GCP-hosted server:

```yaml
# ~/.thand/config.yaml
server:
  url: "https://thand.example.com"
  
gcp:
  project_id: "your-project-id"
  region: "us-central1"
  
agent:
  listen_port: 8080
  
logging:
  level: "info"
```

---

## Step 8: Testing the Setup

### Test Server Connectivity

```bash
# Health check
curl https://thand.example.com/health

# API status  
curl https://thand.example.com/api/v1/status
```

### Test Agent Authentication

```bash
# Start agent
thand-agent start

# Authenticate
thand-agent auth login

# Request GCP access
thand-agent request gcp \
  --project $PROJECT_ID \
  --role roles/viewer \
  --duration 1h
```

---

## Troubleshooting

### Common Issues

1. **Permission Denied Errors**
   ```bash
   # Check service account permissions
   gcloud projects get-iam-policy $PROJECT_ID \
     --flatten="bindings[].members" \
     --format="table(bindings.role)" \
     --filter="bindings.members:thand-server@$PROJECT_ID.iam.gserviceaccount.com"
   ```

2. **Network Connectivity Issues**
   ```bash
   # Check firewall rules
   gcloud compute firewall-rules list --filter="network:thand-network"
   
   # Test connectivity
   gcloud compute instances describe thand-server --zone=$ZONE
   ```

3. **Certificate Issues**
   ```bash
   # Check SSL certificate status
   gcloud compute ssl-certificates describe thand-ssl --global
   ```

### Logs and Monitoring

```bash
# View server logs (GKE)
kubectl logs -l app=thand-server -n thand-system

# View server logs (Compute Engine)
gcloud compute instances get-serial-port-output thand-server --zone=$ZONE
```

---

## Next Steps

- **[AWS Setup](aws)** - Deploy on AWS
- **[Configuration](../configuration/)** - Advanced configuration options
- **[Workflows](../workflows/)** - Set up custom approval workflows
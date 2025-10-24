#!/bin/bash
kubectl -n thand-system port-forward svc/thand-agent 8081:8080

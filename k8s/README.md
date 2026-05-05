# Kubernetes Deployment Manifests

This directory contains all Kubernetes manifests for deploying CES-iBridge Orchestrator.

## Quick Deploy

```bash
# Apply all manifests in order
kubectl apply -f 01-namespace.yaml
kubectl apply -f 02-configmap.yaml
kubectl apply -f 03-secrets.yaml
kubectl apply -f 07-deployment.yaml
kubectl apply -f 08-service.yaml
kubectl apply -f 09-ingress.yaml
kubectl apply -f 10-hpa.yaml
```

## Files

- **01-namespace.yaml** - Namespace definition
- **02-configmap.yaml** - Application configuration
- **03-secrets.yaml** - Sensitive data (⚠️ Update before deploying!)
- **07-deployment.yaml** - Application deployment
- **08-service.yaml** - Service definition
- **09-ingress.yaml** - Ingress configuration
- **10-hpa.yaml** - Horizontal Pod Autoscaler

## Important

Before deploying:
1. Update `03-secrets.yaml` with real credentials
2. Update image in `07-deployment.yaml`
3. Update domain in `09-ingress.yaml`

See ../KUBERNETES_DEPLOYMENT.md for complete guide.

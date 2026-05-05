#!/bin/bash

# ============================================================================
# CES-iBridge Orchestrator Kubernetes Deployment Script
# ============================================================================
# This script deploys the orchestrator to Kubernetes cluster
#
# Usage:
#   ./deploy-k8s.sh [OPTIONS]
#
# Options:
#   --namespace NAME    Kubernetes namespace (default: ces-orchestrator)
#   --image IMAGE       Docker image to deploy
#   --replicas N        Number of replicas (default: 3)
#   --dry-run           Show what would be deployed without applying
#   --delete            Delete all resources
#   --help              Show this help message
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
NAMESPACE="ces-orchestrator"
IMAGE=""
REPLICAS=3
DRY_RUN=false
DELETE=false

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
CES-iBridge Orchestrator Kubernetes Deployment Script

Usage:
  ./deploy-k8s.sh [OPTIONS]

Options:
  --namespace NAME    Kubernetes namespace (default: ces-orchestrator)
  --image IMAGE       Docker image to deploy (required for deployment)
  --replicas N        Number of replicas (default: 3)
  --dry-run           Show what would be deployed without applying
  --delete            Delete all resources from cluster
  --help              Show this help message

Examples:
  # Deploy with default settings
  ./deploy-k8s.sh --image myregistry/orchestrator:v1.0.0

  # Deploy to custom namespace with 5 replicas
  ./deploy-k8s.sh --namespace production --image myregistry/orchestrator:v1.0.0 --replicas 5

  # Dry run to see what would be deployed
  ./deploy-k8s.sh --image myregistry/orchestrator:v1.0.0 --dry-run

  # Delete all resources
  ./deploy-k8s.sh --delete

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --image)
            IMAGE="$2"
            shift 2
            ;;
        --replicas)
            REPLICAS="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --delete)
            DELETE=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Delete resources if requested
if [ "$DELETE" = true ]; then
    log_warning "Deleting all orchestrator resources from cluster..."
    read -p "Are you sure? (yes/no): " -r
    if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        kubectl delete -f k8s/10-hpa.yaml --ignore-not-found=true
        kubectl delete -f k8s/09-ingress.yaml --ignore-not-found=true
        kubectl delete -f k8s/08-service.yaml --ignore-not-found=true
        kubectl delete -f k8s/07-deployment.yaml --ignore-not-found=true
        kubectl delete -f k8s/03-secrets.yaml --ignore-not-found=true
        kubectl delete -f k8s/02-configmap.yaml --ignore-not-found=true
        kubectl delete -f k8s/01-namespace.yaml --ignore-not-found=true
        log_success "Resources deleted"
    else
        log_info "Deletion cancelled"
    fi
    exit 0
fi

# Validate required parameters
if [ -z "$IMAGE" ]; then
    log_error "Docker image is required. Use --image flag"
    show_help
    exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl not found. Please install kubectl first."
    exit 1
fi

# Check if k8s directory exists
if [ ! -d "k8s" ]; then
    log_error "k8s directory not found. Are you in the project root?"
    exit 1
fi

log_info "========================================="
log_info "CES-iBridge Orchestrator Deployment"
log_info "========================================="
log_info "Namespace: $NAMESPACE"
log_info "Image: $IMAGE"
log_info "Replicas: $REPLICAS"
log_info "Dry Run: $DRY_RUN"
log_info "========================================="

# Update deployment with image and replicas
log_info "Updating deployment configuration..."
sed -i.bak "s|image:.*|image: $IMAGE|g" k8s/07-deployment.yaml
sed -i.bak "s|replicas:.*|replicas: $REPLICAS|g" k8s/07-deployment.yaml

# Dry run if requested
if [ "$DRY_RUN" = true ]; then
    log_info "DRY RUN MODE - No changes will be applied"
    log_info ""
    log_info "Would apply the following manifests:"
    echo "  - 01-namespace.yaml"
    echo "  - 02-configmap.yaml"
    echo "  - 03-secrets.yaml"
    echo "  - 07-deployment.yaml"
    echo "  - 08-service.yaml"
    echo "  - 09-ingress.yaml"
    echo "  - 10-hpa.yaml"
    log_info ""
    log_info "With configuration:"
    cat k8s/07-deployment.yaml | grep -A 2 "image:"
    exit 0
fi

# Warning about secrets
log_warning "⚠️  IMPORTANT: Make sure you've updated k8s/03-secrets.yaml with real credentials!"
read -p "Have you updated the secrets? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    log_error "Please update k8s/03-secrets.yaml before deploying"
    exit 1
fi

# Apply manifests
log_info "Creating namespace..."
kubectl apply -f k8s/01-namespace.yaml

log_info "Creating ConfigMap..."
kubectl apply -f k8s/02-configmap.yaml

log_info "Creating Secrets..."
kubectl apply -f k8s/03-secrets.yaml

log_info "Creating Deployment..."
kubectl apply -f k8s/07-deployment.yaml

log_info "Creating Service..."
kubectl apply -f k8s/08-service.yaml

log_info "Creating Ingress..."
kubectl apply -f k8s/09-ingress.yaml

log_info "Creating HPA..."
kubectl apply -f k8s/10-hpa.yaml

# Wait for deployment
log_info "Waiting for deployment to be ready..."
kubectl rollout status deployment/orchestrator -n $NAMESPACE --timeout=300s

# Show deployment status
log_success "Deployment completed successfully!"
log_info ""
log_info "========================================="
log_info "Deployment Status"
log_info "========================================="
kubectl get pods -n $NAMESPACE -l app=orchestrator
log_info ""
kubectl get svc -n $NAMESPACE
log_info ""
kubectl get ing -n $NAMESPACE
log_info ""
kubectl get hpa -n $NAMESPACE

# Get service URL
log_info ""
log_info "========================================="
log_info "Access Information"
log_info "========================================="
EXTERNAL_IP=$(kubectl get svc orchestrator-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
if [ "$EXTERNAL_IP" != "pending" ] && [ -n "$EXTERNAL_IP" ]; then
    log_success "Service URL: http://$EXTERNAL_IP"
else
    log_info "External IP is pending. Run this command to get it later:"
    echo "  kubectl get svc orchestrator-service -n $NAMESPACE"
fi

# Useful commands
log_info ""
log_info "========================================="
log_info "Useful Commands"
log_info "========================================="
echo "View logs:"
echo "  kubectl logs -f deployment/orchestrator -n $NAMESPACE"
echo ""
echo "View pods with k9s:"
echo "  k9s -n $NAMESPACE"
echo ""
echo "Port forward:"
echo "  kubectl port-forward service/orchestrator-service 8081:80 -n $NAMESPACE"
echo ""
echo "Scale deployment:"
echo "  kubectl scale deployment orchestrator --replicas=5 -n $NAMESPACE"
echo ""
echo "Update image:"
echo "  kubectl set image deployment/orchestrator orchestrator=newimage:tag -n $NAMESPACE"
echo ""
echo "Rollback deployment:"
echo "  kubectl rollout undo deployment/orchestrator -n $NAMESPACE"

log_success "Deployment complete! 🚀"

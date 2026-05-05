#!/bin/bash

# ============================================================================
# CES-iBridge Orchestrator Kubernetes Monitoring Script
# ============================================================================
# Quick helper script for monitoring and troubleshooting
# ============================================================================

set -e

NAMESPACE="ces-orchestrator"
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}========================================${NC}"
}

show_menu() {
    echo ""
    log_section "Orchestrator Monitoring Menu"
    echo "1) View all resources"
    echo "2) View pods"
    echo "3) View logs (live)"
    echo "4) View deployment status"
    echo "5) View services"
    echo "6) View ingress"
    echo "7) View HPA status"
    echo "8) View metrics"
    echo "9) Describe pod"
    echo "10) Shell into pod"
    echo "11) Port forward to local"
    echo "12) View events"
    echo "13) Launch k9s"
    echo "14) Health check"
    echo "0) Exit"
    echo ""
    read -p "Select option: " choice
    echo ""
}

view_all() {
    log_section "All Resources in $NAMESPACE"
    kubectl get all -n $NAMESPACE
}

view_pods() {
    log_section "Pods in $NAMESPACE"
    kubectl get pods -n $NAMESPACE -o wide
}

view_logs() {
    log_info "Available pods:"
    kubectl get pods -n $NAMESPACE --no-headers | awk '{print $1}'
    echo ""
    read -p "Enter pod name (or press Enter for deployment logs): " pod_name
    if [ -z "$pod_name" ]; then
        kubectl logs -f deployment/orchestrator -n $NAMESPACE --tail=100
    else
        kubectl logs -f $pod_name -n $NAMESPACE --tail=100
    fi
}

view_deployment() {
    log_section "Deployment Status"
    kubectl get deployment orchestrator -n $NAMESPACE
    echo ""
    kubectl rollout status deployment/orchestrator -n $NAMESPACE
}

view_services() {
    log_section "Services"
    kubectl get svc -n $NAMESPACE
}

view_ingress() {
    log_section "Ingress"
    kubectl get ing -n $NAMESPACE
}

view_hpa() {
    log_section "Horizontal Pod Autoscaler"
    kubectl get hpa -n $NAMESPACE
}

view_metrics() {
    log_section "Pod Metrics"
    kubectl top pods -n $NAMESPACE 2>/dev/null || log_info "Metrics server not available"
}

describe_pod() {
    log_info "Available pods:"
    kubectl get pods -n $NAMESPACE --no-headers | awk '{print $1}'
    echo ""
    read -p "Enter pod name: " pod_name
    if [ -n "$pod_name" ]; then
        kubectl describe pod $pod_name -n $NAMESPACE
    fi
}

shell_into_pod() {
    log_info "Available pods:"
    kubectl get pods -n $NAMESPACE --no-headers | awk '{print $1}'
    echo ""
    read -p "Enter pod name: " pod_name
    if [ -n "$pod_name" ]; then
        kubectl exec -it $pod_name -n $NAMESPACE -- /bin/sh
    fi
}

port_forward() {
    log_section "Port Forward"
    log_info "Forwarding service to localhost:8081..."
    kubectl port-forward service/orchestrator-service 8081:80 -n $NAMESPACE
}

view_events() {
    log_section "Recent Events"
    kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp' | tail -20
}

launch_k9s() {
    log_info "Launching k9s for namespace $NAMESPACE..."
    k9s -n $NAMESPACE
}

health_check() {
    log_section "Health Check"

    # Check pods
    log_info "Checking pods..."
    READY_PODS=$(kubectl get pods -n $NAMESPACE -l app=orchestrator --no-headers | grep "Running" | wc -l | tr -d ' ')
    TOTAL_PODS=$(kubectl get pods -n $NAMESPACE -l app=orchestrator --no-headers | wc -l | tr -d ' ')
    echo "✓ Pods: $READY_PODS/$TOTAL_PODS ready"

    # Check deployment
    log_info "Checking deployment..."
    DESIRED=$(kubectl get deployment orchestrator -n $NAMESPACE -o jsonpath='{.spec.replicas}')
    READY=$(kubectl get deployment orchestrator -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')
    echo "✓ Deployment: $READY/$DESIRED ready"

    # Check service
    log_info "Checking service..."
    SVC_EXISTS=$(kubectl get svc orchestrator-service -n $NAMESPACE -o name 2>/dev/null || echo "")
    if [ -n "$SVC_EXISTS" ]; then
        echo "✓ Service: orchestrator-service exists"
    else
        echo "✗ Service: not found"
    fi

    # Check ingress
    log_info "Checking ingress..."
    ING_EXISTS=$(kubectl get ing orchestrator-ingress -n $NAMESPACE -o name 2>/dev/null || echo "")
    if [ -n "$ING_EXISTS" ]; then
        echo "✓ Ingress: orchestrator-ingress exists"
    else
        echo "✗ Ingress: not found"
    fi

    # Check HPA
    log_info "Checking HPA..."
    HPA_EXISTS=$(kubectl get hpa orchestrator-hpa -n $NAMESPACE -o name 2>/dev/null || echo "")
    if [ -n "$HPA_EXISTS" ]; then
        echo "✓ HPA: orchestrator-hpa exists"
    else
        echo "✗ HPA: not found"
    fi

    echo ""
    log_info "Health check complete"
}

# Main loop
while true; do
    show_menu

    case $choice in
        1) view_all ;;
        2) view_pods ;;
        3) view_logs ;;
        4) view_deployment ;;
        5) view_services ;;
        6) view_ingress ;;
        7) view_hpa ;;
        8) view_metrics ;;
        9) describe_pod ;;
        10) shell_into_pod ;;
        11) port_forward ;;
        12) view_events ;;
        13) launch_k9s ;;
        14) health_check ;;
        0)
            log_info "Goodbye!"
            exit 0
            ;;
        *)
            echo "Invalid option"
            ;;
    esac

    echo ""
    read -p "Press Enter to continue..."
done

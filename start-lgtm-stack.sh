#!/bin/bash

# LGTM Stack Startup Script
# This script starts the complete LGTM observability stack

set -e

echo "🚀 Starting LGTM Stack (Loki + Grafana + Tempo + Mimir/Prometheus)"
echo "=================================================================="
echo ""

# Check if docker compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Error: docker-compose or docker compose not found"
    echo "Please install Docker Compose first"
    exit 1
fi

# Determine docker compose command
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "📋 Checking required configuration files..."
required_files=(
    "docker-compose.monitoring.yml"
    "tempo.yaml"
    "prometheus.yml"
    "grafana-datasources.yaml"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "❌ Error: Required file '$file' not found"
        exit 1
    fi
    echo "  ✓ $file"
done

echo ""
echo "🐳 Starting LGTM Stack containers..."
$DOCKER_COMPOSE -f docker-compose.monitoring.yml up -d

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 5

echo ""
echo "✅ LGTM Stack is running!"
echo ""
echo "📊 Access points:"
echo "  • Grafana:    http://localhost:3001 (admin/admin)"
echo "  • Prometheus: http://localhost:9090"
echo "  • Tempo:      http://localhost:3200"
echo "  • Loki:       http://localhost:3100"
echo ""
echo "📡 OTLP Endpoints for your application:"
echo "  • HTTP: http://localhost:4318"
echo "  • gRPC: http://localhost:4317"
echo ""
echo "🔍 To view logs:"
echo "  $DOCKER_COMPOSE -f docker-compose.monitoring.yml logs -f [service]"
echo ""
echo "🛑 To stop:"
echo "  $DOCKER_COMPOSE -f docker-compose.monitoring.yml down"
echo ""
echo "📖 For more information, see MIGRATION_JAEGER_TO_TEMPO.md"

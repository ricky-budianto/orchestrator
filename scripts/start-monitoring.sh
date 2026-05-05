#!/bin/bash

# Start script for monitoring stack
# Starts Prometheus, Grafana, Jaeger, and other observability tools

set -e  # Exit on error

echo "========================================="
echo "Starting Monitoring Stack"
echo "========================================="

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed."
    echo "Please install Docker to run the monitoring stack."
    exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed."
    echo "Please install Docker Compose to run the monitoring stack."
    exit 1
fi

# Check if monitoring compose file exists
if [ ! -f "docker-compose.monitoring.yml" ]; then
    echo "Error: docker-compose.monitoring.yml not found!"
    echo "Please ensure you're running this script from the project root."
    exit 1
fi

echo ""
echo "----------------------------------------"
echo "Checking Docker daemon..."
echo "----------------------------------------"

if docker info > /dev/null 2>&1; then
    echo "✓ Docker daemon is running"
else
    echo "✗ Docker daemon is not running"
    echo "Please start Docker and try again."
    exit 1
fi

echo ""
echo "----------------------------------------"
echo "Starting monitoring services..."
echo "----------------------------------------"

# Start the monitoring stack
echo "Running: docker-compose -f docker-compose.monitoring.yml up -d"
docker-compose -f docker-compose.monitoring.yml up -d

echo ""
echo "Waiting for services to become healthy..."
sleep 5

# Check service status
echo ""
echo "----------------------------------------"
echo "Service Status:"
echo "----------------------------------------"

docker-compose -f docker-compose.monitoring.yml ps

echo ""
echo "----------------------------------------"
echo "Access URLs:"
echo "----------------------------------------"
echo "📊 Grafana:      http://localhost:3001"
echo "   Username:     admin"
echo "   Password:     admin"
echo ""
echo "📈 Prometheus:   http://localhost:9090"
echo ""
echo "🔍 Jaeger UI:    http://localhost:16686"
echo ""
echo "========================================="
echo "Monitoring stack started successfully!"
echo "========================================="
echo ""
echo "Commands:"
echo "• View logs:    docker-compose -f docker-compose.monitoring.yml logs -f"
echo "• Stop stack:   docker-compose -f docker-compose.monitoring.yml down"
echo "• Restart:      docker-compose -f docker-compose.monitoring.yml restart"
echo ""
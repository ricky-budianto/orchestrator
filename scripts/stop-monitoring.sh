#!/bin/bash

# Stop script for monitoring stack

echo "========================================="
echo "Stopping Monitoring Stack"
echo "========================================="

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed."
    exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed."
    exit 1
fi

# Check if monitoring compose file exists
if [ ! -f "docker-compose.monitoring.yml" ]; then
    echo "Error: docker-compose.monitoring.yml not found!"
    echo "Please ensure you're running this script from the project root."
    exit 1
fi

echo "Stopping monitoring services..."
docker-compose -f docker-compose.monitoring.yml down

echo ""
echo "✓ Monitoring stack stopped successfully"
echo ""
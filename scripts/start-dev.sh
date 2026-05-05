#!/bin/bash

# Development start script with hot reload
# Uses go run with automatic rebuild on file changes

set -e  # Exit on error

echo "========================================="
echo "Starting CES Orchestrator (Development Mode)"
echo "========================================="

# Check if config.env exists
if [ ! -f "config.env" ]; then
    echo "Error: config.env not found!"
    echo "Please run './scripts/setup.sh' first or create config.env manually."
    exit 1
fi

# Load environment variables
echo "Loading environment configuration..."
set -a
source config.env
set +a

# Check if air is installed (for hot reload)
if command -v air &> /dev/null; then
    echo "✓ Air (hot reload) is installed"
    echo ""
    echo "Starting with hot reload..."
    echo "========================================="
    air
else
    echo "ℹ Air not found. Install it for hot reload:"
    echo "  go install github.com/cosmtrek/air@latest"
    echo ""
    echo "Starting without hot reload..."
    echo "========================================="
    go run main.go
fi
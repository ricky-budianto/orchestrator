#!/bin/bash

# Setup script for CES Orchestrator Service
# Downloads all dependencies and prepares the environment

set -e  # Exit on error

echo "========================================="
echo "CES Orchestrator Service Setup"
echo "========================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    echo "Visit: https://golang.org/doc/install"
    exit 1
fi

echo "✓ Go version: $(go version)"
echo ""

# Check if Docker is installed (optional, for monitoring stack)
if command -v docker &> /dev/null; then
    echo "✓ Docker version: $(docker --version)"
else
    echo "⚠ Docker not found. Docker is optional but required for monitoring stack."
fi

if command -v docker-compose &> /dev/null; then
    echo "✓ Docker Compose version: $(docker-compose --version)"
else
    echo "⚠ Docker Compose not found. Required for monitoring stack."
fi

echo ""
echo "----------------------------------------"
echo "Setting up environment configuration..."
echo "----------------------------------------"

# Create config.env if it doesn't exist
if [ ! -f "config.env" ]; then
    if [ -f "config.env.example" ]; then
        echo "Creating config.env from config.env.example..."
        cp config.env.example config.env
        echo "✓ config.env created. Please update it with your settings."
    elif [ -f "default.env" ]; then
        echo "Creating config.env from default.env..."
        cp default.env config.env
        echo "✓ config.env created. Please update it with your settings."
    else
        echo "⚠ Warning: No config.env.example or default.env found."
        echo "  Please create config.env manually."
    fi
else
    echo "✓ config.env already exists"
fi

echo ""
echo "----------------------------------------"
echo "Downloading Go dependencies..."
echo "----------------------------------------"

# Download Go modules
echo "Running: go mod download"
go mod download

echo "✓ Go dependencies downloaded"
echo ""

# Verify and tidy modules
echo "Running: go mod tidy"
go mod tidy

echo "✓ Go modules tidied"
echo ""

# Build the application to verify everything compiles
echo "----------------------------------------"
echo "Building application..."
echo "----------------------------------------"

echo "Running: go build -o main ."
go build -o main .

if [ -f "main" ]; then
    echo "✓ Application built successfully"
    rm main  # Remove the binary as this is just a test build
else
    echo "✗ Build failed. Please check for errors."
    exit 1
fi

echo ""
echo "----------------------------------------"
echo "Running tests..."
echo "----------------------------------------"

echo "Running: go test ./..."
if go test ./... -v -short; then
    echo "✓ All tests passed"
else
    echo "⚠ Some tests failed. Please review the output."
fi

echo ""
echo "========================================="
echo "Setup completed successfully!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Update config.env with your database, RabbitMQ, and Redis settings"
echo "2. Run './scripts/start.sh' to start the service"
echo "3. (Optional) Run './scripts/start-monitoring.sh' to start monitoring stack"
echo ""
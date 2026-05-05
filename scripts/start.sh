#!/bin/bash

# Start script for CES Orchestrator Service
# Handles environment setup and service startup

set -e  # Exit on error

echo "========================================="
echo "Starting CES Orchestrator Service"
echo "========================================="

# Check if config.env exists
if [ ! -f "config.env" ]; then
    echo "Error: config.env not found!"
    echo "Please run './scripts/setup.sh' first or create config.env manually."
    exit 1
fi

# Function to check service availability
check_service() {
    local service=$1
    local host=$2
    local port=$3
    
    echo -n "Checking $service at $host:$port... "
    
    if timeout 2 bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null; then
        echo "✓ Available"
        return 0
    else
        echo "✗ Not available"
        return 1
    fi
}

# Load environment variables
echo "Loading environment configuration..."
set -a
source config.env
set +a

echo ""
echo "----------------------------------------"
echo "Checking required services..."
echo "----------------------------------------"

# Parse database connection from environment
if [ ! -z "$POSTGRE_SQL_HOST" ] && [ ! -z "$POSTGRE_SQL_PORT" ]; then
    check_service "PostgreSQL" "$POSTGRE_SQL_HOST" "$POSTGRE_SQL_PORT" || {
        echo "⚠ Warning: PostgreSQL is not reachable. Service may fail to start."
        echo "  Please ensure PostgreSQL is running at $POSTGRE_SQL_HOST:$POSTGRE_SQL_PORT"
    }
fi

# Parse RabbitMQ connection (extract host and port from URL)
if [ ! -z "$RABBITMQURL" ]; then
    # Extract host and port from amqp://user:pass@host:port/ format
    RABBIT_HOST=$(echo "$RABBITMQURL" | sed -E 's|amqp://[^@]*@([^:]+):.*|\1|')
    RABBIT_PORT=$(echo "$RABBITMQURL" | sed -E 's|amqp://[^@]*@[^:]+:([0-9]+)/.*|\1|')
    
    if [ ! -z "$RABBIT_HOST" ] && [ ! -z "$RABBIT_PORT" ]; then
        check_service "RabbitMQ" "$RABBIT_HOST" "$RABBIT_PORT" || {
            echo "⚠ Warning: RabbitMQ is not reachable. RPC features may not work."
        }
    fi
fi

# Check Redis
if [ ! -z "$REDIS_HOST" ] && [ ! -z "$REDIS_PORT" ]; then
    check_service "Redis" "$REDIS_HOST" "$REDIS_PORT" || {
        echo "⚠ Warning: Redis is not reachable. Caching features may not work."
    }
fi

echo ""
echo "----------------------------------------"
echo "Starting service..."
echo "----------------------------------------"

# Build the application if binary doesn't exist
if [ ! -f "main" ]; then
    echo "Building application..."
    go build -o main .
    echo "✓ Build completed"
fi

# Get port from environment or use default
PORT=${PORT:-3000}
echo "Service will start on port: $PORT"
echo ""

# Function to handle graceful shutdown
cleanup() {
    echo ""
    echo "Shutting down service..."
    exit 0
}

trap cleanup INT TERM

# Start the service
echo "Starting CES Orchestrator Service..."
echo "Press Ctrl+C to stop"
echo "========================================="
echo ""

# Run the service
./main
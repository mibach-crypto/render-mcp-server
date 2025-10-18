#!/bin/bash

# Local development script for Render MCP Server with TypingMind

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "================================================"
echo "Render MCP Server - Local Development"
echo "================================================"
echo ""

# Check for .env file
if [ ! -f .env ]; then
    echo "⚠️  No .env file found!"
    echo "Creating from .env.example..."
    cp .env.example .env
    echo ""
    echo "Please edit .env file with your configuration:"
    echo "  1. Set AUTH_TOKEN (or run: ./scripts/generate_token.sh)"
    echo "  2. Set RENDER_API_KEY from https://dashboard.render.com/account/api-keys"
    echo ""
    exit 1
fi

# Source environment variables
set -a
source .env
set +a

# Check required variables
if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "your_secure_auth_token_here" ]; then
    echo "❌ AUTH_TOKEN not configured in .env"
    echo "Run: ./scripts/generate_token.sh to generate a secure token"
    exit 1
fi

if [ -z "$RENDER_API_KEY" ] || [ "$RENDER_API_KEY" = "your_render_api_key_here" ]; then
    echo "❌ RENDER_API_KEY not configured in .env"
    echo "Get your API key from: https://dashboard.render.com/account/api-keys"
    exit 1
fi

# Set default values if not provided
PORT=${PORT:-8080}
HOST=${HOST:-0.0.0.0}

echo "Configuration:"
echo "  Host: $HOST"
echo "  Port: $PORT"
echo "  Auth Token: [CONFIGURED]"
echo "  Render API Key: [CONFIGURED]"
if [ -n "$REDIS_URL" ]; then
    echo "  Redis: [CONFIGURED]"
else
    echo "  Redis: [NOT CONFIGURED - using in-memory storage]"
fi
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    echo "Please install Go from: https://golang.org/dl/"
    exit 1
fi

echo "Building server..."
go build -o render-mcp-server main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo ""
echo "================================================"
echo "Starting MCP Server..."
echo "================================================"
echo ""
echo "Server endpoints:"
echo "  MCP:    http://localhost:$PORT/mcp"
echo "  Health: http://localhost:$PORT/health"
echo ""
echo "TypingMind Configuration:"
echo "  Server URL: http://localhost:$PORT/mcp"
echo "  Auth Token: $AUTH_TOKEN"
echo ""
echo "Press Ctrl+C to stop the server"
echo "================================================"
echo ""

# Run the server
AUTH_TOKEN="$AUTH_TOKEN" \
RENDER_API_KEY="$RENDER_API_KEY" \
REDIS_URL="$REDIS_URL" \
PORT="$PORT" \
HOST="$HOST" \
./render-mcp-server --transport http
#!/bin/bash

# Quick setup script for Render MCP Server

set -e

echo "================================================"
echo "    Render MCP Server - Setup Assistant"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check prerequisites
echo "Checking prerequisites..."
echo ""

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_info "Go is installed: $GO_VERSION"
else
    print_error "Go is not installed"
    echo "   Please install Go from: https://golang.org/dl/"
    echo ""
fi

# Check Docker (optional)
if command -v docker &> /dev/null; then
    print_info "Docker is installed (optional)"
else
    print_warning "Docker not found (optional for containerized deployment)"
fi

# Check Git
if command -v git &> /dev/null; then
    print_info "Git is installed"
else
    print_error "Git is not installed"
fi

echo ""
echo "================================================"
echo "Setup Options:"
echo "================================================"
echo "1) Generate secure authentication token"
echo "2) Create environment configuration (.env)"
echo "3) Run locally for testing"
echo "4) Build Docker image"
echo "5) View deployment instructions"
echo "6) Exit"
echo ""
read -p "Select an option (1-6): " choice

case $choice in
    1)
        echo ""
        if [ -f "scripts/generate_token.sh" ]; then
            ./scripts/generate_token.sh
        else
            print_error "Token generation script not found"
        fi
        ;;
    
    2)
        echo ""
        if [ -f ".env" ]; then
            print_warning ".env file already exists"
            read -p "Overwrite? (y/N): " overwrite
            if [ "$overwrite" != "y" ] && [ "$overwrite" != "Y" ]; then
                echo "Keeping existing .env file"
                exit 0
            fi
        fi
        
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_info "Created .env file from template"
            echo ""
            echo "Next steps:"
            echo "1. Edit .env file with your configuration"
            echo "2. Set AUTH_TOKEN (run option 1 to generate)"
            echo "3. Add your RENDER_API_KEY"
            echo ""
            read -p "Open .env in editor? (y/N): " edit_env
            if [ "$edit_env" = "y" ] || [ "$edit_env" = "Y" ]; then
                ${EDITOR:-nano} .env
            fi
        else
            print_error ".env.example template not found"
        fi
        ;;
    
    3)
        echo ""
        if [ -f "scripts/run_local.sh" ]; then
            ./scripts/run_local.sh
        else
            print_error "Local run script not found"
        fi
        ;;
    
    4)
        echo ""
        echo "Building Docker image..."
        if command -v docker &> /dev/null; then
            docker build -t render-mcp-server:local .
            print_info "Docker image built successfully"
            echo ""
            echo "Run with: docker run -p 8080:10000 -e AUTH_TOKEN=your_token -e RENDER_API_KEY=your_key render-mcp-server:local"
        else
            print_error "Docker is not installed"
        fi
        ;;
    
    5)
        echo ""
        if [ -f "DEPLOY_TYPINGMIND.md" ]; then
            less DEPLOY_TYPINGMIND.md
        else
            echo "Opening deployment guide in browser..."
            echo "https://github.com/mibach-crypto/render-mcp-server/blob/main/DEPLOY_TYPINGMIND.md"
        fi
        ;;
    
    6)
        echo "Goodbye!"
        exit 0
        ;;
    
    *)
        print_error "Invalid option"
        ;;
esac

echo ""
echo "================================================"
echo "Need help? Check out:"
echo "- DEPLOY_TYPINGMIND.md for deployment guide"
echo "- README.md for general information"
echo "- GitHub Issues for support"
echo "================================================"
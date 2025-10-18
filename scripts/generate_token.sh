#!/bin/bash

# Script to generate a secure authentication token for TypingMind MCP connection

echo "================================================"
echo "Render MCP Server - Token Generator"
echo "================================================"
echo ""

# Function to generate token using different methods
generate_token() {
    if command -v openssl &> /dev/null; then
        # Use OpenSSL if available (preferred)
        openssl rand -hex 32
    elif command -v uuidgen &> /dev/null; then
        # Use uuidgen as fallback
        echo "$(uuidgen)$(uuidgen)" | tr -d '-' | tr '[:upper:]' '[:lower:]'
    else
        # Use /dev/urandom as last resort
        tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w 64 | head -n 1
    fi
}

# Generate the token
TOKEN=$(generate_token)

echo "Your new authentication token is:"
echo ""
echo "  $TOKEN"
echo ""
echo "================================================"
echo "IMPORTANT: Security Notes"
echo "================================================"
echo "1. Save this token securely - you won't see it again"
echo "2. Use this token in both:"
echo "   - Render environment variable: AUTH_TOKEN"
echo "   - TypingMind MCP configuration"
echo "3. Never commit this token to version control"
echo "4. Rotate tokens regularly for security"
echo ""
echo "================================================"
echo "Quick Setup Instructions:"
echo "================================================"
echo "1. In Render Dashboard:"
echo "   Add environment variable: AUTH_TOKEN=$TOKEN"
echo ""
echo "2. In TypingMind:"
echo "   Enter this token in the Authentication Token field"
echo ""
echo "================================================"
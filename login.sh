#!/bin/bash

# Your Login Details
EMAIL="vhamed02@gmail.com"
PASSWORD="trzKHkBl5SF84YF1yU6gRBaTb3UOI7Jdaav2793X7bk="

echo "================================================"
echo "  Logging in as $EMAIL"
echo "================================================"
echo ""

# Check if API is responding
echo "Checking API health..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ API is not responding. Please start services first:"
    echo "   ./start.sh"
    exit 1
fi

echo "✅ API is healthy"
echo ""
echo "Authenticating..."

# Login
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

# Check if login was successful
if echo "$RESPONSE" | grep -q "access_token"; then
    echo ""
    echo "================================================"
    echo "  ✅ Login Successful!"
    echo "================================================"
    echo ""
    
    # Extract tokens
    ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    ORG_NAME=$(echo "$RESPONSE" | grep -o '"name":"[^"]*"' | head -2 | tail -1 | cut -d'"' -f4)
    ROLE=$(echo "$RESPONSE" | grep -o '"role":"[^"]*"' | cut -d'"' -f4)
    
    echo "User Info:"
    echo "  Email: $EMAIL"
    echo "  User ID: $USER_ID"
    echo "  Organization: $ORG_NAME"
    echo "  Role: $ROLE"
    echo ""
    echo "Tokens:"
    echo "  Access Token: ${ACCESS_TOKEN:0:50}..."
    echo "  Refresh Token: ${REFRESH_TOKEN:0:50}..."
    echo ""
    echo "Export token for use:"
    echo "  export TOKEN=\"$ACCESS_TOKEN\""
    echo ""
    echo "Test authenticated endpoint:"
    echo "  curl http://localhost:8080/api/v1/auth/me -H \"Authorization: Bearer \$TOKEN\""
    echo ""
else
    echo ""
    echo "================================================"
    echo "  ❌ Login Failed"
    echo "================================================"
    echo ""
    echo "Response:"
    echo "$RESPONSE"
    echo ""
    
    # Check specific errors
    if echo "$RESPONSE" | grep -q "invalid email or password"; then
        echo "Invalid credentials. Please check your email and password."
    elif echo "$RESPONSE" | grep -q "suspended"; then
        echo "Account is suspended. Contact administrator."
    else
        echo "Unknown error. Check logs with: docker-compose logs control-plane"
    fi
    exit 1
fi

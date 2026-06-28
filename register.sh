#!/bin/bash

# Your Registration Details
EMAIL="vhamed02@gmail.com"
PASSWORD="trzKHkBl5SF84YF1yU6gRBaTb3UOI7Jdaav2793X7bk="
FIRST_NAME="Hamed"
LAST_NAME="Najari"
ORG_NAME="Hamed's Organization"
ORG_SLUG="hamed-org"

echo "================================================"
echo "  Registering Account for $EMAIL"
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
echo "Registering account..."

# Register
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"first_name\": \"$FIRST_NAME\",
    \"last_name\": \"$LAST_NAME\",
    \"org_name\": \"$ORG_NAME\",
    \"org_slug\": \"$ORG_SLUG\"
  }")

# Check if registration was successful
if echo "$RESPONSE" | grep -q "access_token"; then
    echo ""
    echo "================================================"
    echo "  ✅ Registration Successful!"
    echo "================================================"
    echo ""
    echo "Your Credentials:"
    echo "  Email: $EMAIL"
    echo "  Password: $PASSWORD"
    echo ""
    echo "Access Token:"
    echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4
    echo ""
    echo "Save your access token to use the API:"
    echo "  export TOKEN=\"\$(echo '$RESPONSE' | grep -o '\"access_token\":\"[^\"]*\"' | cut -d'\"' -f4)\""
    echo ""
    echo "Test with:"
    echo "  curl http://localhost:8080/api/v1/auth/me -H \"Authorization: Bearer \$TOKEN\""
    echo ""
else
    echo ""
    echo "================================================"
    echo "  ❌ Registration Failed"
    echo "================================================"
    echo ""
    echo "Response:"
    echo "$RESPONSE"
    echo ""
    
    # Check if already exists
    if echo "$RESPONSE" | grep -q "already"; then
        echo "Account already exists. Try logging in instead:"
        echo ""
        echo "curl -X POST http://localhost:8080/api/v1/auth/login \\"
        echo "  -H \"Content-Type: application/json\" \\"
        echo "  -d '{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}'"
    fi
fi

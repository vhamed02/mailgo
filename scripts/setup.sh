#!/bin/bash

# Mailbox Setup Script
# This script helps you get started with Mailbox quickly

set -e

echo "================================================"
echo "  Mailbox - Email Hosting SaaS Setup"
echo "================================================"
echo ""

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi
echo "✅ Docker is installed"

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi
echo "✅ Docker Compose is installed"

echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp .env.example .env
    
    # Generate JWT secret
    JWT_SECRET=$(openssl rand -hex 32)
    
    # Update .env with generated secret
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/change-this-to-a-secure-random-string-in-production/$JWT_SECRET/" .env
    else
        # Linux
        sed -i "s/change-this-to-a-secure-random-string-in-production/$JWT_SECRET/" .env
    fi
    
    echo "✅ .env file created with secure JWT secret"
    echo ""
    echo "⚠️  IMPORTANT: You need to configure the following in .env:"
    echo "   - MAILCOW_API_KEY (get from your Mailcow installation)"
    echo "   - BREVO_API_KEY (get from https://app.brevo.com)"
    echo ""
    read -p "Press Enter to continue after updating .env, or Ctrl+C to exit..."
else
    echo "✅ .env file already exists"
fi

echo ""
echo "Building Docker images..."
docker-compose build

echo ""
echo "Starting services..."
docker-compose up -d

echo ""
echo "Waiting for services to be healthy..."
sleep 10

# Check PostgreSQL
echo -n "Checking PostgreSQL... "
if docker-compose exec -T postgres pg_isready -U mailgo > /dev/null 2>&1; then
    echo "✅"
else
    echo "❌ PostgreSQL is not ready"
    exit 1
fi

# Check Redis
echo -n "Checking Redis... "
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo "✅"
else
    echo "❌ Redis is not ready"
    exit 1
fi

echo ""
echo "Running database migrations..."
# Note: This will fail initially because migrate binary doesn't exist in container
# This is documented in PRODUCTION_ASSESSMENT.md as a critical issue to fix
echo "⚠️  Migration step skipped - see PRODUCTION_ASSESSMENT.md for migration setup"
echo ""

echo "================================================"
echo "  ✅ Mailbox is running!"
echo "================================================"
echo ""
echo "Access the application:"
echo "  • Frontend:  http://localhost:3000"
echo "  • API:       http://localhost:8080"
echo "  • API Health: http://localhost:8080/health"
echo ""
echo "Default credentials: None yet - create an account via /auth/register"
echo ""
echo "Useful commands:"
echo "  make logs      - View logs"
echo "  make down      - Stop services"
echo "  make clean     - Remove all data (destructive)"
echo ""
echo "Next steps:"
echo "  1. Read PRODUCTION_ASSESSMENT.md for known issues"
echo "  2. Read DEPLOYMENT.md for production deployment"
echo "  3. Read ARCHITECTURE.md for technical deep dive"
echo ""
echo "Happy coding! 🚀"

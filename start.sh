#!/bin/bash

echo "================================================"
echo "  Starting MailGo System"
echo "================================================"
echo ""

# Stop any running containers
echo "Stopping any existing containers..."
docker-compose down

# Start services
echo "Starting services..."
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 15

# Check status
echo ""
echo "Checking service status..."
docker-compose ps

echo ""
echo "================================================"
echo "  Services Started!"
echo "================================================"
echo ""
echo "Access URLs:"
echo "  Frontend: http://localhost:3000"
echo "  API: http://localhost:8080"
echo "  Health: http://localhost:8080/health"
echo ""
echo "To register your account, run:"
echo "  ./register.sh"
echo ""
echo "To view logs:"
echo "  docker-compose logs -f"
echo ""

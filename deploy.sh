#!/bin/bash
set -e

echo "=== MailGo Production Deploy ==="

if [ ! -f ".env.production" ]; then
  echo "ERROR: .env.production not found. Copy .env.production.example and fill in values."
  exit 1
fi

export $(grep -v '^#' .env.production | xargs)

echo "1. Pulling latest code..."
git pull origin main

echo "2. Building images..."
docker compose -f docker-compose.prod.yml build --no-cache

echo "3. Starting services..."
docker compose -f docker-compose.prod.yml up -d

echo "4. Waiting for API health..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:8082/health > /dev/null 2>&1; then
    echo "API is healthy."
    break
  fi
  echo "Waiting... ($i/30)"
  sleep 3
done

echo ""
echo "=== Deploy complete ==="
echo "Frontend : https://mail.tracix.net"
echo "API      : https://mail.tracix.net/api/v1"
echo "Health   : http://localhost:8082/health"

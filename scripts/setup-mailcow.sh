#!/bin/bash
set -e

MAILCOW_DIR="$(dirname "$0")/../mailcow-dockerized"

if [ -d "$MAILCOW_DIR" ]; then
  echo "Mailcow already cloned at $MAILCOW_DIR"
else
  echo "Cloning Mailcow..."
  git clone https://github.com/mailcow/mailcow-dockerized "$MAILCOW_DIR"
fi

cd "$MAILCOW_DIR"

if [ ! -f "mailcow.conf" ]; then
  echo "Generating Mailcow config..."
  MAILCOW_HOSTNAME="${MAILCOW_HOSTNAME:-mail.localhost}" \
  MAILCOW_TZ="${TZ:-UTC}" \
  ./generate_config.sh

  # Override HTTP port to 8081 to avoid conflict with other services
  sed -i.bak 's/^HTTP_PORT=.*/HTTP_PORT=8081/' mailcow.conf
  sed -i.bak 's/^HTTP_BIND=.*/HTTP_BIND=0.0.0.0/' mailcow.conf
  sed -i.bak 's/^HTTPS_PORT=.*/HTTPS_PORT=8443/' mailcow.conf
fi

echo "Starting Mailcow..."
docker compose pull --quiet
docker compose up -d

echo ""
echo "Mailcow is starting up (takes ~2 min on first run)"
echo "Webmail: http://localhost:8081"
echo "Admin:   http://localhost:8081/admin  (admin / moohoo)"

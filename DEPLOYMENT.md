# Mailbox Deployment Guide

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- Make (optional, for convenience commands)
- 4GB RAM minimum
- 20GB disk space minimum

## Quick Start (Local Development)

```bash
# 1. Clone repository
git clone <repo-url>
cd mailgo

# 2. Copy environment file
cp .env.example .env

# 3. Update critical environment variables
# Edit .env and set:
# - JWT_SECRET (generate with: openssl rand -hex 32)
# - MAILCOW_API_KEY (get from Mailcow admin panel)
# - BREVO_API_KEY (get from Brevo dashboard)

# 4. Build and start services
make build
make up

# 5. Run database migrations
make migrate

# 6. Access the application
# Frontend: http://localhost:3000
# API: http://localhost:8080
# API Health: http://localhost:8080/health
```

## Production Deployment

### Step 1: Infrastructure Setup

#### Option A: Docker Compose (Small Scale)

```bash
# 1. Provision server (Ubuntu 22.04 recommended)
# Minimum: 4 vCPU, 8GB RAM, 50GB SSD

# 2. Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# 3. Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 4. Clone repository
git clone <repo-url>
cd mailgo

# 5. Configure production environment
cp .env.example .env
nano .env
```

#### Option B: Kubernetes (Large Scale)

```bash
# See kubernetes/ directory for manifests
kubectl apply -f kubernetes/
```

### Step 2: Configure Environment Variables

**Critical Security Settings:**

```bash
# Generate secure JWT secret
JWT_SECRET=$(openssl rand -hex 32)

# Use strong database password
DATABASE_URL=postgresql://mailgo:$(openssl rand -hex 16)@postgres:5432/mailgo?sslmode=require

# Enable SSL mode for production
DATABASE_URL=${DATABASE_URL}?sslmode=require
```

**External Service Configuration:**

```bash
# Mailcow (requires separate installation)
# See: https://github.com/mailcow/mailcow-dockerized
MAILCOW_API_URL=https://mail.yourdomain.com/api/v1
MAILCOW_API_KEY=<your-api-key>

# Brevo
# Get API key from: https://app.brevo.com/settings/keys/api
BREVO_API_KEY=<your-api-key>
BREVO_FROM_EMAIL=noreply@yourdomain.com
```

### Step 3: Setup Mailcow (Infrastructure Component)

Mailcow requires separate installation:

```bash
# 1. Clone Mailcow
cd /opt
git clone https://github.com/mailcow/mailcow-dockerized
cd mailcow-dockerized

# 2. Generate configuration
./generate_config.sh

# 3. Edit mailcow.conf
nano mailcow.conf
# Set: MAILCOW_HOSTNAME=mail.yourdomain.com

# 4. Start Mailcow
docker-compose up -d

# 5. Access admin panel
# https://mail.yourdomain.com
# Default: admin / moohoo

# 6. Generate API key
# Settings > Access > API > Generate Key
```

### Step 4: Database Setup

```bash
# Option 1: Use Docker Compose (development)
docker-compose up -d postgres

# Option 2: Use managed database (production)
# - AWS RDS PostgreSQL
# - Google Cloud SQL
# - DigitalOcean Managed Databases

# Set DATABASE_URL in .env
DATABASE_URL=postgresql://user:pass@host:5432/dbname?sslmode=require
```

### Step 5: SSL/TLS Setup

#### Option A: Nginx Reverse Proxy

```nginx
# /etc/nginx/sites-available/mailgo
server {
    listen 443 ssl http2;
    server_name app.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### Option B: Traefik

```yaml
# docker-compose.override.yml
services:
  traefik:
    image: traefik:v2.10
    command:
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@yourdomain.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./letsencrypt:/letsencrypt

  frontend:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`app.yourdomain.com`)"
      - "traefik.http.routers.frontend.entrypoints=websecure"
      - "traefik.http.routers.frontend.tls.certresolver=letsencrypt"

  control-plane:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.api.entrypoints=websecure"
      - "traefik.http.routers.api.tls.certresolver=letsencrypt"
```

### Step 6: Run Database Migrations

```bash
# Using Make
make migrate

# Or manually
docker-compose exec control-plane sh -c "migrate -path /app/migrations -database \$DATABASE_URL up"

# Verify migrations
docker-compose exec postgres psql -U mailgo -d mailgo -c "\dt"
```

### Step 7: Start Services

```bash
# Production deployment
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Verify services
docker-compose ps
docker-compose logs -f

# Check health
curl http://localhost:8080/health
```

### Step 8: Monitoring Setup

#### Prometheus

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'mailgo-api'
    static_configs:
      - targets: ['control-plane:9090']
```

#### Grafana

```bash
docker run -d \
  -p 3001:3000 \
  --name=grafana \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
  grafana/grafana
```

### Step 9: Backup Configuration

```bash
# Database backup script
#!/bin/bash
# /opt/mailgo/backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"

# Backup database
docker-compose exec -T postgres pg_dump -U mailgo mailgo | gzip > $BACKUP_DIR/mailgo_$DATE.sql.gz

# Backup environment
cp .env $BACKUP_DIR/env_$DATE

# Retain last 7 days
find $BACKUP_DIR -name "mailgo_*.sql.gz" -mtime +7 -delete

# Add to crontab
# 0 2 * * * /opt/mailgo/backup.sh
```

## Scaling Strategy

### Horizontal Scaling

```yaml
# docker-compose.scale.yml
services:
  control-plane:
    deploy:
      replicas: 3
    
  worker:
    deploy:
      replicas: 5
```

```bash
docker-compose -f docker-compose.yml -f docker-compose.scale.yml up -d
```

### Database Read Replicas

```yaml
services:
  postgres-replica:
    image: postgres:16-alpine
    environment:
      POSTGRES_MASTER_HOST: postgres
    command: ["postgres", "-c", "hot_standby=on"]
```

### Redis Cluster

```yaml
services:
  redis-master:
    image: redis:7-alpine
  
  redis-replica-1:
    image: redis:7-alpine
    command: redis-server --slaveof redis-master 6379
  
  redis-replica-2:
    image: redis:7-alpine
    command: redis-server --slaveof redis-master 6379
```

## Health Checks

```bash
# API Health
curl http://localhost:8080/health

# Database Connection
docker-compose exec postgres pg_isready -U mailgo

# Redis Connection
docker-compose exec redis redis-cli ping

# Worker Status
docker-compose exec worker ps aux | grep worker
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs control-plane
docker-compose logs worker

# Check resources
docker stats

# Restart service
docker-compose restart control-plane
```

### Database Connection Issues

```bash
# Test connection
docker-compose exec control-plane sh -c 'psql $DATABASE_URL -c "SELECT 1"'

# Check connection pool
docker-compose logs control-plane | grep "connection pool"
```

### Mailbox Provisioning Failures

```bash
# Check worker logs
docker-compose logs worker | grep "mailbox:provision"

# Check Mailcow logs
docker-compose -f /opt/mailcow-dockerized/docker-compose.yml logs mailcow-dockerized

# Manual retry
docker-compose exec worker sh -c 'asynq inspect pending:mailbox:provision'
```

## Security Checklist

- [ ] Change all default passwords
- [ ] Generate strong JWT secret
- [ ] Enable SSL/TLS
- [ ] Configure firewall (only 80, 443 open)
- [ ] Enable database SSL
- [ ] Set up fail2ban
- [ ] Configure rate limiting
- [ ] Enable audit logging
- [ ] Set up monitoring alerts
- [ ] Configure automated backups
- [ ] Test backup restoration
- [ ] Review .env for secrets
- [ ] Use secrets management (Vault)
- [ ] Enable 2FA for admin accounts

## Performance Tuning

### Database

```sql
-- Increase connection pool
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '2GB';
ALTER SYSTEM SET effective_cache_size = '6GB';
```

### Redis

```conf
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

### Application

```env
# Increase workers
ASYNQ_CONCURRENCY=20

# Increase database connections
DATABASE_MAX_CONNECTIONS=50

# Enable caching
CACHE_ENABLED=true
```

## Rollback Procedure

```bash
# 1. Stop services
docker-compose down

# 2. Restore database backup
gunzip < /backups/mailgo_YYYYMMDD_HHMMSS.sql.gz | \
  docker-compose exec -T postgres psql -U mailgo mailgo

# 3. Restore previous environment
cp /backups/env_YYYYMMDD_HHMMSS .env

# 4. Start services
docker-compose up -d
```

## Monitoring Metrics

Key metrics to track:

- API response time (p50, p95, p99)
- Database connection pool usage
- Redis memory usage
- Worker queue depth
- Failed job count
- Mailbox provisioning success rate
- Storage usage per tenant
- Request rate per tenant

## Support

For issues:
1. Check logs: `make logs`
2. Review health: `make health`
3. Check PRODUCTION_ASSESSMENT.md for known issues
4. Open GitHub issue with logs attached

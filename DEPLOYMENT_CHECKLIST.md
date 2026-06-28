# Production Deployment Checklist

## Pre-Deployment

### Code Quality
- [ ] All P0 issues resolved (see PRODUCTION_ASSESSMENT.md)
- [ ] Transaction management implemented
- [ ] Structured logging implemented (zerolog)
- [ ] Error handling reviewed
- [ ] Input validation verified
- [ ] Code reviewed by team

### Testing
- [ ] Unit tests written and passing
- [ ] Integration tests passing
- [ ] E2E tests passing
- [ ] Load testing completed
- [ ] Security testing completed
- [ ] Manual QA completed

### Security
- [ ] JWT secrets changed from defaults
- [ ] Database passwords strong and unique
- [ ] API keys rotated
- [ ] CORS configured for production domains only
- [ ] Rate limiting configured
- [ ] HTTPS enforced (via reverse proxy)
- [ ] Security headers configured
- [ ] SQL injection testing passed
- [ ] XSS protection verified

### Configuration
- [ ] Environment variables set for production
- [ ] Database connection pool tuned
- [ ] Redis configured for persistence
- [ ] Log levels set appropriately (info/warn)
- [ ] Metrics endpoint configured
- [ ] Health check intervals tuned

### Infrastructure
- [ ] Database backups automated
- [ ] Redis persistence configured (AOF + RDB)
- [ ] Monitoring setup (Prometheus/Grafana)
- [ ] Log aggregation configured
- [ ] Alerting rules defined
- [ ] Resource limits set (CPU/memory)
- [ ] SSL certificates obtained
- [ ] DNS configured

### Documentation
- [ ] API documentation complete (Swagger)
- [ ] Deployment guide written
- [ ] Runbook created (common issues)
- [ ] Team trained on system
- [ ] Emergency contacts documented

---

## Deployment Steps

### 1. Database Setup
```bash
# Create production database
createdb -h prod-db mailgo_prod

# Run migrations
docker-compose exec postgres psql -U mailgo -d mailgo_prod < migrations/000001_initial_schema.up.sql

# Verify schema
psql -h prod-db -U mailgo -d mailgo_prod -c "\dt"
```

### 2. Configuration
```bash
# Copy production environment file
cp .env.production .env

# Update secrets
vi .env
# - JWT_SECRET (generate: openssl rand -base64 64)
# - DATABASE_URL (production database)
# - MAILCOW_API_KEY
# - BREVO_API_KEY
```

### 3. Build & Deploy
```bash
# Build production images
docker-compose -f docker-compose.prod.yml build

# Push to registry (if using)
docker-compose -f docker-compose.prod.yml push

# Deploy
docker-compose -f docker-compose.prod.yml up -d

# Verify health
curl https://api.yourdomain.com/health
```

### 4. Smoke Tests
```bash
# Test registration
curl -X POST https://api.yourdomain.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com",...}'

# Test login
curl -X POST https://api.yourdomain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"..."}'

# Test protected endpoint
curl https://api.yourdomain.com/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

---

## Post-Deployment

### Monitoring
- [ ] Metrics dashboard configured
- [ ] Alerts triggered for test scenarios
- [ ] Log aggregation verified
- [ ] Health checks passing
- [ ] Performance baseline established

### Validation
- [ ] End-to-end user flows tested
- [ ] Performance acceptable
- [ ] No errors in logs
- [ ] Database queries optimized
- [ ] Memory usage stable

### Communication
- [ ] Team notified of deployment
- [ ] Users informed (if applicable)
- [ ] Support team trained
- [ ] Runbook shared

---

## Rollback Plan

### If Deployment Fails

1. **Immediately:**
   ```bash
   # Revert to previous version
   docker-compose -f docker-compose.prod.yml down
   docker tag mailgo-api:previous mailgo-api:latest
   docker-compose -f docker-compose.prod.yml up -d
   ```

2. **If Database Migration Failed:**
   ```bash
   # Rollback migration
   psql -h prod-db -U mailgo -d mailgo_prod < migrations/000001_initial_schema.down.sql
   
   # Restore from backup
   pg_restore -h prod-db -U mailgo -d mailgo_prod /backups/latest.dump
   ```

3. **Communicate:**
   - Notify team
   - Update status page
   - Document what went wrong

---

## Production Environment Variables

```bash
# Application
APP_ENV=production
APP_PORT=8080
APP_WORKERS=4
LOG_LEVEL=info

# Database
DATABASE_URL=postgresql://mailgo:STRONG_PASSWORD@prod-db:5432/mailgo_prod?sslmode=require
DATABASE_MAX_CONNECTIONS=100
DATABASE_MAX_IDLE=25
DATABASE_MAX_LIFETIME=30m

# Redis
REDIS_URL=redis://prod-redis:6379/0
REDIS_PASSWORD=STRONG_PASSWORD

# JWT (Generate: openssl rand -base64 64)
JWT_SECRET=YOUR_PRODUCTION_SECRET_HERE
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
JWT_ISSUER=mailgo-production

# External Services
MAILCOW_API_URL=https://mail.yourdomain.com/api/v1
MAILCOW_API_KEY=YOUR_MAILCOW_KEY

BREVO_API_KEY=YOUR_BREVO_KEY
BREVO_FROM_EMAIL=noreply@yourdomain.com

# Security
BCRYPT_COST=12
PASSWORD_MIN_LENGTH=12
SESSION_TIMEOUT=24h

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_ENABLED=true

# CORS
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9090
TRACING_ENABLED=true
TRACING_ENDPOINT=http://jaeger:14268/api/traces
```

---

## Resource Requirements

### Minimum (Small deployment)
- **Control Plane**: 1 CPU, 2GB RAM
- **Worker**: 1 CPU, 1GB RAM
- **PostgreSQL**: 2 CPU, 4GB RAM, 50GB storage
- **Redis**: 1 CPU, 2GB RAM, 10GB storage
- **Frontend**: 1 CPU, 1GB RAM

### Recommended (Production)
- **Control Plane**: 2 CPU, 4GB RAM (2 replicas)
- **Worker**: 2 CPU, 2GB RAM (2 replicas)
- **PostgreSQL**: 4 CPU, 8GB RAM, 200GB SSD
- **Redis**: 2 CPU, 4GB RAM, 50GB SSD
- **Frontend**: 2 CPU, 2GB RAM (2 replicas)

---

## Backup Strategy

### Database Backups
```bash
# Daily full backup
0 2 * * * pg_dump -h prod-db -U mailgo mailgo_prod | gzip > /backups/mailgo-$(date +\%Y\%m\%d).sql.gz

# Retention: 7 daily, 4 weekly, 12 monthly
```

### Redis Backups
```bash
# AOF + RDB enabled in redis.conf
appendonly yes
save 900 1
save 300 10
save 60 10000
```

### Disaster Recovery
- [ ] Backups tested monthly
- [ ] Recovery procedure documented
- [ ] Recovery time objective (RTO): < 4 hours
- [ ] Recovery point objective (RPO): < 1 hour

---

## Monitoring & Alerts

### Critical Alerts
- API health check fails
- Database connection pool exhausted
- Redis connection fails
- Error rate > 5%
- Response time > 2s (p99)
- CPU usage > 90%
- Memory usage > 90%
- Disk usage > 85%

### Warning Alerts
- Error rate > 1%
- Response time > 1s (p99)
- CPU usage > 70%
- Memory usage > 75%
- Queue depth > 1000

### Metrics to Track
- Request rate (per endpoint)
- Response time (p50, p95, p99)
- Error rate (per endpoint)
- Active users
- Database connections
- Redis memory usage
- Queue depth
- Worker processing rate

---

## Performance Targets

### API Performance
- **Health endpoint**: < 10ms
- **Auth endpoints**: < 200ms (p99)
- **CRUD endpoints**: < 500ms (p99)
- **Throughput**: > 1000 req/s

### Database
- **Query time**: < 100ms (p99)
- **Connection pool**: 80% utilization max
- **Replication lag**: < 1s

### System
- **CPU**: < 70% average
- **Memory**: < 75% average
- **Disk I/O**: < 80% capacity

---

## Security Hardening

### Network
- [ ] Firewall configured (only necessary ports open)
- [ ] VPC/private network for internal services
- [ ] DDoS protection enabled
- [ ] WAF configured

### Application
- [ ] Security headers set (HSTS, CSP, X-Frame-Options)
- [ ] Rate limiting active
- [ ] Input sanitization verified
- [ ] Dependencies scanned for vulnerabilities

### Database
- [ ] SSL/TLS connections enforced
- [ ] Minimal user permissions
- [ ] Query logging enabled
- [ ] Audit logging enabled

### Secrets Management
- [ ] Secrets not in code
- [ ] Environment variables encrypted at rest
- [ ] Secret rotation policy defined
- [ ] Access logs monitored

---

## Compliance (if applicable)

- [ ] GDPR compliance verified
- [ ] Data retention policy implemented
- [ ] Privacy policy published
- [ ] Terms of service published
- [ ] Cookie consent (if applicable)
- [ ] Right to deletion implemented
- [ ] Data export functionality

---

## Sign-off

- [ ] Engineering lead approval
- [ ] Security team approval
- [ ] Operations team approval
- [ ] Product owner approval
- [ ] All checklist items completed

**Deployment Date**: _______________  
**Deployed By**: _______________  
**Version**: _______________  
**Rollback Plan Verified**: _______________  

---

**Good luck with your deployment! 🚀**

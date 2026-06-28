# Getting Started with MailGo

## 🚀 Quick Start (5 minutes)

### Prerequisites
- Docker & Docker Compose installed
- Ports 8080, 3000, 5432, 6379 available

### Step 1: Start the System
```bash
git clone <your-repo>
cd mailgo
./start.sh
```

This will:
- Start PostgreSQL, Redis, API, Worker, and Frontend
- Wait for all health checks to pass
- Take ~30 seconds

### Step 2: Run Database Migrations
```bash
./run-migrations.sh
```

This creates all 8 database tables.

### Step 3: Register Your First Account
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@yourcompany.com",
    "password": "YourSecurePassword123!",
    "first_name": "Admin",
    "last_name": "User",
    "org_name": "Your Company",
    "org_slug": "your-company"
  }'
```

Save the `access_token` from the response!

### Step 4: Test Authentication
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@yourcompany.com",
    "password": "YourSecurePassword123!"
  }'

# Use the token
export TOKEN="your_access_token_here"
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📖 Understanding the System

### What You Just Built

You now have a running SaaS control plane with:

1. **Authentication System**
   - User registration with email/password
   - JWT-based authentication
   - Access & refresh tokens
   - Protected API routes

2. **Multi-Tenancy**
   - Organization-based isolation
   - Role-based access control (owner/admin/user)
   - Per-organization quotas

3. **Database Layer**
   - 8 tables with proper relationships
   - Audit logging for all actions
   - Automatic timestamp management

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────┐
│           Frontend (Next.js)                │
│              Port 3000                      │
└──────────────────┬──────────────────────────┘
                   │ REST API
┌──────────────────▼──────────────────────────┐
│      Control Plane API (Go)                 │
│              Port 8080                      │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │   Application Services              │   │
│  │   (Business Logic)                  │   │
│  └───────┬─────────────────────────────┘   │
│          │                                  │
│  ┌───────▼─────────────────────────────┐   │
│  │   Infrastructure Adapters           │   │
│  │   (Database, Queue, Mail)           │   │
│  └─────────────────────────────────────┘   │
└──────────────────┬──────────────────────────┘
                   │
    ┌──────────────┴───────────────┐
    │                              │
┌───▼──────────┐        ┌──────────▼────────┐
│ PostgreSQL   │        │   Redis           │
│ (Data)       │        │   (Queue/Cache)   │
└──────────────┘        └───────────────────┘
```

**Hexagonal Architecture Benefits:**
- Business logic independent of infrastructure
- Easy to test (no database needed for unit tests)
- Flexible (swap Mailcow for AWS SES easily)
- Clear separation of concerns

---

## 📁 Project Structure

```
mailgo/
├── backend/                    # Go API & Worker
│   ├── cmd/
│   │   ├── api/               # API entrypoint
│   │   └── worker/            # Worker entrypoint
│   ├── internal/
│   │   ├── domain/            # Business entities & rules
│   │   ├── application/       # Use cases & services
│   │   ├── api/rest/          # HTTP handlers
│   │   └── infrastructure/    # External integrations
│   └── migrations/            # Database migrations
│
├── frontend/                   # Next.js UI
│   ├── app/                   # Pages & routing
│   ├── components/            # React components
│   └── lib/                   # Utilities
│
├── docker/                    # Docker configurations
│   ├── Dockerfile.api
│   ├── Dockerfile.worker
│   └── Dockerfile.frontend
│
├── docker-compose.yml         # Service orchestration
├── .env                       # Environment variables
│
└── scripts/                   # Helper scripts
    ├── start.sh               # Start everything
    ├── login.sh               # Test login
    ├── register.sh            # Test registration
    └── run-migrations.sh      # Run DB migrations
```

---

## 🗄️ Database Schema

### Core Tables

**users** - User accounts
- id, email, password_hash, first_name, last_name
- email_verified, status, created_at, updated_at

**organizations** - Tenants/customers
- id, name, slug, status
- created_at, updated_at, deleted_at

**organization_users** - Many-to-many with roles
- id, organization_id, user_id, role (owner/admin/user)
- joined_at

**domains** - Email domains
- id, organization_id, name, status
- dns_verified, spf_record, dkim_record, dmarc_record

**mailboxes** - Email accounts
- id, organization_id, domain_id
- email, local_part, display_name
- quota_bytes, used_bytes, password_hash

**aliases** - Email forwarding
- id, organization_id, domain_id
- source, destination, active

**quotas** - Resource limits
- id, organization_id
- max_domains, max_mailboxes, max_storage_gb, max_aliases

**audit_logs** - Complete audit trail
- id, organization_id, user_id
- action, entity_type, entity_id, details
- ip_address, user_agent, created_at

---

## 🔐 Authentication Flow

### Registration
1. Client sends email, password, name, org details
2. API validates input (email format, password strength)
3. Password hashed with bcrypt (cost 10)
4. User, organization, relationship, and quota created
5. JWT tokens generated and returned

### Login
1. Client sends email and password
2. API finds user by email
3. Bcrypt compares password with stored hash
4. API retrieves user's organizations and roles
5. JWT tokens generated with user context

### Protected Requests
1. Client sends `Authorization: Bearer <token>`
2. Middleware validates JWT signature
3. Middleware extracts user_id, organization_id, role
4. Middleware attaches to request context
5. Handler accesses user context

---

## 🧪 Testing

### Manual API Testing

**Health Check:**
```bash
curl http://localhost:8080/health
```

**Register:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "first_name": "Test",
    "last_name": "User",
    "org_name": "Test Org",
    "org_slug": "test-org"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

**Get Current User (Protected):**
```bash
export TOKEN="your_token_here"
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

### Using Helper Scripts

```bash
# Test registration with your credentials
./register.sh

# Test login
./login.sh

# Both scripts will display results and provide export commands
```

---

## 🐳 Docker Services

### Control Plane API (Port 8080)
- **Image**: golang:alpine
- **Build**: Multi-stage (builder + runtime)
- **Health**: GET /health every 30s
- **Depends**: PostgreSQL, Redis

### Worker (Background Tasks)
- **Image**: golang:alpine
- **Purpose**: Process async provisioning tasks
- **Depends**: PostgreSQL, Redis

### Frontend (Port 3000)
- **Image**: node:20-alpine
- **Framework**: Next.js 14
- **Depends**: Control Plane API

### PostgreSQL (Port 5432)
- **Image**: postgres:16-alpine
- **Volume**: postgres_data
- **Health**: pg_isready check

### Redis (Port 6379)
- **Image**: redis:7-alpine
- **Volume**: redis_data
- **Mode**: AOF persistence

---

## 🔧 Configuration

### Environment Variables (.env)

**Application:**
```bash
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=info
```

**Database:**
```bash
DATABASE_URL=postgresql://mailgo:mailgo_password@postgres:5432/mailgo?sslmode=disable
DATABASE_MAX_CONNECTIONS=25
```

**Authentication:**
```bash
JWT_SECRET=your-secret-key-here
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
```

**External Services:**
```bash
MAILCOW_API_URL=http://mailcow:8080/api/v1
MAILCOW_API_KEY=your-mailcow-key

BREVO_API_KEY=your-brevo-key
BREVO_FROM_EMAIL=noreply@yourdomain.com
```

---

## 📊 Monitoring

### Health Checks
```bash
# API health
curl http://localhost:8080/health

# Check all services
docker-compose ps
```

### Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f control-plane
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Database Access
```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U mailgo -d mailgo

# Common queries
SELECT COUNT(*) FROM users;
SELECT * FROM organizations;
SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;
```

---

## 🚧 What's Next?

### Immediate Tasks
1. **Domain Management** - Add domain CRUD endpoints
2. **Mailbox Management** - Add mailbox CRUD endpoints
3. **Worker Tasks** - Implement Mailcow provisioning
4. **Frontend** - Build login/dashboard UI

### Short Term
1. Transaction management (wrap Register in DB transaction)
2. Structured logging (replace println with zerolog)
3. Unit tests for services
4. Integration tests for API

### Long Term
1. Prometheus metrics
2. Distributed tracing
3. API documentation (Swagger)
4. Admin dashboard

---

## 📚 Additional Resources

- **ARCHITECTURE.md** - Detailed system design
- **PRODUCTION_ASSESSMENT.md** - Known issues & risks
- **FINAL_STATUS.md** - Current implementation status
- **DEPLOYMENT.md** - Production deployment guide

---

## 🆘 Troubleshooting

### Port Already in Use
```bash
# Find what's using port 8080
lsof -i :8080

# Stop the container
docker stop <container_name>
```

### Database Connection Failed
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Verify connection string in .env
```

### Migrations Failed
```bash
# Check if already run
docker-compose exec postgres psql -U mailgo -d mailgo -c "\dt"

# If tables exist, migrations already applied
```

### API Not Responding
```bash
# Restart the service
docker-compose restart control-plane

# Check logs
docker-compose logs control-plane

# Verify health
curl http://localhost:8080/health
```

---

## 💡 Tips

1. **Use helper scripts** - They handle the curl complexity
2. **Check logs first** - Most issues are visible in logs
3. **Verify health checks** - Use docker-compose ps
4. **Database state** - Connect with psql to inspect data
5. **Token expiry** - Access tokens expire in 15 minutes

---

**Ready to build?** Start with the domain management endpoints. The service layer is already implemented, you just need to wire up the HTTP handlers!

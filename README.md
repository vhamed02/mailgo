# Mailbox - Production-Grade Multi-Tenant Email Hosting SaaS

## Architecture Overview

Mailbox is a production-ready, multi-tenant email hosting platform built with strict architectural boundaries following hexagonal architecture principles.

### Core Architectural Layers

```
┌─────────────────────────────────────────────────────────────┐
│                     PRESENTATION LAYER                       │
│                  (Next.js Frontend - SaaS UI)               │
└────────────────────────┬────────────────────────────────────┘
                         │ REST API (/api/v1)
┌────────────────────────▼────────────────────────────────────┐
│                      CONTROL PLANE                           │
│              (Go Backend - Business Logic)                   │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Domain Layer (Entities, Value Objects, Interfaces)  │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  Application Layer (Use Cases, Services)             │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  API Layer (REST Handlers + gRPC Services)           │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────┬────────────────────────────┬───────────────────┘
             │                            │
             │ Interfaces/Ports           │ gRPC
             │                            │
┌────────────▼────────────┐  ┌───────────▼────────────────────┐
│  INFRASTRUCTURE LAYER   │  │      ASYNC LAYER               │
│                         │  │   (Asynq Workers)              │
│  ┌──────────────────┐  │  │                                │
│  │ Mailcow Adapter  │  │  │  - Provisioning Jobs           │
│  │ (Inbound Email)  │  │  │  - DNS Sync                    │
│  ├──────────────────┤  │  │  - Mailbox Lifecycle           │
│  │ Brevo Adapter    │  │  │  - Audit Log Processing        │
│  │ (Outbound Email) │  │  │                                │
│  ├──────────────────┤  │  └────────────────────────────────┘
│  │ PostgreSQL Repo  │  │
│  ├──────────────────┤  │
│  │ Redis Cache      │  │
│  └──────────────────┘  │
└─────────────────────────┘
```

### Key Design Principles

1. **Separation of Concerns**: Control plane is completely independent of email infrastructure
2. **Ports & Adapters**: All external systems accessed via interfaces
3. **Infrastructure Agnostic**: Can swap Mailcow/Brevo without changing business logic
4. **Multi-Tenancy**: Strict tenant isolation at database and application level
5. **API-First**: REST for external, gRPC for internal communication

## Technology Stack

- **Backend**: Go 1.21+ (clean architecture, hexagonal design)
- **Frontend**: Next.js 14+ (App Router, TypeScript)
- **Internal API**: gRPC with Protocol Buffers
- **Public API**: REST (versioned /api/v1)
- **Database**: PostgreSQL 16 (multi-tenant schema)
- **Cache**: Redis 7
- **Queue**: Asynq (Redis-backed job queue)
- **Auth**: JWT-based with RBAC
- **Email Infrastructure**: 
  - Mailcow (inbound, internal only)
  - Brevo (outbound relay)

## Project Structure

```
mailgo/
├── backend/                    # Go backend services
│   ├── cmd/
│   │   ├── api/               # REST API server
│   │   └── worker/            # Asynq worker
│   ├── internal/
│   │   ├── domain/            # Domain entities & interfaces
│   │   ├── application/       # Use cases & business logic
│   │   ├── infrastructure/    # Adapters (DB, Mailcow, Brevo)
│   │   ├── api/               # REST & gRPC handlers
│   │   └── worker/            # Async job handlers
│   ├── pkg/                   # Shared packages
│   ├── proto/                 # gRPC definitions
│   └── migrations/            # Database migrations
├── frontend/                   # Next.js application
│   ├── app/                   # App Router pages
│   ├── components/            # React components
│   ├── lib/                   # API clients & utilities
│   └── types/                 # TypeScript types
├── docker/                     # Dockerfiles
├── docker-compose.yml         # Full stack orchestration
└── docs/                      # Documentation
```

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional, for convenience)

### Running the System

```bash
# 1. Clone the repository
git clone <repo-url>
cd mailgo

# 2. Configure environment
cp .env.example .env
# Edit .env with your configuration

# 3. Start all services
docker-compose up -d

# 4. Run database migrations
docker-compose exec control-plane /app/migrate

# 5. Access the application
# Frontend: http://localhost:3000
# API: http://localhost:8080
# Mailcow (internal): http://localhost:8081
```

### Development Mode

```bash
# Start infrastructure only
docker-compose up -d postgres redis mailcow

# Run backend locally
cd backend
go run cmd/api/main.go

# Run frontend locally
cd frontend
npm run dev

# Run worker locally
cd backend
go run cmd/worker/main.go
```

## API Documentation

### REST API Endpoints

All REST endpoints are versioned under `/api/v1`:

#### Authentication
- `POST /api/v1/auth/register` - Register new organization
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token

#### Organizations
- `GET /api/v1/organizations` - List user's organizations
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations/:id` - Get organization details
- `PATCH /api/v1/organizations/:id` - Update organization

#### Domains
- `GET /api/v1/domains` - List domains
- `POST /api/v1/domains` - Add domain
- `GET /api/v1/domains/:id` - Get domain details
- `DELETE /api/v1/domains/:id` - Remove domain
- `POST /api/v1/domains/:id/verify` - Verify DNS records

#### Mailboxes
- `GET /api/v1/mailboxes` - List mailboxes
- `POST /api/v1/mailboxes` - Create mailbox
- `GET /api/v1/mailboxes/:id` - Get mailbox details
- `PATCH /api/v1/mailboxes/:id` - Update mailbox
- `DELETE /api/v1/mailboxes/:id` - Delete mailbox
- `POST /api/v1/mailboxes/:id/suspend` - Suspend mailbox
- `POST /api/v1/mailboxes/:id/unsuspend` - Unsuspend mailbox

#### Aliases
- `GET /api/v1/aliases` - List aliases
- `POST /api/v1/aliases` - Create alias
- `DELETE /api/v1/aliases/:id` - Delete alias

### gRPC Services (Internal Only)

- `ProvisioningService` - Mailbox provisioning operations
- `TenantService` - Tenant management
- `AuditService` - Audit log operations

## Security

### Authentication & Authorization

- JWT-based authentication with access + refresh tokens
- Role-Based Access Control (RBAC):
  - **Owner**: Full organization control
  - **Admin**: Manage domains, mailboxes, users
  - **User**: Manage own mailbox only

### Security Features

- Rate limiting per tenant and IP
- Audit logging for all sensitive operations
- Secure secret management (environment + vault-ready)
- Password hashing with bcrypt
- SQL injection prevention (parameterized queries)
- CORS configuration
- Request validation middleware

## Multi-Tenancy

### Tenant Isolation

- **Database Level**: All tables have `organization_id` with enforced constraints
- **Application Level**: All queries filtered by tenant context
- **API Level**: JWT contains tenant context, enforced in middleware
- **Resource Quotas**: Per-tenant limits on domains, mailboxes, storage

### Tenant Context Flow

```
Request → JWT Validation → Extract Tenant ID → Inject Context → Handler → Repository (filtered)
```

## Email Infrastructure Integration

### Mailcow Integration (Inbound)

- **Purpose**: Internal mail server for receiving emails
- **Access**: Only via adapter interface, never exposed to frontend
- **Operations**: 
  - Mailbox provisioning
  - Alias management
  - Quota enforcement
- **Abstraction**: `MailServerAdapter` interface in domain layer

### Brevo Integration (Outbound)

- **Purpose**: SMTP/API relay for sending emails
- **Access**: Only via adapter interface
- **Operations**:
  - Transactional emails
  - Welcome emails
  - Password resets
- **Abstraction**: `EmailSenderAdapter` interface in domain layer

### Adapter Swapping

Both adapters implement domain interfaces and can be swapped:

```go
// Domain interface
type MailServerAdapter interface {
    CreateMailbox(ctx context.Context, req CreateMailboxRequest) error
    DeleteMailbox(ctx context.Context, email string) error
    // ...
}

// Implementations
type MailcowAdapter struct { /* ... */ }
type AlternativeAdapter struct { /* ... */ }

// Dependency injection
var adapter MailServerAdapter = &MailcowAdapter{}
```

## Observability

### Logging

- Structured JSON logs
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID tracing
- Tenant context in logs

### Metrics

- Prometheus-compatible endpoint: `/metrics`
- Key metrics:
  - Request rates and latencies
  - Queue job processing
  - Database connection pool
  - Cache hit rates

### Tracing

- OpenTelemetry-ready architecture
- Distributed tracing across services
- gRPC and HTTP instrumentation

## Database Schema

### Core Tables

- `organizations` - Tenant/organization data
- `users` - User accounts
- `organization_users` - Many-to-many with roles
- `domains` - Email domains per organization
- `mailboxes` - Email accounts
- `aliases` - Email aliases and forwarding
- `audit_logs` - Comprehensive audit trail
- `quotas` - Per-tenant resource limits

### Migrations

Located in `backend/migrations/`, using golang-migrate:

```bash
# Apply migrations
migrate -path backend/migrations -database "postgres://..." up

# Rollback
migrate -path backend/migrations -database "postgres://..." down 1
```

## Async Job Processing

### Asynq Workers

Job types:
- `mailbox:provision` - Create mailbox on Mailcow
- `mailbox:delete` - Remove mailbox from Mailcow
- `domain:verify` - Check DNS records
- `email:send` - Send email via Brevo
- `audit:process` - Process audit log entries

### Queue Configuration

- Default queue: general operations
- Critical queue: time-sensitive operations (priority)
- Low priority queue: bulk operations

## Docker Network Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      public network                      │
│                                                          │
│  ┌──────────┐         ┌──────────┐                     │
│  │ frontend │────────▶│   api    │                     │
│  │  :3000   │         │  :8080   │                     │
│  └──────────┘         └─────┬────┘                     │
│                              │                           │
└──────────────────────────────┼───────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────┐
│                    internal network                       │
│                                                           │
│  ┌─────────┐  ┌────────┐  ┌────────┐  ┌─────────────┐ │
│  │ worker  │  │postgres│  │ redis  │  │   mailcow   │ │
│  │         │  │        │  │        │  │  (stack)    │ │
│  └─────────┘  └────────┘  └────────┘  └─────────────┘ │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

## Production Readiness Checklist

- [x] Horizontal scalability (stateless services)
- [x] Database connection pooling
- [x] Graceful shutdown handling
- [x] Health check endpoints
- [x] Rate limiting
- [x] Audit logging
- [x] Error tracking ready
- [x] Metrics collection
- [x] Structured logging
- [x] Secret management abstraction
- [x] Database migration system
- [x] Multi-tenant isolation
- [x] RBAC implementation
- [x] API versioning
- [x] Request validation
- [x] Full Docker orchestration

## Environment Variables

See `.env.example` for full configuration. Key variables:

```bash
# Application
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info

# Database
DATABASE_URL=postgresql://user:pass@postgres:5432/mailgo?sslmode=disable
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE=5

# Redis
REDIS_URL=redis://redis:6379/0

# JWT
JWT_SECRET=<generate-secure-secret>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Mailcow
MAILCOW_API_URL=http://mailcow-dockerized-mailcow-1:8080/api/v1
MAILCOW_API_KEY=<mailcow-api-key>

# Brevo
BREVO_API_KEY=<brevo-api-key>
BREVO_API_URL=https://api.brevo.com/v3

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

## Contributing

See `CONTRIBUTING.md` for development guidelines.

## License

Proprietary - All rights reserved

# MailGo Architecture Deep Dive

## System Overview

MailGo is a multi-tenant email hosting SaaS platform built with strict architectural boundaries following hexagonal (ports & adapters) architecture principles.

## Core Design Principles

### 1. Separation of Concerns

```
┌─────────────────────────────────────────────────────────────────┐
│                        PRESENTATION LAYER                        │
│                    (Next.js - User Interface)                    │
│  • No business logic                                             │
│  • Only UI state management                                      │
│  • Communicates via REST API only                                │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         │ HTTP REST /api/v1
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                        CONTROL PLANE                             │
│                   (Go Backend - Core Product)                    │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                      API LAYER                            │  │
│  │  • REST handlers (Echo framework)                         │  │
│  │  • Request validation                                      │  │
│  │  • Authentication middleware (JWT)                         │  │
│  │  • Authorization (RBAC)                                    │  │
│  │  • Rate limiting                                           │  │
│  └──────────────────────┬───────────────────────────────────┘  │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐  │
│  │                  APPLICATION LAYER                         │  │
│  │  • Use cases (business logic)                              │  │
│  │  • Service orchestration                                   │  │
│  │  • Transaction management                                  │  │
│  │  • Domain validation                                       │  │
│  └──────────────────────┬───────────────────────────────────┘  │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐  │
│  │                    DOMAIN LAYER                            │  │
│  │  • Entities (User, Org, Domain, Mailbox)                   │  │
│  │  • Value Objects                                           │  │
│  │  • Domain Interfaces (Ports)                               │  │
│  │  • Business rules                                          │  │
│  │  • NO infrastructure dependencies                          │  │
│  └────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
                         │
                         │ Interfaces/Ports
                         │
        ┌────────────────┼────────────────┐
        │                │                │
┌───────▼──────┐  ┌──────▼─────┐  ┌──────▼──────┐
│ INFRASTRUCTURE│  │   DATA     │  │    ASYNC    │
│    ADAPTERS   │  │   LAYER    │  │    LAYER    │
│               │  │            │  │             │
│ • Mailcow     │  │ PostgreSQL │  │ Asynq       │
│ • Brevo       │  │ Repositories│  │ Workers     │
│ • Redis Cache │  │            │  │ Jobs        │
└───────────────┘  └────────────┘  └─────────────┘
```

### 2. Dependency Inversion

**CORRECT (Hexagonal Architecture):**
```
Domain Layer (Interfaces)
    ↑
    │ implements
    │
Infrastructure Layer (Adapters)
```

**WRONG (Traditional Layering):**
```
Business Logic
    ↓ depends on
Infrastructure
```

### 3. Infrastructure Agnosticism

The business logic never knows about Mailcow or Brevo:

```go
// ❌ WRONG - Business logic depends on infrastructure
type MailboxService struct {
    mailcowClient *mailcow.Client  // Direct dependency!
}

func (s *MailboxService) CreateMailbox() {
    s.mailcowClient.CreateMailbox()  // Tightly coupled
}

// ✅ CORRECT - Business logic depends on interface
type MailboxService struct {
    mailServer domain.MailServerAdapter  // Interface dependency
}

func (s *MailboxService) CreateMailbox() {
    s.mailServer.CreateMailbox()  // Can swap implementation
}
```

## Component Architecture

### API Server (Control Plane)

```
Request Flow:
1. HTTP Request → Echo Router
2. Rate Limiter Middleware → Check rate limits
3. Auth Middleware → Validate JWT, extract claims
4. RBAC Middleware → Check role permissions
5. Handler → Parse request, validate input
6. Service → Execute business logic
7. Repository → Persist to database
8. Adapter → Call external system (async)
9. Handler → Format response
10. HTTP Response
```

### Worker (Async Processing)

```
Job Flow:
1. API enqueues job → Redis (Asynq)
2. Worker polls queue → Gets job
3. Handler executes → Calls adapter
4. Adapter calls Mailcow/Brevo → External system
5. Success → Remove from queue
6. Failure → Retry with exponential backoff
7. Max retries exceeded → Move to DLQ
```

## Data Flow Patterns

### Synchronous Operations (Read)

```
User → Frontend → API → Service → Repository → PostgreSQL
                                          ↓
                                      Cache Check
                                          ↓
                                        Redis
```

### Asynchronous Operations (Write)

```
User → Frontend → API → Service → Repository → PostgreSQL
                            ↓
                        Queue Job
                            ↓
                          Redis
                            ↓
                         Worker
                            ↓
                     Infrastructure
                       (Mailcow)
```

## Multi-Tenancy Architecture

### Database-Level Isolation

Every table has `organization_id`:

```sql
CREATE TABLE mailboxes (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,  -- Tenant isolation
    email VARCHAR(255) NOT NULL,
    ...
    CONSTRAINT fk_organization 
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id)
);

-- Index for fast tenant queries
CREATE INDEX idx_mailboxes_org ON mailboxes(organization_id);
```

### Application-Level Enforcement

```go
// ❌ WRONG - No tenant filtering
func (r *MailboxRepository) List(ctx context.Context) ([]*Mailbox, error) {
    return r.db.Query("SELECT * FROM mailboxes")
}

// ✅ CORRECT - Always filter by tenant
func (r *MailboxRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Mailbox, error) {
    return r.db.Query("SELECT * FROM mailboxes WHERE organization_id = $1", orgID)
}
```

### JWT-Based Tenant Context

```go
// JWT claims include tenant context
type Claims struct {
    UserID         uuid.UUID   `json:"user_id"`
    OrganizationID uuid.UUID   `json:"organization_id"`  // Tenant ID
    Role           domain.Role `json:"role"`
}

// Middleware injects into request context
func AuthMiddleware(next Handler) Handler {
    return func(c Context) error {
        claims := validateJWT(c.Request())
        c.Set("organization_id", claims.OrganizationID)
        return next(c)
    }
}

// All handlers use tenant context
func (h *MailboxHandler) List(c echo.Context) error {
    orgID := c.Get("organization_id").(uuid.UUID)
    mailboxes, err := h.service.ListMailboxes(c.Request().Context(), orgID)
    // ...
}
```

## Security Architecture

### Authentication Flow

```
1. User submits email + password
2. API validates credentials
3. Checks user status (active/suspended)
4. Checks organization status
5. Generates JWT with claims:
   - user_id
   - organization_id
   - role
   - expiry (15 minutes)
6. Generates refresh token (7 days)
7. Returns both tokens
```

### Authorization (RBAC)

```go
// Role hierarchy
Owner > Admin > User

// Permission matrix
Action              Owner  Admin  User
─────────────────────────────────────
Manage billing        ✓      ✗     ✗
Add/remove domains    ✓      ✓     ✗
Create mailboxes      ✓      ✓     ✗
Manage own mailbox    ✓      ✓     ✓
View audit logs       ✓      ✓     ✗
Invite users          ✓      ✓     ✗
```

### Audit Logging

All sensitive operations logged:

```go
type AuditLog struct {
    OrganizationID uuid.UUID
    UserID         *uuid.UUID
    Action         string  // "create", "update", "delete"
    EntityType     string  // "mailbox", "domain", "user"
    EntityID       *uuid.UUID
    Details        map[string]interface{}
    IPAddress      string
    UserAgent      string
    CreatedAt      time.Time
}
```

## Async Job Architecture

### Job Types

```go
const (
    TypeMailboxProvision   = "mailbox:provision"   // Priority: default
    TypeMailboxDelete      = "mailbox:delete"      // Priority: default
    TypeDomainVerification = "domain:verify"       // Priority: low
    TypeEmailSend          = "email:send"          // Priority: critical
    TypeAuditLogProcess    = "audit:process"       // Priority: low
)
```

### Queue Priority

```yaml
Queues:
  critical: 60%  # Email sending, password resets
  default:  30%  # Mailbox operations
  low:      10%  # Background tasks, analytics
```

### Retry Strategy

```go
Retry Policy:
- Max retries: 5
- Initial delay: 1s
- Max delay: 1 hour
- Backoff: exponential with jitter
- DLQ: After max retries exceeded

Example timeline:
Attempt 1: immediate
Attempt 2: +1s
Attempt 3: +2s
Attempt 4: +4s
Attempt 5: +8s
Attempt 6: +16s (fail, move to DLQ)
```

## Database Schema Design

### Normalized Schema

```
organizations (tenant)
    ↓
    ├─→ organization_users (roles)
    │       ↓
    │     users
    │
    ├─→ domains
    │       ↓
    │     mailboxes
    │       ↓
    │     aliases
    │
    ├─→ quotas (limits)
    │
    └─→ audit_logs
```

### Indexing Strategy

```sql
-- Primary keys (automatic)
id UUID PRIMARY KEY

-- Foreign keys for joins
organization_id (all tables)
domain_id (mailboxes, aliases)
user_id (organization_users)

-- Lookup fields
email (users, mailboxes) - UNIQUE
slug (organizations) - UNIQUE
source (aliases) - UNIQUE

-- Query optimization
status (for filtering active records)
created_at (for ordering by date)
```

### Constraints

```sql
-- Prevent orphaned records
ON DELETE CASCADE

-- Enforce data integrity
CHECK (status IN ('active', 'suspended', 'deleted'))
CHECK (quota_bytes > 0)
CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$')

-- Unique constraints
UNIQUE(email)
UNIQUE(organization_id, user_id)  -- One role per user per org
```

## API Design

### REST Principles

```
Resource-oriented URLs:
  GET    /api/v1/mailboxes           # List
  POST   /api/v1/mailboxes           # Create
  GET    /api/v1/mailboxes/:id       # Get
  PATCH  /api/v1/mailboxes/:id       # Partial update
  DELETE /api/v1/mailboxes/:id       # Delete
  POST   /api/v1/mailboxes/:id/suspend  # Action

HTTP Status Codes:
  200 OK                    # Success (GET, PATCH)
  201 Created               # Success (POST)
  204 No Content            # Success (DELETE)
  400 Bad Request           # Invalid input
  401 Unauthorized          # Missing/invalid token
  403 Forbidden             # Insufficient permissions
  404 Not Found             # Resource doesn't exist
  409 Conflict              # Already exists
  422 Unprocessable Entity  # Validation failed
  429 Too Many Requests     # Rate limit exceeded
  500 Internal Server Error # Server error
```

### Error Response Format

```json
{
  "error": "validation_failed",
  "message": "Invalid email format",
  "details": {
    "field": "email",
    "value": "notanemail"
  },
  "request_id": "req_abc123"
}
```

### Pagination

```
Query params:
  ?limit=50          # Results per page (default: 20, max: 100)
  ?offset=0          # Skip results (default: 0)
  ?sort=created_at   # Sort field
  ?order=desc        # Sort order (asc/desc)

Response:
{
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 50,
    "offset": 0,
    "has_more": true
  }
}
```

## Scaling Strategy

### Horizontal Scaling

```
Load Balancer
    ↓
    ├─→ API Server 1
    ├─→ API Server 2
    ├─→ API Server 3
    │
    └─→ Worker Pool
        ├─→ Worker 1
        ├─→ Worker 2
        └─→ Worker 3
```

### Database Scaling

```
Primary (Write)
    ↓
    ├─→ Replica 1 (Read)
    ├─→ Replica 2 (Read)
    └─→ Replica 3 (Read)

Application routing:
- Writes → Primary
- Reads → Round-robin replicas
```

### Caching Strategy

```
L1 Cache: In-memory (per API server)
  - JWT validation results
  - User sessions
  - TTL: 5 minutes

L2 Cache: Redis (shared)
  - Organization quotas
  - Domain DNS records
  - Mailbox lists
  - TTL: 15 minutes
```

## Disaster Recovery

### Backup Strategy

```
Daily full backups:
  - PostgreSQL: pg_dump
  - Redis: RDB snapshots
  - Retention: 30 days

Hourly incremental:
  - WAL archiving
  - Retention: 7 days

Recovery Point Objective (RPO): 1 hour
Recovery Time Objective (RTO): 4 hours
```

### Failover Procedures

```
Database failure:
1. Promote read replica to primary (automatic)
2. Update connection strings (automated)
3. Verify write operations
4. Alert team

Service failure:
1. Health check fails
2. Load balancer removes from pool
3. Auto-scale launches replacement
4. New instance joins pool
```

## Performance Benchmarks

### Target Latency (p95)

```
Endpoint                    Target    Actual (should measure)
────────────────────────────────────────────────────────────
GET  /api/v1/mailboxes     < 100ms   TBD
POST /api/v1/mailboxes     < 200ms   TBD (sync only)
GET  /api/v1/domains       < 100ms   TBD
POST /api/v1/auth/login    < 200ms   TBD
```

### Throughput Targets

```
API Servers (per instance):
  - 200 req/sec sustained
  - 500 req/sec burst

Workers (per instance):
  - 100 jobs/sec provisioning
  - 1000 jobs/sec email sending
```

## Conclusion

This architecture provides:

✅ **Maintainability**: Clear boundaries, easy to understand
✅ **Testability**: Domain logic isolated, easy to mock
✅ **Scalability**: Stateless services, horizontal scaling
✅ **Flexibility**: Infrastructure can be swapped
✅ **Security**: Multi-layer authentication/authorization
✅ **Observability**: Structured logs, metrics, tracing
✅ **Reliability**: Async processing, retries, monitoring

The key to this architecture is **strict discipline**: never let infrastructure concerns leak into the domain layer. The control plane must remain completely independent of Mailcow, Brevo, or any other external system.

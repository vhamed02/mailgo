# Contributing to Mailbox

## Development Setup

### Prerequisites
- Go 1.21+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16 (optional, for local dev)
- Redis 7 (optional, for local dev)

### Getting Started

```bash
# Clone repository
git clone <repo-url>
cd mailgo

# Install dependencies
make deps

# Start infrastructure (Postgres, Redis)
docker-compose up -d postgres redis

# Run backend API
cd backend
go run cmd/api/main.go

# Run worker (separate terminal)
cd backend
go run cmd/worker/main.go

# Run frontend (separate terminal)
cd frontend
npm run dev
```

## Code Organization

### Backend Structure

```
backend/
├── cmd/              # Entry points
├── internal/
│   ├── domain/      # Business entities (no dependencies)
│   ├── application/ # Use cases (orchestration)
│   ├── api/         # HTTP/gRPC handlers
│   ├── infrastructure/ # External adapters
│   └── worker/      # Async job handlers
├── proto/           # gRPC definitions
└── migrations/      # Database migrations
```

### Frontend Structure

```
frontend/
├── app/             # Next.js App Router pages
├── components/      # Reusable React components
└── lib/             # Utilities and API clients
```

## Coding Standards

### Go Code Style

```go
// ✅ GOOD: Clear naming, proper error handling
func (s *MailboxService) CreateMailbox(ctx context.Context, req CreateMailboxRequest) (*domain.Mailbox, error) {
    // Validate input
    if req.LocalPart == "" {
        return nil, domain.ErrInvalidInput
    }
    
    // Check authorization
    if !s.canCreate(ctx, req.OrganizationID) {
        return nil, domain.ErrForbidden
    }
    
    // Business logic
    mailbox, err := s.create(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create mailbox: %w", err)
    }
    
    return mailbox, nil
}

// ❌ BAD: No error wrapping, unclear naming
func (s *MailboxService) Create(ctx context.Context, r CreateMailboxRequest) (*domain.Mailbox, error) {
    m, e := s.create(ctx, r)
    return m, e
}
```

### Architectural Rules

1. **Domain Layer Must Be Pure**
   ```go
   // ❌ NEVER import infrastructure in domain
   import "github.com/mailgo/backend/internal/infrastructure/mailcow"
   
   // ✅ Only define interfaces
   type MailServerAdapter interface {
       CreateMailbox(ctx context.Context, req CreateMailboxRequest) error
   }
   ```

2. **Always Filter by Tenant**
   ```go
   // ❌ NEVER query without organization_id
   func (r *Repo) ListAll() ([]*Mailbox, error)
   
   // ✅ Always include tenant context
   func (r *Repo) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Mailbox, error)
   ```

3. **Use Adapters for External Systems**
   ```go
   // ❌ Don't call external APIs directly
   resp, err := http.Post("https://mailcow.com/api/...")
   
   // ✅ Use adapter interface
   err := s.mailServer.CreateMailbox(ctx, req)
   ```

### TypeScript Code Style

```typescript
// ✅ GOOD: Proper types, error handling
async function createMailbox(data: CreateMailboxRequest): Promise<Mailbox> {
  try {
    const response = await apiClient.post<Mailbox>('/mailboxes', data)
    return response.data
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw new Error(error.response?.data?.message || 'Failed to create mailbox')
    }
    throw error
  }
}

// ❌ BAD: No types, poor error handling
async function createMailbox(data: any) {
  const response = await apiClient.post('/mailboxes', data)
  return response.data
}
```

## Testing Guidelines

### Unit Tests

```go
// backend/internal/application/mailbox_service_test.go
func TestCreateMailbox_Success(t *testing.T) {
    // Setup
    mockRepo := &mockMailboxRepository{}
    mockAdapter := &mockMailServerAdapter{}
    service := NewMailboxService(mockRepo, mockAdapter)
    
    // Execute
    mailbox, err := service.CreateMailbox(context.Background(), validRequest)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, mailbox)
    assert.Equal(t, "test@example.com", mailbox.Email)
}
```

### Integration Tests

```go
func TestMailboxAPI_E2E(t *testing.T) {
    // Start test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Start test API
    api := setupTestAPI(t, db)
    
    // Test flow
    token := registerAndLogin(t, api)
    domain := createDomain(t, api, token)
    mailbox := createMailbox(t, api, token, domain.ID)
    
    assert.NotNil(t, mailbox)
}
```

## Pull Request Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/add-alias-management
   ```

2. **Make Changes**
   - Follow coding standards
   - Add tests
   - Update documentation

3. **Run Tests**
   ```bash
   make test
   make lint
   ```

4. **Commit with Conventional Commits**
   ```
   feat: add alias management API
   fix: resolve database connection leak
   docs: update API documentation
   refactor: simplify mailbox service
   test: add integration tests for domains
   ```

5. **Push and Create PR**
   ```bash
   git push origin feature/add-alias-management
   ```

6. **PR Requirements**
   - [ ] Tests pass
   - [ ] Linter passes
   - [ ] Documentation updated
   - [ ] No breaking changes (or documented)
   - [ ] Reviewed by at least one maintainer

## Database Migrations

### Creating a Migration

```bash
# Install migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir backend/migrations -seq add_aliases_table
```

### Migration Rules

1. **Always provide up and down**
   ```sql
   -- 000002_add_aliases_table.up.sql
   CREATE TABLE aliases (...);
   
   -- 000002_add_aliases_table.down.sql
   DROP TABLE aliases;
   ```

2. **Never modify existing migrations**
   - Create a new migration instead

3. **Test rollback**
   ```bash
   migrate -path backend/migrations -database $DATABASE_URL up
   migrate -path backend/migrations -database $DATABASE_URL down 1
   ```

## API Design Guidelines

### RESTful Endpoints

```
✅ Good:
  POST   /api/v1/mailboxes              # Create
  GET    /api/v1/mailboxes              # List
  GET    /api/v1/mailboxes/:id          # Get
  PATCH  /api/v1/mailboxes/:id          # Update
  DELETE /api/v1/mailboxes/:id          # Delete
  POST   /api/v1/mailboxes/:id/suspend  # Action

❌ Bad:
  POST   /api/v1/createMailbox          # RPC-style
  GET    /api/v1/getMailbox?id=123      # Query params for resource
  POST   /api/v1/mailboxes/suspend      # Action without ID
```

### Request/Response Format

```json
// Request
{
  "local_part": "john",
  "domain_id": "uuid-here",
  "display_name": "John Doe",
  "quota_bytes": 1073741824
}

// Success Response (201)
{
  "id": "uuid-here",
  "email": "john@example.com",
  "display_name": "John Doe",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z"
}

// Error Response (400)
{
  "error": "validation_failed",
  "message": "Invalid local_part format",
  "details": {
    "field": "local_part",
    "constraint": "alphanumeric_only"
  }
}
```

## Documentation

### Code Comments

```go
// ✅ Good: Explain why, not what
// CreateMailbox provisions a mailbox in two steps:
// 1. Store in database (control plane)
// 2. Enqueue provisioning job (infrastructure)
// This decouples our system from Mailcow availability.
func (s *MailboxService) CreateMailbox(...) { }

// ❌ Bad: Obvious statement
// CreateMailbox creates a mailbox
func (s *MailboxService) CreateMailbox(...) { }
```

### API Documentation

Update `README.md` API section when adding endpoints:

```markdown
#### Mailboxes

- `POST /api/v1/mailboxes` - Create mailbox
  - Auth: Required
  - Role: Admin, Owner
  - Body: `CreateMailboxRequest`
  - Response: `Mailbox`
```

## Performance Considerations

### Database Queries

```go
// ✅ Good: Use indexes, limit results
SELECT * FROM mailboxes 
WHERE organization_id = $1 
  AND status = 'active'
ORDER BY created_at DESC
LIMIT 100;

// ❌ Bad: Full table scan, no limit
SELECT * FROM mailboxes;
```

### Caching Strategy

```go
// ✅ Cache expensive operations
func (s *Service) GetQuota(ctx context.Context, orgID uuid.UUID) (*Quota, error) {
    // Try cache first
    cached, err := s.cache.Get(ctx, fmt.Sprintf("quota:%s", orgID))
    if err == nil {
        return cached, nil
    }
    
    // Fetch from database
    quota, err := s.repo.GetQuota(ctx, orgID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache.Set(ctx, fmt.Sprintf("quota:%s", orgID), quota, 15*time.Minute)
    return quota, nil
}
```

## Security Guidelines

1. **Never Log Sensitive Data**
   ```go
   // ❌ Bad
   log.Info().Str("password", password).Msg("Creating user")
   
   // ✅ Good
   log.Info().Str("email", email).Msg("Creating user")
   ```

2. **Always Validate Input**
   ```go
   // ✅ Good
   if !isValidEmail(req.Email) {
       return domain.ErrInvalidInput
   }
   ```

3. **Use Parameterized Queries**
   ```go
   // ✅ Good
   db.Query("SELECT * FROM users WHERE email = $1", email)
   
   // ❌ Bad (SQL injection risk)
   db.Query(fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email))
   ```

## Getting Help

- Read `ARCHITECTURE.md` for technical design decisions
- Read `PRODUCTION_ASSESSMENT.md` for known issues
- Check existing issues before creating new ones
- Ask questions in discussions

## License

By contributing, you agree that your contributions will be licensed under the project's license.

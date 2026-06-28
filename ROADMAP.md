# MailGo — Development Roadmap

**Last updated**: June 28, 2026  
**Current state**: Authentication working end-to-end. All other features scaffolded but not wired.

---

## What's actually done right now

| Area | Status | Notes |
|------|--------|-------|
| User registration & login | ✅ Done | JWT tokens, bcrypt, validation |
| JWT auth middleware | ✅ Done | Protects routes, extracts claims |
| RBAC middleware | ✅ Done | Owner / Admin / User roles |
| Multi-tenant org management | ✅ Done | Orgs, slugs, org-user relationships |
| Database schema + migrations | ✅ Done | 8 tables, indexes, constraints |
| UserRepository | ✅ Done | Full CRUD + GetByEmail |
| OrganizationRepository | ✅ Done | Full CRUD + GetBySlug |
| OrganizationUserRepository | ✅ Done | Many-to-many with roles |
| QuotaRepository | ✅ Done | Create, Get, Update |
| AuditLogRepository | ✅ Done | Create, List by org/user |
| MailboxRepository | ✅ Done | Full CRUD, list by org/domain, count |
| DomainService (business logic) | ✅ Done | Add, get, list, verify, delete |
| MailboxService (business logic) | ✅ Done | Create, get, list, update, suspend, delete |
| Queue adapter (Redis/Asynq) | ✅ Done | Enqueue jobs |
| Mailcow adapter (HTTP) | ✅ Done | Create, update, delete, suspend mailboxes |
| Brevo adapter (HTTP) | ✅ Done | Send transactional email |
| Worker handlers | ✅ Done | Provision, delete, verify, send email |
| Docker Compose stack | ✅ Done | All 5 services running |
| Frontend login page | ✅ Done | Form, validation, token storage |
| Frontend register page | ✅ Done | Full form with org slug auto-gen |
| Frontend dashboard | ✅ Done | Stats, quick actions, org info |
| API client (axios) | ✅ Done | All endpoints defined, interceptors |
| Modal system (Radix UI) | ✅ Done | Replaces all `alert()` calls |

---

## Phase 1 — Complete the backend API
**Estimated effort: 2–3 days**

Everything in this phase has a working service layer already. These are all wiring tasks.

### 1.1 Domain HTTP handlers
File to create: `backend/internal/api/rest/handlers/domain_handler.go`

```
GET    /api/v1/domains            → DomainService.ListDomains
POST   /api/v1/domains            → DomainService.AddDomain
GET    /api/v1/domains/:id        → DomainService.GetDomain
DELETE /api/v1/domains/:id        → DomainService.DeleteDomain
POST   /api/v1/domains/:id/verify → DomainService.VerifyDomain
```

Then uncomment the 5 domain route lines in `backend/cmd/api/main.go`.

### 1.2 Mailbox HTTP handlers
File to create: `backend/internal/api/rest/handlers/mailbox_handler.go`

```
GET    /api/v1/mailboxes              → MailboxService.ListMailboxes
POST   /api/v1/mailboxes              → MailboxService.CreateMailbox
GET    /api/v1/mailboxes/:id          → MailboxService.GetMailbox
PATCH  /api/v1/mailboxes/:id          → MailboxService.UpdateMailbox
DELETE /api/v1/mailboxes/:id          → MailboxService.DeleteMailbox
POST   /api/v1/mailboxes/:id/suspend  → MailboxService.SuspendMailbox
POST   /api/v1/mailboxes/:id/unsuspend → MailboxService.UnsuspendMailbox
```

Then uncomment the 7 mailbox route lines in `backend/cmd/api/main.go`.

### 1.3 DomainRepository implementation
File to update: `backend/internal/infrastructure/postgres/stubs.go`

`DomainRepository` currently returns `nil` for everything. Needs real SQL queries — same pattern as `MailboxRepository` which is already fully implemented.

### 1.4 AliasRepository implementation
Same file as above. `AliasRepository` also stubs. Needed before alias management is built.

### 1.5 Transaction management in Register()
File to update: `backend/internal/application/auth_service.go`

`TxManager` already exists in `backend/internal/infrastructure/postgres/transaction.go`. Just needs to wrap the Register method so that creating user → org → org-user → quota is atomic. If any step fails, the whole thing rolls back.

### 1.6 Automatic database migrations on startup
File to update: `backend/cmd/api/main.go` (or the Dockerfile)

Migrations exist in `backend/migrations/` but require manual execution. The fix is to run `golang-migrate` programmatically before the server starts. This is a one-time ~30 min task.

### 1.7 Fix DNS verification
File to update: `backend/internal/application/domain_service.go`

`checkDNSRecords()` is hardcoded to return `false`, meaning no domain can ever be verified and therefore no mailbox can ever be created. Needs a real `net.LookupTXT` call to check SPF, DKIM, DMARC records.

### 1.8 Fix password handling in worker
File to update: `backend/internal/worker/handlers.go`

`HandleMailboxProvision` sends `mailbox.PasswordHash` (bcrypt) to Mailcow instead of the actual plaintext password. Mailcow will reject this. The correct approach: store the plaintext password temporarily in Redis with a short TTL when the mailbox is created, pass only the Redis key in the job payload, have the worker retrieve and delete it.

---

## Phase 2 — Production hardening
**Estimated effort: 3–4 days**

### 2.1 Structured logging throughout
Replace all `fmt.Printf(...)` calls in service files with `zerolog` (`log.Error().Err(err).Msg(...)`). The logger is already initialized in `main.go` and imported in the worker — just not used in the application layer.

Files affected:
- `backend/internal/application/auth_service.go`
- `backend/internal/application/domain_service.go`
- `backend/internal/application/mailbox_service.go`

### 2.2 Circuit breakers for external adapters
Both `mailcow/adapter.go` and `brevo/adapter.go` make single HTTP requests with no fallback. If either service goes down, every request hangs for 30 seconds until timeout. Add `github.com/sony/gobreaker` around the HTTP calls in both adapters.

### 2.3 Retry logic in adapters
Related to 2.2. Add `github.com/cenkalti/backoff/v4` exponential backoff to the `makeRequest` helper in both adapters so transient network errors don't cause permanent failures.

### 2.4 Per-tenant rate limiting
Current rate limiting is a global 100 req/sec memory store. One tenant can exhaust this for everyone. Replace with a Redis-backed rate limiter keyed on `organization_id` extracted from the JWT claims.

### 2.5 Prometheus metrics
`/metrics` endpoint in `main.go` currently returns a stub JSON response. Replace with `promhttp.Handler()` and instrument key operations: request count, request latency, active connections, job queue depth, mailbox provisioning latency.

### 2.6 Storage quota enforcement
`QuotaRepository` is fully implemented and quotas are checked for mailbox count and domain count, but total storage (GB) is never checked before creating a mailbox. Add the calculation in `MailboxService.CreateMailbox`.

### 2.7 Dead letter queue monitoring
Asynq has a DLQ built in, but nothing monitors it. Add an admin endpoint to list failed jobs and a mechanism to retry them. Optionally add alerting (webhook or email) when jobs land in the DLQ.

---

## Phase 3 — Frontend pages
**Estimated effort: 4–5 days**

The API client (`frontend/lib/api-client.ts`) already has all the endpoint methods defined. This phase is purely UI work.

### 3.1 Domain management page
Route: `/dashboard/domains`

Features:
- List domains with status badges (pending / active / failed)
- Add domain form with name validation
- DNS records panel — show SPF, DKIM, DMARC values to copy-paste
- Trigger verification button
- Delete domain (with confirmation modal)

### 3.2 Mailbox management page
Route: `/dashboard/mailboxes`

Features:
- List mailboxes with status, storage usage bar, domain
- Create mailbox form (pick domain, set local part, password, display name, quota)
- Edit mailbox (display name, password, quota)
- Suspend / unsuspend toggle
- Delete mailbox (with confirmation modal)

### 3.3 Team management page
Route: `/dashboard/team`

Features:
- List team members with role badges
- Invite by email (sends invitation email via Brevo)
- Change role (owner can demote/promote admins)
- Remove member

### 3.4 Settings page
Route: `/dashboard/settings`

Features:
- Organization name / slug update
- Danger zone — delete organization

### 3.5 Navigation sidebar
The dashboard currently has no navigation. Add a persistent sidebar with links to Domains, Mailboxes, Team, Settings. Replace the single-page layout with a shared layout component for all `/dashboard/*` routes.

### 3.6 Token refresh flow
`api-client.ts` intercepts 401s and redirects to login, but never attempts a token refresh first. Add a retry interceptor that calls `POST /api/v1/auth/refresh` with the stored refresh token before giving up and redirecting.

---

## Phase 4 — Testing
**Estimated effort: 4–5 days**

Currently zero test coverage.

### 4.1 Unit tests for services
Test the application layer in isolation using mock implementations of the repository and adapter interfaces. Go's interfaces make this straightforward — no mocking framework needed.

Priority order:
1. `AuthService` — registration, login, token validation
2. `DomainService` — add domain, quota checks, DNS validation
3. `MailboxService` — create, suspend, delete, quota checks

### 4.2 Repository integration tests
Spin up a test PostgreSQL instance (use `testcontainers-go`) and run the real SQL against it. Verify the repositories handle edge cases: not found, duplicate email, cascade deletes.

### 4.3 API end-to-end tests
Use Go's `net/http/httptest` to test the full HTTP stack. Cover the happy path and error cases for each endpoint group.

### 4.4 Frontend tests
Add `vitest` + `@testing-library/react`. Test the auth forms, the modal components, and the API client interceptors.

---

## Phase 5 — Pre-launch
**Estimated effort: 3–4 days**

### 5.1 HTTPS / reverse proxy
Add Nginx or Traefik in front of the API and frontend containers. Configure Let's Encrypt for TLS. Enforce HTTPS redirect.

### 5.2 Security headers
Add `Strict-Transport-Security`, `X-Content-Type-Options`, `X-Frame-Options`, `Content-Security-Policy` via middleware.

### 5.3 OpenAPI / Swagger documentation
Generate an OpenAPI spec from the Echo routes (use `swaggo/echo-swagger`). Makes onboarding and third-party integrations much easier.

### 5.4 Production Docker Compose
The current `docker-compose.yml` is for development (mounts source, no restart policies). Create a `docker-compose.prod.yml` with:
- `restart: unless-stopped` on all services
- No volume mounts of source code
- Production environment variable references
- Proper health check intervals
- Log rotation config

### 5.5 Backup and restore
Add a cron job or scheduled task to run `pg_dump` daily and ship the archive to S3 or equivalent. Document the restore procedure.

### 5.6 Load testing
Run `k6` or `vegeta` against the API before going live. Validate the system handles expected concurrency without connection pool exhaustion (the current pool cap is 25 connections).

---

## Summary table

| Phase | What | Effort | Blocker for |
|-------|------|--------|------------|
| 1 | Backend API wiring | 2–3 days | Everything else |
| 2 | Production hardening | 3–4 days | Launch |
| 3 | Frontend pages | 4–5 days | User-facing MVP |
| 4 | Testing | 4–5 days | Confidence / launch |
| 5 | Pre-launch | 3–4 days | Production deploy |

**Realistic time to a working MVP (phases 1 + 3):** ~1 week  
**Realistic time to production-ready (all phases):** ~4–5 weeks

# Webmail — Implementation Plan

## Overview

A fully custom webmail UI built into MailGo, opening in a new tab, backed by a Go IMAP/SMTP bridge. No third-party webmail (no Roundcube, no SOGo).

---

## Architecture

```
Browser (/webmail)
    ↕ REST + JWT
Go API (/api/v1/mail/*)
    ↕ IMAP (read)  /  SMTP (send)
Mailcow (Dovecot + Postfix)
```

The browser never talks to IMAP directly. The Go API acts as a bridge, authenticating to Mailcow on behalf of the user.

---

## Password Handling (Option A — Session-only)

Mailbox passwords are bcrypt-hashed in the DB and cannot be reversed. The webmail needs the plaintext password to authenticate to IMAP/SMTP.

**Flow:**
1. User clicks "Open Mailbox" on the mailboxes page → opens `/webmail?mailbox=info@yerevan.digital` in a new tab
2. Webmail page prompts for the mailbox password (one-time per session)
3. Password is held only in React state / `sessionStorage` — never persisted to DB or localStorage
4. Every API call to `/api/v1/mail/*` includes the password in a custom header `X-Mailbox-Password` (HTTPS only)
5. Backend uses it to authenticate to IMAP/SMTP, never logs or stores it

---

## Backend — New Endpoints

All under `/api/v1/mail/` — require JWT auth + `X-Mailbox-Password` header.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/mail/folders` | List IMAP folders |
| GET | `/api/v1/mail/folders/:name/messages` | List messages in folder (paginated, ?page=1&limit=50) |
| GET | `/api/v1/mail/messages/:uid` | Get full message (headers + HTML body) |
| PATCH | `/api/v1/mail/messages/:uid/read` | Mark as read/unread |
| DELETE | `/api/v1/mail/messages/:uid` | Move to Trash |
| POST | `/api/v1/mail/compose` | Send new email |
| POST | `/api/v1/mail/reply` | Reply to a message |

### Request context (all mail endpoints)
```
Header: Authorization: Bearer <jwt>        — identifies the MailGo user
Header: X-Mailbox-Address: info@example.com — which mailbox to open
Header: X-Mailbox-Password: plaintext       — IMAP/SMTP credential
```

### Go IMAP library
`github.com/emersion/go-imap/v2` — maintained, RFC-compliant

### Go SMTP library
`net/smtp` from stdlib — sufficient for sending

---

## New Files — Backend

```
backend/internal/infrastructure/imap/
    adapter.go          — IMAP connection, folder list, message fetch, flag operations

backend/internal/infrastructure/smtp/
    adapter.go          — SMTP send / reply

backend/internal/domain/
    mail.go             — Mail entities (Message, Folder, Attachment, etc.)
    mail_interfaces.go  — IMAPAdapter, SMTPAdapter interfaces

backend/internal/application/
    mail_service.go     — Business logic: list, read, compose, reply

backend/internal/api/rest/handlers/
    mail_handler.go     — HTTP handlers for all /mail/* endpoints
```

---

## Frontend — New Pages

### Route
`/webmail` — opens in new tab from the mailboxes page

### URL structure
`/webmail?mailbox=info@yerevan.digital` — pre-selects the mailbox

### Layout
```
┌─────────────────────────────────────────────────────────┐
│  [MailGo Webmail]   info@yerevan.digital ▾   [Compose]  │  ← top bar
├──────────┬──────────────────────┬──────────────────────-─┤
│          │                      │                        │
│ Folders  │   Email List         │   Email Reader         │
│          │                      │   (or Compose panel)   │
│ Inbox 12 │   From / Subject /   │                        │
│ Sent     │   Date               │   From:                │
│ Drafts   │   ─────────────────  │   To:                  │
│ Trash    │   ...                │   Subject:             │
│          │                      │                        │
│ ─────    │                      │   [HTML body]          │
│ Switch   │                      │                        │
│ mailbox  │                      │   [Reply] [Forward]    │
└──────────┴──────────────────────┴────────────────────────┘
```

### New Files — Frontend

```
frontend/app/webmail/
    page.tsx               — main webmail shell (auth gate + layout)
    layout.tsx             — minimal layout (no sidebar nav from dashboard)

frontend/components/webmail/
    MailboxUnlock.tsx      — password prompt modal
    FolderList.tsx         — folder sidebar
    MessageList.tsx        — email list (sender, subject, date, unread indicator)
    MessageReader.tsx      — full email view with HTML rendering
    ComposeModal.tsx       — compose new email
    ReplyPanel.tsx         — inline reply form

frontend/lib/mail-api.ts   — API client for /api/v1/mail/* endpoints
```

---

## Message Entity

```go
type Message struct {
    UID         uint32
    MessageID   string
    From        string
    To          []string
    CC          []string
    Subject     string
    Date        time.Time
    IsRead      bool
    HasAttachment bool
    BodyHTML    string
    BodyText    string
    Folder      string
}

type Folder struct {
    Name        string
    DisplayName string
    UnreadCount int
    TotalCount  int
}
```

---

## Security Notes

- `X-Mailbox-Password` only travels over HTTPS in production
- Never logged, never stored, lives only in memory during request
- IMAP connection is opened per-request (stateless) — no persistent connections
- JWT still required — proves the user has access to this MailGo account
- The mailbox address is validated against the user's org before connecting to IMAP

---

## Phases

| Phase | Work | Est. |
|-------|------|------|
| 1 | IMAP adapter + domain entities + mail_service | 1 day |
| 2 | HTTP handlers + routes wired into API | 0.5 day |
| 3 | Frontend: layout, folder list, message list | 1 day |
| 4 | Frontend: message reader, HTML rendering | 0.5 day |
| 5 | Frontend: compose + reply | 1 day |
| 6 | Polish, error states, loading skeletons | 0.5 day |

**Total: ~4.5 days**

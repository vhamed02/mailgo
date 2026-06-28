# Webmail — Progress

## Status: 🔨 In Progress

---

## Phase 1 — IMAP adapter + domain entities + mail_service
**Status: ✅ Done**

- [x] Add `go-imap/v2` dependency
- [x] `domain/mail.go` — Message, Folder entities
- [x] `domain/mail_interfaces.go` — IMAPAdapter, SMTPAdapter interfaces (in mail.go)
- [x] `infrastructure/imap/adapter.go` — IMAP connection, folders, messages
- [x] `infrastructure/smtp/adapter.go` — SMTP send/reply
- [x] `application/mail_service.go` — business logic layer

## Phase 2 — HTTP handlers + routes
**Status: ✅ Done**

- [x] `api/rest/handlers/mail_handler.go` — ListFolders, ListMessages, GetMessage, MarkRead, DeleteMessage, Compose, Reply
- [x] Routes wired in `main.go` under `/api/v1/mail/*`
- [x] IMAP/SMTP host/port from env vars

## Phase 3 — Frontend layout + folder list + message list
**Status: ✅ Done**

- [x] `/webmail` route + standalone layout
- [x] `UnlockScreen` — password prompt (session-only, never stored)
- [x] Folder sidebar with active state
- [x] Message list with unread indicators, pagination
- [x] `lib/mail-api.ts` — typed API client

## Phase 4 — Message reader
**Status: ✅ Done**

- [x] `MessageReader` inline in webmail page
- [x] HTML email rendering via sandboxed iframe
- [x] Plaintext fallback
- [x] Auto mark-as-read on open

## Phase 5 — Compose + Reply
**Status: ✅ Done**

- [x] `ComposeModal` — floating compose window
- [x] Reply pre-fills To/Subject/In-Reply-To
- [x] Webmail button on mailboxes dashboard opens new tab
- [x] Multi-mailbox switcher in sidebar

## Phase 6 — Polish
**Status: ✅ Done**

- [x] Loading spinners on message list and reader
- [x] Empty states
- [x] Pagination controls
- [x] Session-based password per mailbox (survives tab reload, clears on close)

- [ ] `api/rest/handlers/mail_handler.go`
- [ ] Routes wired in `main.go`

## Phase 3 — Frontend layout + folder list + message list
**Status: ⏳ Pending**

- [ ] `/webmail` route + layout
- [ ] `MailboxUnlock.tsx` — password prompt
- [ ] `FolderList.tsx`
- [ ] `MessageList.tsx`
- [ ] `lib/mail-api.ts`

## Phase 4 — Message reader
**Status: ⏳ Pending**

- [ ] `MessageReader.tsx`
- [ ] HTML email rendering (safe sandbox)

## Phase 5 — Compose + Reply
**Status: ⏳ Pending**

- [ ] `ComposeModal.tsx`
- [ ] `ReplyPanel.tsx`

## Phase 6 — Polish
**Status: ⏳ Pending**

- [ ] Loading skeletons
- [ ] Error states
- [ ] Unread counts
- [ ] Mark read on open

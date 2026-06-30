package application

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/mail"
	"strings"

	"github.com/mailgo/backend/internal/domain"
)

// normalizeEmailAddress extracts the plain email address from an RFC 5322
// address string such as "John Doe <john@example.com>" or "john@example.com".
// If parsing fails it returns the trimmed input unchanged so we never silently
// drop an address.
func normalizeEmailAddress(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	addr, err := mail.ParseAddress(raw)
	if err != nil {
		// Already a bare address (or malformed – return as-is).
		return raw
	}
	return addr.Address
}

// normalizeEmailList normalises every entry in a slice.
func normalizeEmailList(list []string) []string {
	out := make([]string, 0, len(list))
	for _, raw := range list {
		out = append(out, normalizeEmailAddress(raw))
	}
	return out
}

type MailService struct {
	imap          domain.IMAPAdapter
	emailProvider domain.EmailSenderAdapter
	mailboxRepo   domain.MailboxRepository
	fromName      string
	domain        string
}

func NewMailService(imap domain.IMAPAdapter, emailProvider domain.EmailSenderAdapter, fromName, sendDomain string) *MailService {
	return &MailService{imap: imap, emailProvider: emailProvider, fromName: fromName, domain: sendDomain}
}

// WithMailboxRepo allows the mail service to resolve display names from the DB.
func (s *MailService) WithMailboxRepo(repo domain.MailboxRepository) *MailService {
	s.mailboxRepo = repo
	return s
}

// senderDisplayName returns the mailbox DisplayName from DB if available,
// falling back to the global fromName config value.
func (s *MailService) senderDisplayName(ctx context.Context, mailboxAddr string) string {
	if s.mailboxRepo != nil {
		mb, err := s.mailboxRepo.GetByEmail(ctx, mailboxAddr)
		if err == nil && mb.DisplayName != "" {
			return mb.DisplayName
		}
	}
	return s.fromName
}

// generateMessageID creates a fresh RFC 2822 Message-ID for every new message.
func generateMessageID(sendDomain string) string {
	if sendDomain == "" {
		sendDomain = "mailgo.local"
	}
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("<%x.%x@%s>", b[:8], b[8:], sendDomain)
}

// buildReferences produces the References header for a reply:
// previous References + previous Message-ID (RFC 5322 §3.6.4).
func buildReferences(prevRefs, prevMessageID string) string {
	parts := strings.Fields(prevRefs)
	if prevMessageID != "" {
		parts = append(parts, prevMessageID)
	}
	seen := make(map[string]bool, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

func (s *MailService) ListFolders(addr, password string) ([]*domain.MailFolder, error) {
	return s.imap.ListFolders(addr, password)
}

func (s *MailService) ListMessages(addr, password, folder string, page, limit int) ([]*domain.MailMessage, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.imap.ListMessages(addr, password, folder, page, limit)
}

func (s *MailService) GetMessage(addr, password, folder string, uid uint32) (*domain.MailMessage, error) {
	return s.imap.GetMessage(addr, password, folder, uid)
}

func (s *MailService) MarkRead(addr, password, folder string, uid uint32, read bool) error {
	return s.imap.MarkRead(addr, password, folder, uid, read)
}

func (s *MailService) MoveToTrash(addr, password, folder string, uid uint32) error {
	return s.imap.MoveToTrash(addr, password, folder, uid)
}

func (s *MailService) Compose(req domain.ComposeRequest) error {
	displayName := s.senderDisplayName(context.Background(), req.MailboxAddress)
	if req.From == "" {
		req.From = fmt.Sprintf("%s <%s>", displayName, req.MailboxAddress)
	}

	log.Printf("[Compose] raw To: %v", req.To)
	req.To = normalizeEmailList(req.To)
	req.CC = normalizeEmailList(req.CC)
	log.Printf("[Compose] normalized To: %v", req.To)

	if err := s.emailProvider.SendTransactionalEmail(context.Background(), domain.SendEmailRequest{
		To:        req.To,
		Subject:   req.Subject,
		Body:      req.Body,
		IsHTML:    req.IsHTML,
		From:      &domain.EmailAddress{Email: req.MailboxAddress, Name: displayName},
		InReplyTo: req.InReplyTo,
	}); err != nil {
		return err
	}
	_ = s.imap.AppendSent(req.MailboxAddress, req.MailboxPassword, req)
	return nil
}

func (s *MailService) Reply(req domain.ComposeRequest) error {
	displayName := s.senderDisplayName(context.Background(), req.MailboxAddress)
	if req.From == "" {
		req.From = fmt.Sprintf("%s <%s>", displayName, req.MailboxAddress)
	}
	if !hasRePrefix(req.Subject) {
		req.Subject = "Re: " + req.Subject
	}

	log.Printf("[Reply] raw To: %v", req.To)
	req.To = normalizeEmailList(req.To)
	req.CC = normalizeEmailList(req.CC)
	log.Printf("[Reply] normalized To: %v", req.To)

	// Generate a fresh Message-ID for this reply and build the References chain.
	if req.MessageID == "" {
		req.MessageID = generateMessageID(s.domain)
	}
	if req.InReplyTo != "" && req.References == "" {
		req.References = buildReferences("", req.InReplyTo)
	}

	if err := s.emailProvider.SendTransactionalEmail(context.Background(), domain.SendEmailRequest{
		To:        req.To,
		Subject:   req.Subject,
		Body:      req.Body,
		IsHTML:    req.IsHTML,
		From:      &domain.EmailAddress{Email: req.MailboxAddress, Name: displayName},
		InReplyTo: req.InReplyTo,
	}); err != nil {
		return err
	}
	_ = s.imap.AppendSent(req.MailboxAddress, req.MailboxPassword, req)
	return nil
}

func hasRePrefix(s string) bool {
	return strings.HasPrefix(s, "Re:") || strings.HasPrefix(s, "re:")
}

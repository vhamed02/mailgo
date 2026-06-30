package application

import (
	"context"
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
	fromName      string
}

func NewMailService(imap domain.IMAPAdapter, emailProvider domain.EmailSenderAdapter, fromName string) *MailService {
	return &MailService{imap: imap, emailProvider: emailProvider, fromName: fromName}
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
	if req.From == "" {
		req.From = fmt.Sprintf("%s <%s>", s.fromName, req.MailboxAddress)
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
		From:      &domain.EmailAddress{Email: req.MailboxAddress, Name: s.fromName},
		InReplyTo: req.InReplyTo,
	}); err != nil {
		return err
	}
	_ = s.imap.AppendSent(req.MailboxAddress, req.MailboxPassword, req)
	return nil
}

func (s *MailService) Reply(req domain.ComposeRequest) error {
	if req.From == "" {
		req.From = fmt.Sprintf("%s <%s>", s.fromName, req.MailboxAddress)
	}
	if !hasRePrefix(req.Subject) {
		req.Subject = "Re: " + req.Subject
	}

	log.Printf("[Reply] raw To: %v", req.To)
	req.To = normalizeEmailList(req.To)
	req.CC = normalizeEmailList(req.CC)
	log.Printf("[Reply] normalized To: %v", req.To)

	if err := s.emailProvider.SendTransactionalEmail(context.Background(), domain.SendEmailRequest{
		To:        req.To,
		Subject:   req.Subject,
		Body:      req.Body,
		IsHTML:    req.IsHTML,
		From:      &domain.EmailAddress{Email: req.MailboxAddress, Name: s.fromName},
		InReplyTo: req.InReplyTo,
	}); err != nil {
		return err
	}
	_ = s.imap.AppendSent(req.MailboxAddress, req.MailboxPassword, req)
	return nil
}

func hasRePrefix(s string) bool {
	if len(s) >= 3 && (s[:3] == "Re:" || s[:3] == "re:") {
		return true
	}
	return false
}

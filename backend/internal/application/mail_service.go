package application

import (
	"fmt"

	"github.com/mailgo/backend/internal/domain"
)

type MailService struct {
	imap domain.IMAPAdapter
	smtp domain.SMTPAdapter
	fromName string
}

func NewMailService(imap domain.IMAPAdapter, smtp domain.SMTPAdapter, fromName string) *MailService {
	return &MailService{imap: imap, smtp: smtp, fromName: fromName}
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
	return s.smtp.Send(req)
}

func (s *MailService) Reply(req domain.ComposeRequest) error {
	if req.From == "" {
		req.From = fmt.Sprintf("%s <%s>", s.fromName, req.MailboxAddress)
	}
	if !hasRePrefix(req.Subject) {
		req.Subject = "Re: " + req.Subject
	}
	return s.smtp.Send(req)
}

func hasRePrefix(s string) bool {
	if len(s) >= 3 && (s[:3] == "Re:" || s[:3] == "re:") {
		return true
	}
	return false
}

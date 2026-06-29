package domain

import "time"

type MailFolder struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	UnreadCount int    `json:"unread_count"`
	TotalCount  int    `json:"total_count"`
}

type MailMessage struct {
	UID           uint32    `json:"uid"`
	MessageID     string    `json:"message_id"`
	From          string    `json:"from"`
	To            []string  `json:"to"`
	CC            []string  `json:"cc"`
	Subject       string    `json:"subject"`
	Date          time.Time `json:"date"`
	IsRead        bool      `json:"is_read"`
	HasAttachment bool      `json:"has_attachment"`
	Folder        string    `json:"folder"`
	BodyHTML      string    `json:"body_html,omitempty"`
	BodyText      string    `json:"body_text,omitempty"`
	Snippet       string    `json:"snippet"`
}

type ComposeRequest struct {
	MailboxAddress  string
	MailboxPassword string
	From            string
	To              []string
	CC              []string
	Subject         string
	Body            string
	IsHTML          bool
	InReplyTo       string
}

type IMAPAdapter interface {
	ListFolders(addr, password string) ([]*MailFolder, error)
	ListMessages(addr, password, folder string, page, limit int) ([]*MailMessage, int, error)
	GetMessage(addr, password, folder string, uid uint32) (*MailMessage, error)
	MarkRead(addr, password, folder string, uid uint32, read bool) error
	MoveToTrash(addr, password, folder string, uid uint32) error
	AppendSent(addr, password string, req ComposeRequest) error
}

type SMTPAdapter interface {
	Send(req ComposeRequest) error
}

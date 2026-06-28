package smtp

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/mailgo/backend/internal/domain"
)

type Adapter struct {
	host string
	port int
}

func NewAdapter(host string, port int) *Adapter {
	return &Adapter{host: host, port: port}
}

func (a *Adapter) Send(req domain.ComposeRequest) error {
	auth := smtp.PlainAuth("", req.MailboxAddress, req.MailboxPassword, a.host)

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", req.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(req.To, ", ")))
	if len(req.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(req.CC, ", ")))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", req.Subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	if req.InReplyTo != "" {
		buf.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", req.InReplyTo))
		buf.WriteString(fmt.Sprintf("References: %s\r\n", req.InReplyTo))
	}
	if req.IsHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	buf.WriteString("\r\n")
	buf.WriteString(req.Body)

	allRecipients := append(req.To, req.CC...)
	addr := fmt.Sprintf("%s:%d", a.host, a.port)

	if err := smtp.SendMail(addr, auth, req.MailboxAddress, allRecipients, buf.Bytes()); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

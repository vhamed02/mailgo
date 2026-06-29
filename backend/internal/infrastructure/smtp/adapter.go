package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
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
	addr := fmt.Sprintf("%s:%d", a.host, a.port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	c, err := smtp.NewClient(conn, a.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	tlsCfg := &tls.Config{InsecureSkipVerify: true, ServerName: a.host}
	if err := c.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}

	auth := smtp.PlainAuth("", req.MailboxAddress, req.MailboxPassword, a.host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := c.Mail(req.MailboxAddress); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	all := append(req.To, req.CC...)
	for _, rcpt := range all {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

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

	if _, err := wc.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return c.Quit()
}

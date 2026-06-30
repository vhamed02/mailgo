package imap

import (
	"crypto/tls"
	"fmt"
	"io"
	"strings"
	"time"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/mailgo/backend/internal/domain"
)

type Adapter struct {
	host   string
	port   int
	useTLS bool
}

func NewAdapter(host string, port int, useTLS bool) *Adapter {
	return &Adapter{host: host, port: port, useTLS: useTLS}
}

func (a *Adapter) connect(addr, password string) (*imapclient.Client, error) {
	target := fmt.Sprintf("%s:%d", a.host, a.port)
	var c *imapclient.Client
	var err error
	if a.useTLS {
		c, err = imapclient.DialTLS(target, &imapclient.Options{
			TLSConfig: &tls.Config{InsecureSkipVerify: true},
		})
	} else {
		c, err = imapclient.DialInsecure(target, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("imap dial: %w", err)
	}
	if err := c.Login(addr, password).Wait(); err != nil {
		c.Close()
		return nil, fmt.Errorf("imap login: %w", err)
	}
	return c, nil
}

func (a *Adapter) ListFolders(addr, password string) ([]*domain.MailFolder, error) {
	c, err := a.connect(addr, password)
	if err != nil {
		return nil, err
	}
	defer c.Logout()

	mailboxes, err := c.List("", "*", nil).Collect()
	if err != nil {
		return nil, fmt.Errorf("imap list: %w", err)
	}

	var folders []*domain.MailFolder
	for _, mb := range mailboxes {
		folders = append(folders, &domain.MailFolder{
			Name:        mb.Mailbox,
			DisplayName: friendlyName(mb.Mailbox),
		})
	}
	return folders, nil
}

func (a *Adapter) ListMessages(addr, password, folder string, page, limit int) ([]*domain.MailMessage, int, error) {
	c, err := a.connect(addr, password)
	if err != nil {
		return nil, 0, err
	}
	defer c.Logout()

	selData, err := c.Select(folder, nil).Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("imap select: %w", err)
	}
	total := int(selData.NumMessages)
	if total == 0 {
		return []*domain.MailMessage{}, 0, nil
	}

	high := uint32(total - (page-1)*limit)
	if high < 1 {
		return []*domain.MailMessage{}, total, nil
	}
	low := uint32(1)
	if total-page*limit+1 > 0 {
		low = uint32(total - page*limit + 1)
	}

	seqSet := imap.SeqSet{imap.SeqRange{Start: low, Stop: high}}

	msgs, err := c.Fetch(seqSet, &imap.FetchOptions{
		Flags:    true,
		Envelope: true,
		UID:      true,
	}).Collect()
	if err != nil {
		return nil, total, fmt.Errorf("imap fetch: %w", err)
	}

	var result []*domain.MailMessage
	for _, msg := range msgs {
		m := &domain.MailMessage{
			UID:    uint32(msg.UID),
			Folder: folder,
			IsRead: hasFlag(msg.Flags, "\\Seen"),
		}
		if msg.Envelope != nil {
			m.Subject = msg.Envelope.Subject
			m.Date = msg.Envelope.Date
			m.MessageID = msg.Envelope.MessageID
			if len(msg.Envelope.From) > 0 {
				m.From = formatAddress(msg.Envelope.From[0])
			}
			for _, to := range msg.Envelope.To {
				m.To = append(m.To, formatAddress(to))
			}
		}
		result = append(result, m)
	}

	reverseMessages(result)
	return result, total, nil
}

func (a *Adapter) GetMessage(addr, password, folder string, uid uint32) (*domain.MailMessage, error) {
	c, err := a.connect(addr, password)
	if err != nil {
		return nil, err
	}
	defer c.Logout()

	if _, err := c.Select(folder, nil).Wait(); err != nil {
		return nil, fmt.Errorf("imap select: %w", err)
	}

	seqSet := imap.SeqSetNum(uid)
	bodySec := &imap.FetchItemBodySection{}

	msgs, err := c.Fetch(seqSet, &imap.FetchOptions{
		Flags:       true,
		Envelope:    true,
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{bodySec},
	}).Collect()
	if err != nil || len(msgs) == 0 {
		return nil, domain.ErrNotFound
	}

	msg := msgs[0]
	m := &domain.MailMessage{
		UID:    uint32(msg.UID),
		Folder: folder,
		IsRead: hasFlag(msg.Flags, "\\Seen"),
	}

	if msg.Envelope != nil {
		m.Subject = msg.Envelope.Subject
		m.Date = msg.Envelope.Date
		m.MessageID = msg.Envelope.MessageID
		if len(msg.Envelope.From) > 0 {
			m.From = formatAddress(msg.Envelope.From[0])
		}
		for _, to := range msg.Envelope.To {
			m.To = append(m.To, formatAddress(to))
		}
		for _, cc := range msg.Envelope.Cc {
			m.CC = append(m.CC, formatAddress(cc))
		}
	}

	if raw := msg.FindBodySection(bodySec); raw != nil {
		mr, err := mail.CreateReader(strings.NewReader(string(raw)))
		if err == nil {
			for {
				part, err := mr.NextPart()
				if err != nil {
					break
				}
				ct := part.Header.Get("Content-Type")
				partBody, _ := io.ReadAll(part.Body)
				switch {
				case strings.HasPrefix(ct, "text/html"):
					m.BodyHTML = string(partBody)
				case strings.HasPrefix(ct, "text/plain"):
					if m.BodyText == "" {
						m.BodyText = string(partBody)
					}
				}
			}
		} else {
			m.BodyText = string(raw)
		}
	}

	return m, nil
}

func (a *Adapter) MarkRead(addr, password, folder string, uid uint32, read bool) error {
	c, err := a.connect(addr, password)
	if err != nil {
		return err
	}
	defer c.Logout()

	if _, err := c.Select(folder, nil).Wait(); err != nil {
		return err
	}

	op := imap.StoreFlagsDel
	if read {
		op = imap.StoreFlagsAdd
	}

	cmd := c.Store(imap.SeqSetNum(uid), &imap.StoreFlags{
		Op:     op,
		Silent: true,
		Flags:  []imap.Flag{"\\Seen"},
	}, nil)
	return cmd.Close()
}

func (a *Adapter) AppendSent(addr, password string, req domain.ComposeRequest) error {
	c, err := a.connect(addr, password)
	if err != nil {
		return err
	}
	defer c.Logout()

	var buf strings.Builder
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

	raw := []byte(buf.String())

	appendCmd := c.Append("Sent", int64(len(raw)), &imap.AppendOptions{
		Flags: []imap.Flag{"\\Seen"},
		Time:  time.Now(),
	})
	if _, err := appendCmd.Write(raw); err != nil {
		appendCmd.Close()
		return fmt.Errorf("imap append write: %w", err)
	}
	if err := appendCmd.Close(); err != nil {
		return fmt.Errorf("imap append close: %w", err)
	}
	_, err = appendCmd.Wait()
	return err
}

func (a *Adapter) MoveToTrash(addr, password, folder string, uid uint32) error {
	c, err := a.connect(addr, password)
	if err != nil {
		return err
	}
	defer c.Logout()

	if _, err := c.Select(folder, nil).Wait(); err != nil {
		return err
	}

	_, err = c.Move(imap.SeqSetNum(uid), "Trash").Wait()
	return err
}

func friendlyName(name string) string {
	switch strings.ToUpper(name) {
	case "INBOX":
		return "Inbox"
	case "SENT", "SENT MESSAGES", "SENT ITEMS":
		return "Sent"
	case "DRAFTS":
		return "Drafts"
	case "TRASH", "DELETED", "DELETED ITEMS", "DELETED MESSAGES":
		return "Trash"
	case "SPAM", "JUNK":
		return "Spam"
	default:
		return name
	}
}

func formatAddress(addr imap.Address) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s@%s>", addr.Name, addr.Mailbox, addr.Host)
	}
	return fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
}

func hasFlag(flags []imap.Flag, target string) bool {
	for _, f := range flags {
		if strings.EqualFold(string(f), target) {
			return true
		}
	}
	return false
}

func reverseMessages(msgs []*domain.MailMessage) {
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
}

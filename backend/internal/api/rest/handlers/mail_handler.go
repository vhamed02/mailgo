package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mailgo/backend/internal/application"
	"github.com/mailgo/backend/internal/domain"
)

type MailHandler struct {
	service *application.MailService
}

func NewMailHandler(service *application.MailService) *MailHandler {
	return &MailHandler{service: service}
}

func (h *MailHandler) mailboxCreds(c echo.Context) (addr, password string, err error) {
	addr = c.Request().Header.Get("X-Mailbox-Address")
	password = c.Request().Header.Get("X-Mailbox-Password")
	if addr == "" || password == "" {
		return "", "", errors.New("missing X-Mailbox-Address or X-Mailbox-Password header")
	}
	return addr, password, nil
}

func mailErr(c echo.Context, err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
	}
	msg := err.Error()
	if strings.Contains(msg, "imap login") {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid mailbox credentials"})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": msg})
}

func (h *MailHandler) ListFolders(c echo.Context) error {
	addr, password, err := h.mailboxCreds(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	folders, err := h.service.ListFolders(addr, password)
	if err != nil {
		return mailErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": folders})
}

func (h *MailHandler) ListMessages(c echo.Context) error {
	addr, password, err := h.mailboxCreds(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	folder := c.Param("folder")
	if folder == "" {
		folder = "INBOX"
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	msgs, total, err := h.service.ListMessages(addr, password, folder, page, limit)
	if err != nil {
		return mailErr(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  msgs,
		"total": total,
	})
}

func (h *MailHandler) GetMessage(c echo.Context) error {
	addr, password, err := h.mailboxCreds(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	folder := c.QueryParam("folder")
	if folder == "" {
		folder = "INBOX"
	}

	uid64, err := strconv.ParseUint(c.Param("uid"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid uid"})
	}

	msg, err := h.service.GetMessage(addr, password, folder, uint32(uid64))
	if err != nil {
		return mailErr(c, err)
	}

	_ = h.service.MarkRead(addr, password, folder, uint32(uid64), true)

	return c.JSON(http.StatusOK, msg)
}

func (h *MailHandler) MarkRead(c echo.Context) error {
	addr, password, err := h.mailboxCreds(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	folder := c.QueryParam("folder")
	if folder == "" {
		folder = "INBOX"
	}

	uid64, err := strconv.ParseUint(c.Param("uid"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid uid"})
	}

	var body struct {
		Read bool `json:"read"`
	}
	body.Read = true
	_ = c.Bind(&body)

	if err := h.service.MarkRead(addr, password, folder, uint32(uid64), body.Read); err != nil {
		return mailErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

func (h *MailHandler) DeleteMessage(c echo.Context) error {
	addr, password, err := h.mailboxCreds(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	folder := c.QueryParam("folder")
	if folder == "" {
		folder = "INBOX"
	}

	uid64, err := strconv.ParseUint(c.Param("uid"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid uid"})
	}

	if err := h.service.MoveToTrash(addr, password, folder, uint32(uid64)); err != nil {
		return mailErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

type ComposeDTO struct {
	To         []string `json:"to"         validate:"required"`
	CC         []string `json:"cc"`
	Subject    string   `json:"subject"    validate:"required"`
	Body       string   `json:"body"       validate:"required"`
	IsHTML     bool     `json:"is_html"`
	InReplyTo  string   `json:"in_reply_to"`
	References string   `json:"references"`
}

func (h *MailHandler) Compose(c echo.Context) error {
	addr, password, err := h.mailboxCreds(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req ComposeDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.service.Compose(domain.ComposeRequest{
		MailboxAddress:  addr,
		MailboxPassword: password,
		To:              req.To,
		CC:              req.CC,
		Subject:         req.Subject,
		Body:            req.Body,
		IsHTML:          req.IsHTML,
		InReplyTo:       req.InReplyTo,
	}); err != nil {
		return mailErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

func (h *MailHandler) Reply(c echo.Context) error {
	addr, password, err := h.mailboxCreds(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var req ComposeDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := h.service.Reply(domain.ComposeRequest{
		MailboxAddress:  addr,
		MailboxPassword: password,
		To:              req.To,
		CC:              req.CC,
		Subject:         req.Subject,
		Body:            req.Body,
		IsHTML:          req.IsHTML,
		InReplyTo:       req.InReplyTo,
		References:      req.References,
	}); err != nil {
		return mailErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

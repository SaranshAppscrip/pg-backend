package notification

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/nivas/server/internal/config"
)

type StaffInviteParams struct {
	To               string
	OrganizationName string
	OrganizationID   string
	Email            string
	TempPassword     string
}

type PasswordResetParams struct {
	To        string
	ResetURL  string
	ForTenant bool
}

type RentReminderParams struct {
	To           string
	TenantName   string
	PropertyName string
	ForMonth     string
	MonthlyFee   float64
	Paid         float64
	Due          float64
	ReminderType string // due | overdue
}

type PaymentReceiptParams struct {
	To       string
	TenantName string
	Subject  string
	Text     string
	HTML     string
	PDF      []byte
	Filename string
}

type EmailSender interface {
	SendStaffInvite(ctx context.Context, p StaffInviteParams) error
	SendPasswordReset(ctx context.Context, p PasswordResetParams) error
	SendRentReminder(ctx context.Context, p RentReminderParams) error
	SendPaymentReceipt(ctx context.Context, p PaymentReceiptParams) error
}

type resendSender struct {
	cfg    config.EmailConfig
	appEnv string
	log    *slog.Logger
	client *http.Client
}

func NewEmailSender(cfg config.EmailConfig, appEnv string, log *slog.Logger) EmailSender {
	return &resendSender{
		cfg:    cfg,
		appEnv: appEnv,
		log:    log,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *resendSender) SendStaffInvite(ctx context.Context, p StaffInviteParams) error {
	loginURL := strings.TrimRight(s.cfg.FrontendURL, "/") + "/login"
	subject := fmt.Sprintf("You're invited to %s on Nivas", p.OrganizationName)
	text := staffInviteText(p, loginURL)
	htmlBody := staffInviteHTML(p, loginURL)
	return s.send(ctx, p.To, subject, text, htmlBody, nil)
}

func (s *resendSender) SendPasswordReset(ctx context.Context, p PasswordResetParams) error {
	subject := "Reset your Nivas password"
	text := passwordResetText(p)
	htmlBody := passwordResetHTML(p)
	return s.send(ctx, p.To, subject, text, htmlBody, nil)
}

func (s *resendSender) SendRentReminder(ctx context.Context, p RentReminderParams) error {
	subject := rentReminderSubject(p)
	text := rentReminderText(p)
	htmlBody := rentReminderHTML(p)
	return s.send(ctx, p.To, subject, text, htmlBody, nil)
}

func (s *resendSender) SendPaymentReceipt(ctx context.Context, p PaymentReceiptParams) error {
	var attachments []resendAttachment
	if len(p.PDF) > 0 {
		attachments = []resendAttachment{{
			Filename: p.Filename,
			Content:  base64.StdEncoding.EncodeToString(p.PDF),
		}}
	}
	return s.send(ctx, p.To, p.Subject, p.Text, p.HTML, attachments)
}

type resendAttachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type resendPayload struct {
	From        string             `json:"from"`
	To          []string           `json:"to"`
	Subject     string             `json:"subject"`
	Text        string             `json:"text"`
	HTML        string             `json:"html"`
	Attachments []resendAttachment `json:"attachments,omitempty"`
}

type resendErrorBody struct {
	Message string `json:"message"`
}

// ResendErrorMessage extracts a user-facing message from a Resend API error.
func ResendErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	const prefix = "resend API error"
	idx := strings.Index(msg, prefix)
	if idx < 0 {
		return ""
	}
	jsonPart := msg[idx+len(prefix):]
	if colon := strings.Index(jsonPart, ":"); colon >= 0 {
		jsonPart = strings.TrimSpace(jsonPart[colon+1:])
	}
	var body resendErrorBody
	if json.Unmarshal([]byte(jsonPart), &body) == nil && body.Message != "" {
		return body.Message
	}
	return ""
}

func (s *resendSender) send(ctx context.Context, to, subject, text, htmlBody string, attachments []resendAttachment) error {
	if s.cfg.ResendAPIKey == "" {
		if s.appEnv == "development" {
			s.log.Info("email (dev mode, not sent)",
				"to", to,
				"subject", subject,
				"text", text,
			)
			return nil
		}
		return fmt.Errorf("RESEND_API_KEY is not configured")
	}

	actualTo := to
	if s.appEnv == "development" && s.cfg.DevRedirectTo != "" && !strings.EqualFold(to, s.cfg.DevRedirectTo) {
		s.log.Info("email dev redirect",
			"intended_to", to,
			"redirected_to", s.cfg.DevRedirectTo,
		)
		actualTo = s.cfg.DevRedirectTo
		text, htmlBody = wrapDevRedirect(to, text, htmlBody)
		subject = "[Dev] " + subject
	}

	payload, err := json.Marshal(resendPayload{
		From:        s.cfg.From,
		To:          []string{actualTo},
		Subject:     subject,
		Text:        text,
		HTML:        htmlBody,
		Attachments: attachments,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("resend API error %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

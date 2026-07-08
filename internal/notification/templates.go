package notification

import (
	"fmt"
	"html"
	"strings"
)

func staffInviteText(p StaffInviteParams, loginURL string) string {
	return fmt.Sprintf(`Hello,

You have been invited to join %s on Nivas.

Sign in: %s

Email: %s
Temporary password: %s

After signing in, use "Forgot password" on the login page to set your own password.

— Nivas
`, p.OrganizationName, loginURL, p.Email, p.TempPassword)
}

func staffInviteHTML(p StaffInviteParams, loginURL string) string {
	org := html.EscapeString(p.OrganizationName)
	email := html.EscapeString(p.Email)
	pass := html.EscapeString(p.TempPassword)
	url := html.EscapeString(loginURL)

	content := fmt.Sprintf(`
<p style="margin:0 0 16px;font-size:16px;line-height:1.6;color:#3d3d3d;">Hello,</p>
<p style="margin:0 0 24px;font-size:16px;line-height:1.6;color:#3d3d3d;">
  You have been invited to join <strong>%s</strong> on Nivas.
</p>
%s
<table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="margin:0 0 24px;background:#faf7f2;border:1px solid #e8e0d4;border-radius:8px;">
  <tr>
    <td style="padding:20px;">
      <p style="margin:0 0 12px;font-size:13px;font-weight:600;color:#8b7355;text-transform:uppercase;letter-spacing:0.04em;">Your sign-in details</p>
      <p style="margin:0 0 8px;font-size:15px;line-height:1.5;color:#3d3d3d;"><strong>Email:</strong> %s</p>
      <p style="margin:0;font-size:15px;line-height:1.5;color:#3d3d3d;"><strong>Temporary password:</strong> <code style="font-family:ui-monospace,monospace;background:#fff;padding:2px 6px;border-radius:4px;">%s</code></p>
    </td>
  </tr>
</table>
<p style="margin:0 0 8px;font-size:14px;line-height:1.6;color:#6b6b6b;">
  After your first sign-in, use <strong>Forgot password</strong> on the login page to choose your own password.
</p>`,
		org,
		ctaButton("Sign in to Nivas", url),
		email,
		pass,
	)

	return emailLayout("You're invited to Nivas", content, fallbackLink("Sign in", url))
}

func passwordResetText(p PasswordResetParams) string {
	account := "staff"
	if p.ForTenant {
		account = "tenant"
	}
	return fmt.Sprintf(`Hello,

We received a request to reset your Nivas %s password.

Reset your password here (link expires in 1 hour):
%s

If you did not request this, you can ignore this email.

— Nivas
`, account, p.ResetURL)
}

func passwordResetHTML(p PasswordResetParams) string {
	account := "staff"
	if p.ForTenant {
		account = "tenant"
	}
	url := html.EscapeString(p.ResetURL)

	content := fmt.Sprintf(`
<p style="margin:0 0 16px;font-size:16px;line-height:1.6;color:#3d3d3d;">Hello,</p>
<p style="margin:0 0 24px;font-size:16px;line-height:1.6;color:#3d3d3d;">
  We received a request to reset your Nivas <strong>%s</strong> password. Click the button below to choose a new password.
</p>
%s
<p style="margin:0;font-size:14px;line-height:1.6;color:#6b6b6b;">
  This link expires in <strong>1 hour</strong>. If you did not request a password reset, you can safely ignore this email.
</p>`,
		html.EscapeString(account),
		ctaButton("Reset password", url),
	)

	return emailLayout("Reset your Nivas password", content, fallbackLink("Reset password", url))
}

func ctaButton(label, href string) string {
	return fmt.Sprintf(`
<table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 0 24px;">
  <tr>
    <td style="border-radius:8px;background:#c45c6a;">
      <a href="%s" target="_blank" style="display:inline-block;padding:14px 28px;font-size:16px;font-weight:600;color:#ffffff;text-decoration:none;border-radius:8px;">%s</a>
    </td>
  </tr>
</table>`, href, html.EscapeString(label))
}

func fallbackLink(label, href string) string {
	return fmt.Sprintf(`If the button doesn't work, copy and paste this link into your browser:<br>
<a href="%s" style="color:#c45c6a;word-break:break-all;">%s</a>`, href, html.EscapeString(href))
}

func emailLayout(title, content, fallback string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background:#f5f0e8;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="background:#f5f0e8;padding:32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="max-width:560px;background:#ffffff;border:1px solid #e8e0d4;border-radius:12px;overflow:hidden;">
          <tr>
            <td style="padding:28px 32px 12px;text-align:center;border-bottom:1px solid #f0ebe3;">
              <p style="margin:0;font-family:Georgia,'Times New Roman',serif;font-size:28px;font-weight:600;color:#c45c6a;">Nivas</p>
            </td>
          </tr>
          <tr>
            <td style="padding:32px;">
              %s
            </td>
          </tr>
          <tr>
            <td style="padding:0 32px 28px;">
              <p style="margin:0;font-size:13px;line-height:1.6;color:#9a9a9a;border-top:1px solid #f0ebe3;padding-top:20px;">
                %s
              </p>
            </td>
          </tr>
        </table>
        <p style="margin:16px 0 0;font-size:12px;color:#9a9a9a;">&copy; Nivas PG Management</p>
      </td>
    </tr>
  </table>
</body>
</html>`, html.EscapeString(title), content, fallback)
}

func wrapDevRedirect(recipient, text, htmlBody string) (string, string) {
	prefix := fmt.Sprintf("[Dev redirect — intended recipient: %s]\n\n", recipient)
	text = prefix + text
	htmlPrefix := fmt.Sprintf(`<p style="margin:0 0 16px;padding:12px;background:#fff3cd;border:1px solid #ffc107;border-radius:6px;font-size:13px;color:#664d03;"><strong>Dev redirect</strong> — intended recipient: %s</p>`, html.EscapeString(recipient))
	htmlBody = strings.Replace(htmlBody, `<td style="padding:32px;">`, `<td style="padding:32px;">`+htmlPrefix, 1)
	return text, htmlBody
}

func rentReminderSubject(p RentReminderParams) string {
	if p.ReminderType == "overdue" {
		return fmt.Sprintf("Overdue rent reminder — %s", p.ForMonth)
	}
	return fmt.Sprintf("Rent due reminder — %s", p.ForMonth)
}

func rentReminderText(p RentReminderParams) string {
	return fmt.Sprintf(`Hello %s,

This is a friendly reminder about your rent for %s at %s.

Monthly rent: Rs. %.0f
Paid so far: Rs. %.0f
Balance due: Rs. %.0f

Please contact the property office if you have already paid.

— Nivas
`, p.TenantName, p.ForMonth, p.PropertyName, p.MonthlyFee, p.Paid, p.Due)
}

func rentReminderHTML(p RentReminderParams) string {
	intro := "Your rent for this month is due soon."
	if p.ReminderType == "overdue" {
		intro = "Your rent payment is overdue."
	}
	content := fmt.Sprintf(`
<p style="margin:0 0 16px;font-size:16px;line-height:1.6;color:#3d3d3d;">Hello %s,</p>
<p style="margin:0 0 24px;font-size:16px;line-height:1.6;color:#3d3d3d;">%s</p>
<table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="margin:0 0 24px;background:#faf7f2;border:1px solid #e8e0d4;border-radius:8px;">
  <tr><td style="padding:20px;">
    <p style="margin:0 0 8px;font-size:15px;color:#3d3d3d;"><strong>Property:</strong> %s</p>
    <p style="margin:0 0 8px;font-size:15px;color:#3d3d3d;"><strong>Month:</strong> %s</p>
    <p style="margin:0 0 8px;font-size:15px;color:#3d3d3d;"><strong>Monthly rent:</strong> Rs. %.0f</p>
    <p style="margin:0 0 8px;font-size:15px;color:#3d3d3d;"><strong>Paid:</strong> Rs. %.0f</p>
    <p style="margin:0;font-size:15px;color:#c45c6a;"><strong>Balance due:</strong> Rs. %.0f</p>
  </td></tr>
</table>`,
		html.EscapeString(p.TenantName), intro,
		html.EscapeString(p.PropertyName), html.EscapeString(p.ForMonth),
		p.MonthlyFee, p.Paid, p.Due,
	)
	return emailLayout(rentReminderSubject(p), content, "")
}

func PaymentReceiptHTML(d struct {
	TenantName, PropertyName, OrganizationName, Date, ForMonth, Mode string
	Amount float64
	PaymentID string
}) (string, string) {
	subject := fmt.Sprintf("Payment receipt — %s", d.ForMonth)
	text := fmt.Sprintf(`Hello %s,

We received your payment of Rs. %.0f for %s (%s).

Property: %s
Date: %s
Mode: %s
Receipt ID: %s

— %s
`, d.TenantName, d.Amount, d.ForMonth, d.PropertyName, d.PropertyName, d.Date, d.Mode, d.PaymentID, d.OrganizationName)

	content := fmt.Sprintf(`
<p style="margin:0 0 16px;font-size:16px;line-height:1.6;color:#3d3d3d;">Hello %s,</p>
<p style="margin:0 0 24px;font-size:16px;line-height:1.6;color:#3d3d3d;">Thank you — we received your payment of <strong>Rs. %.0f</strong> for <strong>%s</strong>.</p>
<table role="presentation" cellpadding="0" cellspacing="0" width="100%%" style="background:#faf7f2;border:1px solid #e8e0d4;border-radius:8px;">
  <tr><td style="padding:20px;">
    <p style="margin:0 0 8px;font-size:15px;color:#3d3d3d;"><strong>Property:</strong> %s</p>
    <p style="margin:0 0 8px;font-size:15px;color:#3d3d3d;"><strong>Date:</strong> %s</p>
    <p style="margin:0 0 8px;font-size:15px;color:#3d3d3d;"><strong>Mode:</strong> %s</p>
    <p style="margin:0;font-size:13px;color:#6b6b6b;">Receipt ID: %s</p>
  </td></tr>
</table>
<p style="margin:16px 0 0;font-size:14px;color:#6b6b6b;">A PDF copy is attached for your records.</p>`,
		html.EscapeString(d.TenantName), d.Amount, html.EscapeString(d.ForMonth),
		html.EscapeString(d.PropertyName), html.EscapeString(d.Date), html.EscapeString(d.Mode), html.EscapeString(d.PaymentID),
	)
	return subject, emailLayout(subject, content, text)
}

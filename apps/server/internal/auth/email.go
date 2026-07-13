package auth

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/acmhot100/server/internal/config"
)

// SendVerificationEmail sends an email verification link to the user.
// In development or mock mode, it just logs the email instead of sending.
func SendVerificationEmail(cfg *config.Config, toEmail string, rawToken string, baseURL string) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", strings.TrimRight(baseURL, "/"), rawToken)
	subject := "Verify your email - ACM Hot 100"
	htmlBody := fmt.Sprintf(`
<html><body>
<h2>Welcome to ACM Hot 100!</h2>
<p>Please verify your email address by clicking the link below:</p>
<p><a href="%s">Verify Email</a></p>
<p>If you did not create an account, you can safely ignore this email.</p>
<p>This link expires in 30 minutes.</p>
</body></html>
`, verifyURL)

	return sendEmail(cfg, toEmail, subject, htmlBody)
}

// SendResetPasswordEmail sends a password reset link to the user.
// In development or mock mode, it just logs the email instead of sending.
func SendResetPasswordEmail(cfg *config.Config, toEmail string, rawToken string, baseURL string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(baseURL, "/"), rawToken)
	subject := "Reset your password - ACM Hot 100"
	htmlBody := fmt.Sprintf(`
<html><body>
<h2>ACM Hot 100 - Password Reset</h2>
<p>You requested to reset your password. Click the link below:</p>
<p><a href="%s">Reset Password</a></p>
<p>If you did not request a password reset, you can safely ignore this email.</p>
<p>This link expires in 20 minutes.</p>
</body></html>
`, resetURL)

	return sendEmail(cfg, toEmail, subject, htmlBody)
}

// sendEmail sends an email or logs it in development/mock mode.
func sendEmail(cfg *config.Config, toEmail string, subject string, htmlBody string) error {
	// In development or mock mode, just log the email
	if cfg.AppEnv == "development" || cfg.JudgeMode == "mock" {
		log.Printf("[DEV EMAIL] To: %s, Subject: %s", toEmail, subject)
		log.Printf("[DEV EMAIL] Body: %s", htmlBody)
		return nil
	}

	// Validate SMTP config
	if cfg.SMTPHost == "" || cfg.SMTPFrom == "" {
		log.Printf("[EMAIL] SMTP not configured, skipping email to %s", toEmail)
		return nil
	}

	from := mail.Address{Name: "ACM Hot 100", Address: cfg.SMTPFrom}
	to := mail.Address{Address: toEmail}

	// Build email headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build the full email message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	// Connect and send
	addr := cfg.SMTPAddr()
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)

	var err error
	switch cfg.SMTPTLSMode {
	case "tls":
		err = sendWithTLS(addr, auth, from.Address, []string{to.Address}, []byte(msg.String()), cfg.SMTPHost)
	default:
		// "none" or "starttls" - use standard smtp.SendMail which uses STARTTLS if available
		err = smtp.SendMail(addr, auth, from.Address, []string{to.Address}, []byte(msg.String()))
	}

	if err != nil {
		log.Printf("[EMAIL] Failed to send email to %s: %v", toEmail, err)
		return err
	}

	return nil
}

// sendWithTLS sends an email using implicit TLS (port 465).
func sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer c.Close()

	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	if err = c.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}

	for _, rcpt := range to {
		if err = c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("SMTP RCPT TO failed: %w", err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write failed: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP close failed: %w", err)
	}

	return c.Quit()
}

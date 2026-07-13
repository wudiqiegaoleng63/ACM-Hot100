package auth

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/acmhot100/server/internal/config"
)

func TestSendVerificationEmailUsesConfiguredSMTPInDevelopment(t *testing.T) {
	messages, address := startSMTPServer(t)
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("split SMTP address: %v", err)
	}

	cfg := &config.Config{
		AppEnv:      "development",
		MailMode:    "smtp",
		SMTPHost:    host,
		SMTPPort:    mustParsePort(t, port),
		SMTPFrom:    "no-reply@example.local",
		SMTPTLSMode: "none",
	}

	if err := SendVerificationEmail(cfg, "user@example.local", "raw-token", "http://localhost:5173"); err != nil {
		t.Fatalf("send verification email: %v", err)
	}

	select {
	case message := <-messages:
		if !strings.Contains(message, "verify-email?token=raw-token") {
			t.Fatalf("SMTP message missing verification link: %s", message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("configured SMTP server did not receive a message")
	}
}

func TestSendVerificationEmailRequiresSMTPConfiguration(t *testing.T) {
	cfg := &config.Config{MailMode: "smtp"}
	if err := SendVerificationEmail(cfg, "user@example.local", "raw-token", "http://localhost:5173"); err == nil {
		t.Fatal("SMTP mode accepted missing SMTP configuration")
	}
}

func TestSendVerificationEmailLogModeSkipsSMTP(t *testing.T) {
	cfg := &config.Config{
		AppEnv:   "development",
		MailMode: "log",
		SMTPHost: "127.0.0.1",
		SMTPPort: 1,
		SMTPFrom: "no-reply@example.local",
	}
	if err := SendVerificationEmail(cfg, "user@example.local", "raw-token", "http://localhost:5173"); err != nil {
		t.Fatalf("log mail mode returned an error: %v", err)
	}
}

func startSMTPServer(t *testing.T) (<-chan string, string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for SMTP: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	messages := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)
		writeSMTPLine(writer, "220 localhost ESMTP")
		var message strings.Builder
		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			trimmed := strings.TrimRight(line, "\r\n")
			if inData {
				if trimmed == "." {
					messages <- message.String()
					writeSMTPLine(writer, "250 queued")
					inData = false
					continue
				}
				message.WriteString(line)
				continue
			}

			command := strings.ToUpper(strings.Fields(trimmed)[0])
			switch command {
			case "EHLO":
				writeSMTPLine(writer, "250-localhost")
				writeSMTPLine(writer, "250 OK")
			case "HELO", "MAIL", "RCPT", "RSET":
				writeSMTPLine(writer, "250 OK")
			case "DATA":
				writeSMTPLine(writer, "354 End data with <CR><LF>.<CR><LF>")
				inData = true
			case "QUIT":
				writeSMTPLine(writer, "221 Bye")
				return
			default:
				writeSMTPLine(writer, "250 OK")
			}
		}
	}()

	return messages, listener.Addr().String()
}

func writeSMTPLine(writer *bufio.Writer, line string) {
	_, _ = fmt.Fprintf(writer, "%s\r\n", line)
	_ = writer.Flush()
}

func mustParsePort(t *testing.T, value string) int {
	t.Helper()
	var port int
	if _, err := fmt.Sscanf(value, "%d", &port); err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return port
}

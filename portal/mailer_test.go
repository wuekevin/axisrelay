package portal

import (
	"bufio"
	"context"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSMTPSenderDeliversMultipartVerificationMessage(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	received := make(chan string, 1)
	go serveFakeSMTP(listener, received)

	host, rawPort, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(rawPort)
	config := smtpConfig{Host: host, Port: port, FromAddress: "noreply@example.com", FromName: "AxisRelay", TLSMode: "none", PublicURL: "https://relay.example", Timeout: 2 * time.Second}
	message := verificationMessage(config, "user@example.com", "one-time-token")
	if err := (&smtpSender{config: config}).Send(context.Background(), message); err != nil {
		t.Fatal(err)
	}

	select {
	case body := <-received:
		for _, expected := range []string{"multipart/alternative", "text/plain", "text/html", "Subject: =?UTF-8?", "完成邮箱验证", "https://relay.example/auth/verify-email?token=one-time-token"} {
			if !strings.Contains(body, expected) {
				t.Fatalf("message missing %q", expected)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("fake SMTP server did not receive a message")
	}
}

func TestPasswordResetMessageHasExpectedExpiryAndLink(t *testing.T) {
	message := passwordResetMessage(smtpConfig{PublicURL: "https://relay.example"}, "user@example.com", "reset-token")
	if !strings.Contains(message.TextBody, "30 分钟") || !strings.Contains(message.HTMLBody, "/auth/reset-password?token=reset-token") {
		t.Fatal("password reset template is missing expiry or reset link")
	}
}

func serveFakeSMTP(listener net.Listener, received chan<- string) {
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	writeSMTPLine(writer, "220 fake-smtp ESMTP")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(command, "EHLO"):
			writeSMTPLine(writer, "250-fake-smtp")
			writeSMTPLine(writer, "250 8BITMIME")
		case strings.HasPrefix(command, "MAIL FROM"), strings.HasPrefix(command, "RCPT TO"):
			writeSMTPLine(writer, "250 OK")
		case command == "DATA":
			writeSMTPLine(writer, "354 End data with <CR><LF>.<CR><LF>")
			var body strings.Builder
			for {
				dataLine, readErr := reader.ReadString('\n')
				if readErr != nil {
					return
				}
				if dataLine == ".\r\n" {
					break
				}
				body.WriteString(dataLine)
			}
			received <- body.String()
			writeSMTPLine(writer, "250 queued")
		case command == "QUIT":
			writeSMTPLine(writer, "221 bye")
			return
		default:
			writeSMTPLine(writer, "250 OK")
		}
	}
}

func writeSMTPLine(writer *bufio.Writer, line string) {
	_, _ = io.WriteString(writer, line+"\r\n")
	_ = writer.Flush()
}

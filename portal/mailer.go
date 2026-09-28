package portal

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type emailMessage struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
}

type emailSender interface {
	Send(context.Context, emailMessage) error
}

type smtpConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	TLSMode     string
	PublicURL   string
	Timeout     time.Duration
}

type smtpSender struct{ config smtpConfig }

func smtpConfigFromEnv() (smtpConfig, error) {
	port := 587
	if raw := strings.TrimSpace(os.Getenv("AXISRELAY_SMTP_PORT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 65535 {
			return smtpConfig{}, errors.New("AXISRELAY_SMTP_PORT must be a valid TCP port")
		}
		port = parsed
	}
	config := smtpConfig{
		Host:        strings.TrimSpace(os.Getenv("AXISRELAY_SMTP_HOST")),
		Port:        port,
		Username:    strings.TrimSpace(os.Getenv("AXISRELAY_SMTP_USERNAME")),
		Password:    os.Getenv("AXISRELAY_SMTP_PASSWORD"),
		FromAddress: strings.TrimSpace(os.Getenv("AXISRELAY_SMTP_FROM_ADDRESS")),
		FromName:    strings.TrimSpace(os.Getenv("AXISRELAY_SMTP_FROM_NAME")),
		TLSMode:     strings.ToLower(strings.TrimSpace(os.Getenv("AXISRELAY_SMTP_TLS_MODE"))),
		PublicURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("AXISRELAY_PUBLIC_BASE_URL")), "/"),
		Timeout:     10 * time.Second,
	}
	if config.FromName == "" {
		config.FromName = "AxisRelay"
	}
	if config.TLSMode == "" {
		config.TLSMode = "starttls"
	}
	if config.Host == "" || config.FromAddress == "" || config.PublicURL == "" {
		return smtpConfig{}, errors.New("public registration requires AXISRELAY_PUBLIC_BASE_URL, AXISRELAY_SMTP_HOST and AXISRELAY_SMTP_FROM_ADDRESS")
	}
	if _, err := mail.ParseAddress(config.FromAddress); err != nil {
		return smtpConfig{}, errors.New("AXISRELAY_SMTP_FROM_ADDRESS is invalid")
	}
	publicURL, err := url.Parse(config.PublicURL)
	if err != nil || publicURL.Host == "" || (publicURL.Scheme != "https" && publicURL.Scheme != "http") {
		return smtpConfig{}, errors.New("AXISRELAY_PUBLIC_BASE_URL must be an absolute http or https URL")
	}
	if config.TLSMode != "starttls" && config.TLSMode != "implicit" && config.TLSMode != "none" {
		return smtpConfig{}, errors.New("AXISRELAY_SMTP_TLS_MODE must be starttls, implicit, or none")
	}
	if (config.Username == "") != (config.Password == "") {
		return smtpConfig{}, errors.New("SMTP username and password must be configured together")
	}
	return config, nil
}

func (s *smtpSender) Send(ctx context.Context, message emailMessage) error {
	if _, err := mail.ParseAddress(message.To); err != nil {
		return errors.New("invalid recipient")
	}
	deadline := time.Now().Add(s.config.Timeout)
	address := net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port))
	dialer := &net.Dialer{Timeout: s.config.Timeout}
	var conn net.Conn
	var err error
	if s.config.TLSMode == "implicit" {
		conn, err = tls.DialWithDialer(dialer, "tcp", address, &tls.Config{ServerName: s.config.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return fmt.Errorf("connect to mail server: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(deadline)

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("initialize mail session: %w", err)
	}
	defer client.Close()
	if s.config.TLSMode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("mail server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: s.config.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("start mail TLS: %w", err)
		}
	}
	if s.config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)); err != nil {
			return fmt.Errorf("authenticate mail session: %w", err)
		}
	}
	if err := client.Mail(s.config.FromAddress); err != nil {
		return fmt.Errorf("set mail sender: %w", err)
	}
	if err := client.Rcpt(message.To); err != nil {
		return fmt.Errorf("set mail recipient: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("open mail body: %w", err)
	}
	if _, err = w.Write(buildMIMEMessage(s.config, message)); err != nil {
		_ = w.Close()
		return fmt.Errorf("write mail body: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("send mail body: %w", err)
	}
	return client.Quit()
}

func buildMIMEMessage(config smtpConfig, message emailMessage) []byte {
	boundary := "axisrelay-alternative-7f2a1d"
	var body bytes.Buffer
	fmt.Fprintf(&body, "From: %s\r\n", (&mail.Address{Name: sanitizeHeader(config.FromName), Address: config.FromAddress}).String())
	fmt.Fprintf(&body, "To: %s\r\n", (&mail.Address{Address: message.To}).String())
	fmt.Fprintf(&body, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", sanitizeHeader(message.Subject)))
	body.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&body, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
	fmt.Fprintf(&body, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", boundary, normalizeMailBody(message.TextBody))
	fmt.Fprintf(&body, "--%s\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", boundary, normalizeMailBody(message.HTMLBody))
	fmt.Fprintf(&body, "--%s--\r\n", boundary)
	return body.Bytes()
}

func sanitizeHeader(value string) string {
	return strings.TrimSpace(strings.NewReplacer("\r", " ", "\n", " ").Replace(value))
}

func normalizeMailBody(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\n", "\r\n")
}

func verificationMessage(config smtpConfig, recipient, token string) emailMessage {
	link := config.PublicURL + "/auth/verify-email?token=" + url.QueryEscape(token)
	return actionMessage(recipient, "验证你的 AxisRelay 邮箱", "完成邮箱验证", "请确认这是你本人创建的 AxisRelay 账号。链接将在 24 小时后失效。", link)
}

func passwordResetMessage(config smtpConfig, recipient, token string) emailMessage {
	link := config.PublicURL + "/auth/reset-password?token=" + url.QueryEscape(token)
	return actionMessage(recipient, "重置你的 AxisRelay 密码", "重置登录密码", "如果你没有发起密码重置，可以忽略这封邮件。链接将在 30 分钟后失效。", link)
}

func actionMessage(recipient, subject, title, description, link string) emailMessage {
	escapedTitle := template.HTMLEscapeString(title)
	escapedDescription := template.HTMLEscapeString(description)
	escapedLink := template.HTMLEscapeString(link)
	return emailMessage{
		To:       recipient,
		Subject:  subject,
		TextBody: title + "\n\n" + description + "\n\n" + link + "\n\n为保护账号安全，请勿转发此链接。",
		HTMLBody: `<!doctype html><html><body style="margin:0;background:#f4f7fb;font-family:Arial,sans-serif;color:#132238"><div style="max-width:560px;margin:40px auto;background:#fff;border:1px solid #dbe4ef;border-radius:16px;padding:32px"><div style="font-size:20px;font-weight:700;color:#0b66e4">AxisRelay</div><h1 style="font-size:24px;margin:28px 0 12px">` + escapedTitle + `</h1><p style="line-height:1.7;color:#52657a">` + escapedDescription + `</p><p style="margin:28px 0"><a href="` + escapedLink + `" style="display:inline-block;background:#0b66e4;color:white;text-decoration:none;padding:12px 22px;border-radius:10px;font-weight:700">继续</a></p><p style="font-size:13px;line-height:1.6;color:#738499">若按钮无法打开，请复制以下地址到浏览器：<br><span style="word-break:break-all">` + escapedLink + `</span></p></div></body></html>`,
	}
}

package portal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/wuekevin/axisrelay/database"
	"github.com/wuekevin/axisrelay/internal/testmysql"
)

type recordingMailer struct {
	mu       sync.Mutex
	messages []emailMessage
}

func (m *recordingMailer) Send(_ context.Context, message emailMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	return nil
}

func (m *recordingMailer) last() emailMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages[len(m.messages)-1]
}

func TestEmailVerificationAndResetHTTPFlowMySQL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.New("mysql", testmysql.DSN(t, "portal-handler-email-flow"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mailer := &recordingMailer{}
	handler := newHandlerWithMailer(db, mailer, smtpConfig{PublicURL: "https://relay.example"})
	router := gin.New()
	handler.RegisterRoutes(router)

	register := performJSON(router, http.MethodPost, "/api/auth/register", map[string]any{
		"email": "new.user@example.com", "password": "StrongPass123", "display_name": "新用户", "accept_terms": true,
	}, nil)
	if register.Code != http.StatusAccepted || !strings.Contains(register.Body.String(), `"pending_verification":true`) {
		t.Fatalf("register=%d %s", register.Code, register.Body.String())
	}

	loginBeforeVerify := performJSON(router, http.MethodPost, "/api/auth/login", map[string]any{"email": "new.user@example.com", "password": "StrongPass123"}, nil)
	if loginBeforeVerify.Code != http.StatusForbidden || !strings.Contains(loginBeforeVerify.Body.String(), "email_unverified") {
		t.Fatalf("login before verify=%d %s", loginBeforeVerify.Code, loginBeforeVerify.Body.String())
	}

	verifyToken := tokenFromMessage(t, mailer.last(), "/auth/verify-email")
	verified := performJSON(router, http.MethodPost, "/api/auth/email/verify", map[string]any{"token": verifyToken}, nil)
	if verified.Code != http.StatusOK {
		t.Fatalf("verify=%d %s", verified.Code, verified.Body.String())
	}
	cookies := verified.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName || !cookies[0].HttpOnly {
		t.Fatalf("verification cookie=%#v", cookies)
	}

	reused := performJSON(router, http.MethodPost, "/api/auth/email/verify", map[string]any{"token": verifyToken}, nil)
	if reused.Code != http.StatusBadRequest {
		t.Fatalf("reused verify=%d %s", reused.Code, reused.Body.String())
	}

	known := performJSON(router, http.MethodPost, "/api/auth/password/forgot", map[string]any{"email": "new.user@example.com"}, nil)
	unknown := performJSON(router, http.MethodPost, "/api/auth/password/forgot", map[string]any{"email": "missing@example.com"}, nil)
	if known.Code != http.StatusAccepted || unknown.Code != http.StatusAccepted || known.Body.String() != unknown.Body.String() {
		t.Fatalf("forgot responses differ: known=%q unknown=%q", known.Body.String(), unknown.Body.String())
	}
	resetToken := tokenFromMessage(t, mailer.last(), "/auth/reset-password")
	reset := performJSON(router, http.MethodPost, "/api/auth/password/reset", map[string]any{"token": resetToken, "new_password": "EvenStronger456"}, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset=%d %s", reset.Code, reset.Body.String())
	}

	session := performJSON(router, http.MethodGet, "/api/auth/session", nil, cookies[0])
	if session.Code != http.StatusOK || !strings.Contains(session.Body.String(), `"authenticated":false`) {
		t.Fatalf("old session=%d %s", session.Code, session.Body.String())
	}
	resetReused := performJSON(router, http.MethodPost, "/api/auth/password/reset", map[string]any{"token": resetToken, "new_password": "AnotherPass789"}, nil)
	if resetReused.Code != http.StatusBadRequest {
		t.Fatalf("reused reset=%d %s", resetReused.Code, resetReused.Body.String())
	}
}

func performJSON(handler http.Handler, method, target string, payload any, cookie *http.Cookie) *httptest.ResponseRecorder {
	var body bytes.Buffer
	if payload != nil {
		_ = json.NewEncoder(&body).Encode(payload)
	}
	request := httptest.NewRequest(method, target, &body)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func tokenFromMessage(t *testing.T, message emailMessage, expectedPath string) string {
	t.Helper()
	start := strings.Index(message.TextBody, "https://")
	if start < 0 {
		t.Fatal("email has no action URL")
	}
	end := strings.IndexByte(message.TextBody[start:], '\n')
	if end < 0 {
		end = len(message.TextBody) - start
	}
	link, err := url.Parse(strings.TrimSpace(message.TextBody[start : start+end]))
	if err != nil || link.Path != expectedPath || link.Query().Get("token") == "" {
		t.Fatalf("invalid action link %q: %v", link, err)
	}
	return link.Query().Get("token")
}

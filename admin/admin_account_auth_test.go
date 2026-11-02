package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminAccountSetupLoginAndSession(t *testing.T) {
	db, err := newTestDatabase(t, "admin-account-auth")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := &Handler{db: db}
	router := gin.New()
	router.GET("/api/admin/auth/status", handler.GetAdminAuthStatus)
	router.POST("/api/admin/auth/setup", handler.PostAdminSetup)
	router.POST("/api/admin/auth/login", handler.PostAdminLogin)
	router.POST("/api/admin/auth/password", handler.PostAdminChangePassword)
	router.GET("/api/admin/protected", handler.adminAuthMiddleware(), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	request := func(method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
		var payload *bytes.Reader
		if body == nil {
			payload = bytes.NewReader(nil)
		} else {
			raw, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			payload = bytes.NewReader(raw)
		}
		req := httptest.NewRequest(method, path, payload)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		return recorder
	}

	status := request(http.MethodGet, "/api/admin/auth/status", nil, nil)
	if status.Code != http.StatusOK || !bytes.Contains(status.Body.Bytes(), []byte(`"needs_setup":true`)) {
		t.Fatalf("initial status: %d %s", status.Code, status.Body.String())
	}
	setup := request(http.MethodPost, "/api/admin/auth/setup", map[string]any{"username": "admin", "password": "Strong-admin-password-123", "confirm_password": "Strong-admin-password-123"}, nil)
	if setup.Code != http.StatusCreated {
		t.Fatalf("setup: %d %s", setup.Code, setup.Body.String())
	}
	if len(setup.Result().Cookies()) == 0 || !setup.Result().Cookies()[0].HttpOnly {
		t.Fatal("setup must issue HttpOnly session cookie")
	}
	cookie := setup.Result().Cookies()[0]
	protected := request(http.MethodGet, "/api/admin/protected", nil, cookie)
	if protected.Code != http.StatusOK {
		t.Fatalf("protected: %d %s", protected.Code, protected.Body.String())
	}
	login := request(http.MethodPost, "/api/admin/auth/login", map[string]any{"username": "admin", "password": "Strong-admin-password-123"}, nil)
	if login.Code != http.StatusOK {
		t.Fatalf("login: %d %s", login.Code, login.Body.String())
	}
	changed := request(http.MethodPost, "/api/admin/auth/password", map[string]any{"password": "Different-strong-password-456", "confirm_password": "Different-strong-password-456"}, cookie)
	if changed.Code != http.StatusOK || len(changed.Result().Cookies()) == 0 {
		t.Fatalf("change password: %d %s", changed.Code, changed.Body.String())
	}
	if stale := request(http.MethodGet, "/api/admin/protected", nil, cookie); stale.Code != http.StatusUnauthorized {
		t.Fatalf("old session survived password change: %d %s", stale.Code, stale.Body.String())
	}
	newCookie := changed.Result().Cookies()[0]
	if current := request(http.MethodGet, "/api/admin/protected", nil, newCookie); current.Code != http.StatusOK {
		t.Fatalf("new session rejected: %d %s", current.Code, current.Body.String())
	}
	if oldPassword := request(http.MethodPost, "/api/admin/auth/login", map[string]any{"username": "admin", "password": "Strong-admin-password-123"}, nil); oldPassword.Code != http.StatusUnauthorized {
		t.Fatalf("old password survived change: %d %s", oldPassword.Code, oldPassword.Body.String())
	}
	if newPassword := request(http.MethodPost, "/api/admin/auth/login", map[string]any{"username": "admin", "password": "Different-strong-password-456"}, nil); newPassword.Code != http.StatusOK {
		t.Fatalf("new password rejected: %d %s", newPassword.Code, newPassword.Body.String())
	}
	count, err := db.AdminUserCount(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("admin count=%d err=%v", count, err)
	}
}

func TestAdminLoginRateLimitTracksFailures(t *testing.T) {
	key := t.Name()
	clearAdminLoginFailures(key)
	t.Cleanup(func() { clearAdminLoginFailures(key) })
	if adminLoginRateLimited(key) {
		t.Fatal("fresh login key is rate limited")
	}
	for range 10 {
		recordAdminLoginFailure(key)
	}
	if !adminLoginRateLimited(key) {
		t.Fatal("ten failures must trigger the limit")
	}
	clearAdminLoginFailures(key)
	if adminLoginRateLimited(key) {
		t.Fatal("successful login cleanup did not clear failures")
	}
}

package main

import (
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func validInstallRequest() *installRequest {
	req := &installRequest{Port: 8081, BindAddress: "127.0.0.1"}
	req.Database.Driver = "mysql"
	req.Database.Host = "127.0.0.1"
	req.Database.Port = 3306
	req.Database.User = "axisrelay"
	req.Database.Password = "db-secret"
	req.Database.Name = "axisrelay"
	req.Cache.Driver = "memory"
	req.Admin.Username = "admin"
	req.Admin.Password = "A-strong-password-123"
	req.Admin.ConfirmPassword = req.Admin.Password
	return req
}

func TestValidateInstallRequest(t *testing.T) {
	req := validInstallRequest()
	if err := validateInstallRequest(req); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	req.Admin.ConfirmPassword = "different"
	if err := validateInstallRequest(req); err == nil || !strings.Contains(err.Error(), "不一致") {
		t.Fatalf("mismatch error = %v", err)
	}
}

func TestWriteInstallConfigIsPrivateAndRemovesLegacySecret(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)
	if err := os.WriteFile(".env", []byte("KEEP_ME=yes\nAXISRELAY_ADMIN_SECRET=legacy\n"), 0644); err != nil {
		t.Fatal(err)
	}
	req := validInstallRequest()
	if err := writeInstallConfig(req); err != nil {
		t.Fatalf("writeInstallConfig: %v", err)
	}
	info, err := os.Stat(".env")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf(".env mode = %o, want 600", info.Mode().Perm())
	}
	content, err := os.ReadFile(".env")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if strings.Contains(text, "AXISRELAY_ADMIN_SECRET") {
		t.Fatal("legacy admin secret must be removed")
	}
	parsed, err := godotenv.Read(".env")
	if err != nil {
		t.Fatal(err)
	}
	if parsed["KEEP_ME"] != "yes" || parsed["AXISRELAY_DATABASE_PASSWORD"] != "db-secret" {
		t.Fatalf("unexpected .env content: %s", text)
	}
	if _, err := os.Stat(filepath.Join("data", "install.lock")); err != nil {
		t.Fatalf("install lock: %v", err)
	}
}

func TestSafeInstallErrorRedactsSecrets(t *testing.T) {
	got := safeInstallError(errors.New("access denied for db-password"), "db-password")
	if strings.Contains(got, "db-password") {
		t.Fatalf("secret leaked: %q", got)
	}
}

func TestInstallClientAllowedHonorsForwardedPublicIP(t *testing.T) {
	req := httptest.NewRequest("POST", "http://axisrelay.example/api/install/complete", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	if installClientAllowed(req) {
		t.Fatal("public forwarded client must be denied by default")
	}
	t.Setenv("AXISRELAY_INSTALL_ALLOWED_CIDR", "203.0.113.7/32")
	if !installClientAllowed(req) {
		t.Fatal("explicitly allowed installer client was denied")
	}
}

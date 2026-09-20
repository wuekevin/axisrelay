package security

import (
	"path/filepath"
	"testing"
)

func TestFileLogsDisabled(t *testing.T) {
	t.Setenv("AXISRELAY_LOG_DISABLED", "true")
	if !FileLogsDisabled() {
		t.Fatal("FileLogsDisabled() = false, want true")
	}
}

func TestSecurityLogDirUsesEnv(t *testing.T) {
	t.Setenv("AXISRELAY_LOG_DIR", "/tmp/codex2api-logs")
	if got, want := securityLogDir(), filepath.Join("/tmp/codex2api-logs", "security"); got != want {
		t.Fatalf("securityLogDir() = %q, want %q", got, want)
	}

	t.Setenv("AXISRELAY_SECURITY_LOG_DIR", "/tmp/codex2api-security")
	if got, want := securityLogDir(), "/tmp/codex2api-security"; got != want {
		t.Fatalf("securityLogDir() = %q, want %q", got, want)
	}
}

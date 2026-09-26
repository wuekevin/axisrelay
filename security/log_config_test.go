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
	t.Setenv("AXISRELAY_LOG_DIR", "/tmp/axisrelay-logs")
	if got, want := securityLogDir(), filepath.Join("/tmp/axisrelay-logs", "security"); got != want {
		t.Fatalf("securityLogDir() = %q, want %q", got, want)
	}

	t.Setenv("AXISRELAY_SECURITY_LOG_DIR", "/tmp/axisrelay-security")
	if got, want := securityLogDir(), "/tmp/axisrelay-security"; got != want {
		t.Fatalf("securityLogDir() = %q, want %q", got, want)
	}
}

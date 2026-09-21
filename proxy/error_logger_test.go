package proxy

import "testing"

func TestErrorLogDirUsesEnv(t *testing.T) {
	t.Setenv("AXISRELAY_LOG_DIR", "/tmp/axisrelay-logs")
	if got, want := errorLogDir(), "/tmp/axisrelay-logs"; got != want {
		t.Fatalf("errorLogDir() = %q, want %q", got, want)
	}
}

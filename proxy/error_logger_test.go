package proxy

import "testing"

func TestErrorLogDirUsesEnv(t *testing.T) {
	t.Setenv("AXISRELAY_LOG_DIR", "/tmp/codex2api-logs")
	if got, want := errorLogDir(), "/tmp/codex2api-logs"; got != want {
		t.Fatalf("errorLogDir() = %q, want %q", got, want)
	}
}

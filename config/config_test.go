package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadDefaultsToMySQLAndRedis(t *testing.T) {
	keys := []string{
		"AXISRELAY_PORT",
		"AXISRELAY_MAX_REQUEST_BODY_SIZE_MB",
		"AXISRELAY_ADMIN_SECRET",
		"AXISRELAY_DATABASE_DRIVER",
		"AXISRELAY_DATABASE_PATH",
		"AXISRELAY_DATABASE_HOST",
		"AXISRELAY_DATABASE_PORT",
		"AXISRELAY_DATABASE_USER",
		"AXISRELAY_DATABASE_PASSWORD",
		"AXISRELAY_DATABASE_NAME",
		"AXISRELAY_CACHE_DRIVER",
		"AXISRELAY_REDIS_ADDR",
		"AXISRELAY_REDIS_USERNAME",
		"AXISRELAY_REDIS_PASSWORD",
		"AXISRELAY_REDIS_DB",
		"AXISRELAY_REDIS_TLS",
		"AXISRELAY_REDIS_INSECURE_SKIP_VERIFY",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}

	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}

	if got := cfg.Database.Driver; got != "mysql" {
		t.Fatalf("Database.Driver = %q, want %q", got, "mysql")
	}
	if got := cfg.Cache.Driver; got != "redis" {
		t.Fatalf("Cache.Driver = %q, want %q", got, "redis")
	}
	if got := cfg.Port; got != 8080 {
		t.Fatalf("Port = %d, want %d", got, 8080)
	}
	if got := cfg.MaxRequestBodySize; got != 48*1024*1024 {
		t.Fatalf("MaxRequestBodySize = %d, want %d", got, 48*1024*1024)
	}
	if got := strings.Join(cfg.TrustedProxies, ","); got != "127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16" {
		t.Fatalf("TrustedProxies = %q, want loopback and private-network defaults", got)
	}
}

func TestLoadRejectsSQLiteDatabaseDriver(t *testing.T) {
	keys := []string{
		"AXISRELAY_PORT",
		"AXISRELAY_MAX_REQUEST_BODY_SIZE_MB",
		"AXISRELAY_ADMIN_SECRET",
		"AXISRELAY_DATABASE_DRIVER",
		"AXISRELAY_DATABASE_HOST",
		"AXISRELAY_DATABASE_PORT",
		"AXISRELAY_DATABASE_USER",
		"AXISRELAY_DATABASE_PASSWORD",
		"AXISRELAY_DATABASE_NAME",
		"AXISRELAY_CACHE_DRIVER",
		"AXISRELAY_REDIS_ADDR",
		"AXISRELAY_REDIS_USERNAME",
		"AXISRELAY_REDIS_PASSWORD",
		"AXISRELAY_REDIS_DB",
		"AXISRELAY_REDIS_TLS",
		"AXISRELAY_REDIS_INSECURE_SKIP_VERIFY",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}

	t.Setenv("AXISRELAY_DATABASE_DRIVER", "sqlite")
	t.Setenv("AXISRELAY_CACHE_DRIVER", "memory")

	_, err := Load("__not_exists__.env")
	if err == nil {
		t.Fatal("Load() accepted sqlite database driver, want rejection")
	}
	if !strings.Contains(err.Error(), "仅支持 mysql 或 postgres") {
		t.Fatalf("Load() error = %q, want unsupported database driver error", err)
	}
}

func TestLoadReadsAdminSecretFromEnv(t *testing.T) {
	keys := []string{
		"AXISRELAY_PORT",
		"AXISRELAY_MAX_REQUEST_BODY_SIZE_MB",
		"AXISRELAY_ADMIN_SECRET",
		"AXISRELAY_DATABASE_DRIVER",
		"AXISRELAY_DATABASE_PATH",
		"AXISRELAY_DATABASE_HOST",
		"AXISRELAY_DATABASE_PORT",
		"AXISRELAY_DATABASE_USER",
		"AXISRELAY_DATABASE_PASSWORD",
		"AXISRELAY_DATABASE_NAME",
		"AXISRELAY_CACHE_DRIVER",
		"AXISRELAY_REDIS_ADDR",
		"AXISRELAY_REDIS_USERNAME",
		"AXISRELAY_REDIS_PASSWORD",
		"AXISRELAY_REDIS_DB",
		"AXISRELAY_REDIS_TLS",
		"AXISRELAY_REDIS_INSECURE_SKIP_VERIFY",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}

	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")
	t.Setenv("AXISRELAY_ADMIN_SECRET", "from-env-secret")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}

	if got := cfg.AdminSecret; got != "from-env-secret" {
		t.Fatalf("AdminSecret = %q, want %q", got, "from-env-secret")
	}
}

func TestLoadReadsMaxRequestBodySizeFromEnv(t *testing.T) {
	keys := []string{
		"AXISRELAY_PORT",
		"AXISRELAY_MAX_REQUEST_BODY_SIZE_MB",
		"AXISRELAY_ADMIN_SECRET",
		"AXISRELAY_DATABASE_DRIVER",
		"AXISRELAY_DATABASE_PATH",
		"AXISRELAY_DATABASE_HOST",
		"AXISRELAY_DATABASE_PORT",
		"AXISRELAY_DATABASE_USER",
		"AXISRELAY_DATABASE_PASSWORD",
		"AXISRELAY_DATABASE_NAME",
		"AXISRELAY_CACHE_DRIVER",
		"AXISRELAY_REDIS_ADDR",
		"AXISRELAY_REDIS_USERNAME",
		"AXISRELAY_REDIS_PASSWORD",
		"AXISRELAY_REDIS_DB",
		"AXISRELAY_REDIS_TLS",
		"AXISRELAY_REDIS_INSECURE_SKIP_VERIFY",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}

	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")
	t.Setenv("AXISRELAY_MAX_REQUEST_BODY_SIZE_MB", "64")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}

	if got := cfg.MaxRequestBodySize; got != 64*1024*1024 {
		t.Fatalf("MaxRequestBodySize = %d, want %d", got, 64*1024*1024)
	}
}

func TestLoadParsesTrustedProxiesEnv(t *testing.T) {
	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_CACHE_DRIVER", "")
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")
	t.Setenv("AXISRELAY_TRUSTED_PROXIES", "10.0.0.0/8, 172.16.0.0/12;192.168.1.10")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}
	if got := strings.Join(cfg.TrustedProxies, ","); got != "10.0.0.0/8,172.16.0.0/12,192.168.1.10" {
		t.Fatalf("TrustedProxies = %q", got)
	}
}

func TestLoadCanDisableTrustedProxies(t *testing.T) {
	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_CACHE_DRIVER", "")
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")
	t.Setenv("AXISRELAY_TRUSTED_PROXIES", "none")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}
	if cfg.TrustedProxies != nil {
		t.Fatalf("TrustedProxies = %#v, want nil", cfg.TrustedProxies)
	}
}

func TestLoadDefaultsCodexUpstreamTransportToHTTP(t *testing.T) {
	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_CACHE_DRIVER", "")
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")
	t.Setenv("AXISRELAY_UPSTREAM_TRANSPORT", "")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}
	if got := cfg.CodexUpstreamTransport; got != "http" {
		t.Fatalf("CodexUpstreamTransport = %q, want http", got)
	}
	if cfg.UseWebsocket {
		t.Fatal("UseWebsocket = true, want false")
	}
}

func TestLoadHonorsCodexUpstreamTransportWS(t *testing.T) {
	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_CACHE_DRIVER", "")
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")
	t.Setenv("AXISRELAY_UPSTREAM_TRANSPORT", "websocket")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}
	if got := cfg.CodexUpstreamTransport; got != "ws" {
		t.Fatalf("CodexUpstreamTransport = %q, want ws", got)
	}
	if !cfg.UseWebsocket {
		t.Fatal("UseWebsocket = false, want true")
	}
}

func TestLoadReadsRedisTLSSettings(t *testing.T) {
	keys := []string{
		"AXISRELAY_PORT",
		"AXISRELAY_MAX_REQUEST_BODY_SIZE_MB",
		"AXISRELAY_ADMIN_SECRET",
		"AXISRELAY_DATABASE_DRIVER",
		"AXISRELAY_DATABASE_PATH",
		"AXISRELAY_DATABASE_HOST",
		"AXISRELAY_DATABASE_PORT",
		"AXISRELAY_DATABASE_USER",
		"AXISRELAY_DATABASE_PASSWORD",
		"AXISRELAY_DATABASE_NAME",
		"AXISRELAY_CACHE_DRIVER",
		"AXISRELAY_REDIS_ADDR",
		"AXISRELAY_REDIS_USERNAME",
		"AXISRELAY_REDIS_PASSWORD",
		"AXISRELAY_REDIS_DB",
		"AXISRELAY_REDIS_TLS",
		"AXISRELAY_REDIS_INSECURE_SKIP_VERIFY",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}

	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_REDIS_ADDR", "rediss://default:url-pass@example.upstash.io:6379/2")
	t.Setenv("AXISRELAY_REDIS_USERNAME", "env-user")
	t.Setenv("AXISRELAY_REDIS_PASSWORD", "env-pass")
	t.Setenv("AXISRELAY_REDIS_DB", "3")
	t.Setenv("AXISRELAY_REDIS_TLS", "true")
	t.Setenv("AXISRELAY_REDIS_INSECURE_SKIP_VERIFY", "1")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}

	if got := cfg.Cache.Redis.Addr; got != "rediss://default:url-pass@example.upstash.io:6379/2" {
		t.Fatalf("Redis.Addr = %q, want rediss URL", got)
	}
	if got := cfg.Cache.Redis.Username; got != "env-user" {
		t.Fatalf("Redis.Username = %q, want env-user", got)
	}
	if got := cfg.Cache.Redis.Password; got != "env-pass" {
		t.Fatalf("Redis.Password = %q, want env-pass", got)
	}
	if got := cfg.Cache.Redis.DB; got != 3 {
		t.Fatalf("Redis.DB = %d, want 3", got)
	}
	if !cfg.Cache.Redis.TLS {
		t.Fatal("Redis.TLS = false, want true")
	}
	if !cfg.Cache.Redis.InsecureSkipVerify {
		t.Fatal("Redis.InsecureSkipVerify = false, want true")
	}
}

// issue #498: .env 里的 TZ 必须作用于 time.Local,否则自然日限额按宿主机时区重置。
func TestLoadAppliesTimezoneFromEnvFile(t *testing.T) {
	origLocal := time.Local
	t.Cleanup(func() { time.Local = origLocal })
	t.Setenv("TZ", "") // 注册测试结束后的恢复
	if err := os.Unsetenv("TZ"); err != nil {
		t.Fatalf("Unsetenv(TZ) 失败: %v", err)
	}
	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("TZ=America/Chicago\n"), 0o600); err != nil {
		t.Fatalf("写临时 .env 失败: %v", err)
	}

	if _, err := Load(envPath); err != nil {
		t.Fatalf("Load() 返回错误: %v", err)
	}
	if got := time.Local.String(); got != "America/Chicago" {
		t.Fatalf("time.Local = %q, want %q", got, "America/Chicago")
	}
}

func TestApplyTimezoneHandlesPosixColonPrefix(t *testing.T) {
	origLocal := time.Local
	t.Cleanup(func() { time.Local = origLocal })
	t.Setenv("TZ", ":Asia/Tokyo")

	applyTimezone()
	if got := time.Local.String(); got != "Asia/Tokyo" {
		t.Fatalf("time.Local = %q, want %q", got, "Asia/Tokyo")
	}
}

func TestApplyTimezoneKeepsLocalOnInvalidTZ(t *testing.T) {
	origLocal := time.Local
	t.Cleanup(func() { time.Local = origLocal })
	t.Setenv("TZ", "Not/A_Zone")

	applyTimezone()
	if time.Local != origLocal {
		t.Fatalf("非法 TZ 不应改动 time.Local, got %q", time.Local)
	}
}

func TestLoadIgnoresLegacyEnvironmentAliases(t *testing.T) {
	setTestMySQLConfig(t)
	t.Setenv("AXISRELAY_REDIS_ADDR", "redis:6379")
	t.Setenv("AXISRELAY_PORT", "")
	t.Setenv("AXISRELAY_ADMIN_SECRET", "")
	t.Setenv("AXISRELAY_UPSTREAM_TRANSPORT", "")

	legacyCodexPort := "CODE" + "X_PORT"
	legacyPort := "PO" + "RT"
	legacyAdminSecret := "ADMIN_" + "SECRET"
	legacyUseWebsocket := "USE_" + "WEBSOCKET"
	t.Setenv(legacyCodexPort, "19090")
	t.Setenv(legacyPort, "19091")
	t.Setenv(legacyAdminSecret, "legacy-secret")
	t.Setenv(legacyUseWebsocket, "true")

	cfg, err := Load("__not_exists__.env")
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want default 8080 when legacy port aliases are set", cfg.Port)
	}
	if cfg.AdminSecret != "" {
		t.Fatalf("AdminSecret = %q, want empty when only legacy ADMIN_SECRET is set", cfg.AdminSecret)
	}
	if cfg.CodexUpstreamTransport != "http" || cfg.UseWebsocket {
		t.Fatalf("transport = %q websocket=%v, want http/false when only legacy USE_WEBSOCKET is set", cfg.CodexUpstreamTransport, cfg.UseWebsocket)
	}
}

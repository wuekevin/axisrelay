package config

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	platformmysql "github.com/wuekevin/axisrelay/internal/platform/mysql"
)

// DatabaseConfig 数据库核心配置。
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// DSN 返回当前驱动的连接字符串。
func (d *DatabaseConfig) DSN() string {
	switch strings.ToLower(strings.TrimSpace(d.Driver)) {
	case "mysql":
		dsn, err := (platformmysql.Config{
			Host:     d.Host,
			Port:     d.Port,
			User:     d.User,
			Password: d.Password,
			Database: d.DBName,
		}).DSN()
		if err == nil {
			return dsn
		}
	case "postgres", "postgresql":
		port := d.Port
		if port == 0 {
			port = 5432
		}
		u := &url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(d.User, d.Password),
			Host:   net.JoinHostPort(d.Host, strconv.Itoa(port)),
			Path:   d.DBName,
		}
		return u.String()
	}
	return ""
}

// Label 返回用于展示的数据库标签。
func (d *DatabaseConfig) Label() string {
	switch strings.ToLower(strings.TrimSpace(d.Driver)) {
	case "mysql":
		return "MySQL"
	case "postgres", "postgresql":
		return "PostgreSQL"
	}
	return strings.ToUpper(strings.TrimSpace(d.Driver))
}

// RedisConfig Redis 核心配置
type RedisConfig struct {
	Addr               string
	Username           string
	Password           string
	DB                 int
	TLS                bool
	InsecureSkipVerify bool
}

// CacheConfig 缓存核心配置。
type CacheConfig struct {
	Driver string
	Redis  RedisConfig
}

// Label 返回用于展示的缓存标签。
func (c *CacheConfig) Label() string {
	if strings.EqualFold(c.Driver, "memory") {
		return "Memory"
	}
	return "Redis"
}

// Config 全局核心环境配置（物理隔离的服务器参数）
// 业务逻辑参数（如 ProxyURL，APIKeys，MaxConcurrency）已全部移至数据库 SystemSettings 进行化
type Config struct {
	Port                      int
	BindAddress               string // 监听地址，默认 0.0.0.0（兼容 Docker / 反代 / 公网）；如需仅本机访问可设为 127.0.0.1
	AdminSecret               string
	AllowAnonymousV1          bool // 显式允许 /v1/* 在未配置 API Key 时无鉴权放行（默认禁止）
	APIKeyAuthCacheEnabled    bool
	MaxRequestBodySize        int
	RequestMemoryBudgetBytes  int64 // Process-local retained HTTP/WS request body budget.
	SchedulerMaxWaiters       int   // Process-local waiting request budget.
	SchedulerMaxWaitersPerKey int
	Database                  DatabaseConfig
	Cache                     CacheConfig
	UseWebsocket              bool     // 是否启用 WebSocket 传输
	CodexUpstreamTransport    string   // http|auto|ws，默认 http；仅由 AXISRELAY_UPSTREAM_TRANSPORT 配置
	TrustedProxies            []string // Gin 可信反向代理 CIDR/IP；默认信任回环与私有网段以兼容 Docker 反代，none/off/false/0 表示禁用
}

// applyTimezone 让 TZ 环境变量(含 .env 里的)真正作用于自然日限额等本地时间语义。
// Go 的 time.Local 在进程首次格式化本地时间时就已锁定(main 里 config.Load 之前的
// 启动日志即触发),godotenv 事后 os.Setenv 无法再影响它,必须显式覆盖 time.Local。
func applyTimezone() {
	tz := strings.TrimSpace(os.Getenv("TZ"))
	tz = strings.TrimPrefix(tz, ":") // POSIX 允许 ":Asia/Shanghai" 写法
	if tz == "" {
		return
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("警告: TZ=%q 无法解析(%v),继续使用 %s", tz, err, time.Local)
		return
	}
	time.Local = loc
}

// Load 从 .env 文件加载核心环境配置，支持环境变量覆盖
func Load(envPath string) (*Config, error) {
	// 尝试加载 .env 文件（可选，如果文件不存在则忽略并使用当前环境变量）
	if envPath == "" {
		envPath = ".env"
	}
	_ = godotenv.Load(envPath)
	applyTimezone()

	cfg := &Config{
		Port:               8080,
		MaxRequestBodySize: 48 * 1024 * 1024,
	}

	// Web服务端口
	if port := os.Getenv("AXISRELAY_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Port)
	}
	cfg.AdminSecret = strings.TrimSpace(os.Getenv("AXISRELAY_ADMIN_SECRET"))
	cfg.AllowAnonymousV1 = parseBoolEnv(os.Getenv("AXISRELAY_ALLOW_ANONYMOUS"))
	cfg.APIKeyAuthCacheEnabled = true
	if value := strings.TrimSpace(os.Getenv("AXISRELAY_API_KEY_AUTH_CACHE_ENABLED")); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("AXISRELAY_API_KEY_AUTH_CACHE_ENABLED must be a boolean")
		}
		cfg.APIKeyAuthCacheEnabled = enabled
	}
	// 默认绑 0.0.0.0 以兼容 Docker 端口映射、反向代理、生产服务器等常规部署。
	// 安全防护由 fail-closed 中间件 + 首启自助初始化 (/api/admin/bootstrap) + 启动 banner 共同保证；
	// 想要严格仅本机访问的用户可设 AXISRELAY_BIND=127.0.0.1。
	cfg.BindAddress = strings.TrimSpace(os.Getenv("AXISRELAY_BIND"))
	if cfg.BindAddress == "" {
		cfg.BindAddress = "0.0.0.0"
	}
	if v := strings.TrimSpace(os.Getenv("AXISRELAY_MAX_REQUEST_BODY_SIZE_MB")); v != "" {
		if mb, err := strconv.Atoi(v); err == nil && mb > 0 {
			cfg.MaxRequestBodySize = mb * 1024 * 1024
		}
	}
	cfg.RequestMemoryBudgetBytes = max(int64(128<<20), int64(cfg.MaxRequestBodySize))
	if value := strings.TrimSpace(os.Getenv("AXISRELAY_REQUEST_MEMORY_BUDGET_MB")); value != "" {
		mb, err := strconv.ParseInt(value, 10, 64)
		if err != nil || mb <= 0 || mb > (1<<63-1)/(1<<20) {
			return nil, fmt.Errorf("AXISRELAY_REQUEST_MEMORY_BUDGET_MB must be a positive MiB value")
		}
		cfg.RequestMemoryBudgetBytes = mb << 20
		if cfg.RequestMemoryBudgetBytes < int64(cfg.MaxRequestBodySize) {
			return nil, fmt.Errorf("AXISRELAY_REQUEST_MEMORY_BUDGET_MB must be at least AXISRELAY_MAX_REQUEST_BODY_SIZE_MB")
		}
	}
	cfg.TrustedProxies = parseTrustedProxiesEnv(os.Getenv("AXISRELAY_TRUSTED_PROXIES"))
	for _, setting := range []struct {
		name   string
		target *int
	}{
		{"AXISRELAY_SCHEDULER_MAX_WAITERS", &cfg.SchedulerMaxWaiters},
		{"AXISRELAY_SCHEDULER_MAX_WAITERS_PER_KEY", &cfg.SchedulerMaxWaitersPerKey},
	} {
		if value := strings.TrimSpace(os.Getenv(setting.name)); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("%s must be a positive integer", setting.name)
			}
			*setting.target = n
		}
	}

	// Codex 上游传输仅接受 AxisRelay 命名空间，不读取旧兼容变量。
	cfg.CodexUpstreamTransport = normalizeCodexUpstreamTransport(os.Getenv("AXISRELAY_UPSTREAM_TRANSPORT"))
	if cfg.CodexUpstreamTransport == "" {
		cfg.CodexUpstreamTransport = "http"
	}
	if cfg.CodexUpstreamTransport == "ws" {
		cfg.UseWebsocket = true
	}

	// 数据库配置
	cfg.Database.Driver = normalizeDriver(os.Getenv("AXISRELAY_DATABASE_DRIVER"), "mysql")
	cfg.Database.Host = os.Getenv("AXISRELAY_DATABASE_HOST")
	if v := os.Getenv("AXISRELAY_DATABASE_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = p
		}
	}
	cfg.Database.User = os.Getenv("AXISRELAY_DATABASE_USER")
	cfg.Database.Password = os.Getenv("AXISRELAY_DATABASE_PASSWORD")
	cfg.Database.DBName = os.Getenv("AXISRELAY_DATABASE_NAME")
	if cfg.Database.Port == 0 {
		switch cfg.Database.Driver {
		case "mysql":
			cfg.Database.Port = 3306
		case "postgres", "postgresql":
			cfg.Database.Port = 5432
		}
	}

	// 缓存配置
	cfg.Cache.Driver = normalizeDriver(os.Getenv("AXISRELAY_CACHE_DRIVER"), "redis")
	cfg.Cache.Redis.Addr = strings.TrimSpace(os.Getenv("AXISRELAY_REDIS_ADDR"))
	cfg.Cache.Redis.Username = strings.TrimSpace(os.Getenv("AXISRELAY_REDIS_USERNAME"))
	cfg.Cache.Redis.Password = os.Getenv("AXISRELAY_REDIS_PASSWORD")
	if v := os.Getenv("AXISRELAY_REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			cfg.Cache.Redis.DB = db
		}
	}
	cfg.Cache.Redis.TLS = parseBoolEnv(os.Getenv("AXISRELAY_REDIS_TLS"))
	cfg.Cache.Redis.InsecureSkipVerify = parseBoolEnv(os.Getenv("AXISRELAY_REDIS_INSECURE_SKIP_VERIFY"))

	// 校验必填物理层配置
	switch cfg.Database.Driver {
	case "mysql", "postgres", "postgresql":
		if strings.TrimSpace(cfg.Database.Host) == "" {
			return nil, fmt.Errorf("必须配置数据库主机 (AXISRELAY_DATABASE_HOST)")
		}
		if strings.TrimSpace(cfg.Database.User) == "" {
			return nil, fmt.Errorf("必须配置数据库用户 (AXISRELAY_DATABASE_USER)")
		}
		if strings.TrimSpace(cfg.Database.DBName) == "" {
			return nil, fmt.Errorf("必须配置数据库名称 (AXISRELAY_DATABASE_NAME)")
		}
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s（仅支持 mysql 或 postgres）", cfg.Database.Driver)
	}

	switch cfg.Cache.Driver {
	case "memory":
	case "redis":
		if cfg.Cache.Redis.Addr == "" {
			return nil, fmt.Errorf("必须通过 .env 或环境变量配置 Redis (AXISRELAY_REDIS_ADDR)")
		}
	default:
		return nil, fmt.Errorf("不支持的缓存驱动: %s", cfg.Cache.Driver)
	}

	return cfg, nil
}

func normalizeDriver(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.ToLower(strings.TrimSpace(value))
}

func parseBoolEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func parseTrustedProxiesEnv(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	}
	switch strings.ToLower(value) {
	case "0", "false", "off", "none", "no":
		return nil
	}

	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		if proxy := strings.TrimSpace(part); proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
	return proxies
}

func normalizeCodexUpstreamTransport(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http", "https", "sse":
		return "http"
	case "auto":
		return "auto"
	case "ws", "websocket", "wss":
		return "ws"
	default:
		return ""
	}
}

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/joho/godotenv"
	"github.com/wuekevin/axisrelay/cache"
	"github.com/wuekevin/axisrelay/config"
	"github.com/wuekevin/axisrelay/database"
	"golang.org/x/crypto/bcrypt"
)

const installLockPath = "data/install.lock"

type installRequest struct {
	PublicBaseURL string `json:"public_base_url"`
	Port          int    `json:"port"`
	BindAddress   string `json:"bind_address"`
	Database      struct {
		Driver   string `json:"driver"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	} `json:"database"`
	Cache struct {
		Driver        string `json:"driver"`
		RedisAddr     string `json:"redis_addr"`
		RedisUsername string `json:"redis_username"`
		RedisPassword string `json:"redis_password"`
		RedisDB       int    `json:"redis_db"`
		RedisTLS      bool   `json:"redis_tls"`
	} `json:"cache"`
	Admin struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	} `json:"admin"`
}

func pathExistsOrIsUnreadable(path string) bool {
	_, err := os.Lstat(path)
	return err == nil || !errors.Is(err, os.ErrNotExist)
}

func installationLocked() bool { return pathExistsOrIsUnreadable(installLockPath) }

func existingInstallConfig() bool { return pathExistsOrIsUnreadable(".env") }

func installerPort() int {
	port := 8081
	if value := strings.TrimSpace(os.Getenv("AXISRELAY_PORT")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 && parsed < 65536 {
			port = parsed
		}
	}
	return port
}

func installClientIP(r *http.Request) net.IP {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	peer := net.ParseIP(strings.Trim(host, "[]"))
	if peer != nil && (peer.IsLoopback() || peer.IsPrivate()) {
		if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
			if parsed := net.ParseIP(forwarded); parsed != nil {
				return parsed
			}
		}
		if realIP := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); realIP != nil {
			return realIP
		}
	}
	return peer
}

func installClientAllowed(r *http.Request) bool {
	ip := installClientIP(r)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return true
	}
	for _, raw := range strings.Split(os.Getenv("AXISRELAY_INSTALL_ALLOWED_CIDR"), ",") {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		if !strings.Contains(entry, "/") {
			if parsed := net.ParseIP(entry); parsed != nil && parsed.Equal(ip) {
				return true
			}
			continue
		}
		_, network, err := net.ParseCIDR(entry)
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func validateInstallRequest(req *installRequest) error {
	if req == nil {
		return errors.New("安装参数不能为空")
	}
	req.Database.Driver = strings.ToLower(strings.TrimSpace(req.Database.Driver))
	if req.Database.Driver == "postgresql" {
		req.Database.Driver = "postgres"
	}
	if req.Database.Driver != "mysql" && req.Database.Driver != "postgres" {
		return errors.New("数据库仅支持 MySQL 或 PostgreSQL")
	}
	if strings.TrimSpace(req.Database.Host) == "" || strings.TrimSpace(req.Database.User) == "" || strings.TrimSpace(req.Database.Name) == "" {
		return errors.New("请完整填写数据库主机、账号和数据库名")
	}
	if req.Database.Port == 0 {
		if req.Database.Driver == "mysql" {
			req.Database.Port = 3306
		} else {
			req.Database.Port = 5432
		}
	}
	if req.Database.Port < 1 || req.Database.Port > 65535 {
		return errors.New("数据库端口无效")
	}
	req.Cache.Driver = strings.ToLower(strings.TrimSpace(req.Cache.Driver))
	if req.Cache.Driver == "" {
		req.Cache.Driver = "redis"
	}
	if req.Cache.Driver != "redis" && req.Cache.Driver != "memory" {
		return errors.New("缓存仅支持 Redis 或单机内存")
	}
	if req.Cache.Driver == "redis" && strings.TrimSpace(req.Cache.RedisAddr) == "" {
		return errors.New("请填写 Redis 地址")
	}
	username := strings.TrimSpace(req.Admin.Username)
	if !database.ValidAdminUsername(username) {
		return errors.New("管理员账号需为 3 至 64 位字母、数字、点、下划线或连字符")
	}
	if len(req.Admin.Password) < 12 {
		return errors.New("管理员密码至少 12 位")
	}
	if len(req.Admin.Password) > 72 {
		return errors.New("管理员密码不能超过 72 字节")
	}
	if req.Admin.Password != req.Admin.ConfirmPassword {
		return errors.New("两次输入的管理员密码不一致")
	}
	if req.Port == 0 {
		req.Port = installerPort()
	}
	if req.Port < 1 || req.Port > 65535 {
		return errors.New("服务端口无效")
	}
	if strings.TrimSpace(req.BindAddress) == "" {
		req.BindAddress = "127.0.0.1"
	}
	if net.ParseIP(strings.TrimSpace(req.BindAddress)) == nil {
		return errors.New("监听地址必须是有效 IP")
	}
	if req.PublicBaseURL != "" {
		parsed, err := url.ParseRequestURI(strings.TrimSpace(req.PublicBaseURL))
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return errors.New("站点公开地址必须是完整的 http 或 https 地址")
		}
	}
	return nil
}

func safeInstallError(err error, secrets ...string) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	for _, secret := range secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[REDACTED]")
		}
	}
	return message
}

func verifyAndInitialize(req *installRequest) error {
	dbCfg := config.DatabaseConfig{Driver: req.Database.Driver, Host: strings.TrimSpace(req.Database.Host), Port: req.Database.Port, User: strings.TrimSpace(req.Database.User), Password: req.Database.Password, DBName: strings.TrimSpace(req.Database.Name)}
	db, err := database.New(dbCfg.Driver, dbCfg.DSN())
	if err != nil {
		return fmt.Errorf("数据库连接或迁移失败（%s）: %s", dbCfg.SafeTarget(), safeInstallError(err, req.Database.Password, dbCfg.DSN()))
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	count, err := db.AdminUserCount(ctx)
	if err != nil {
		return fmt.Errorf("读取管理员状态失败: %w", err)
	}
	if count > 0 {
		return errors.New("目标数据库已存在管理员账号，安装程序不会覆盖；请改用原配置启动")
	}
	if req.Cache.Driver == "redis" {
		client, err := cache.NewRedisWithOptions(cache.RedisOptions{Addr: strings.TrimSpace(req.Cache.RedisAddr), Username: strings.TrimSpace(req.Cache.RedisUsername), Password: req.Cache.RedisPassword, DB: req.Cache.RedisDB, TLS: req.Cache.RedisTLS, PoolSize: 2})
		if err != nil {
			return fmt.Errorf("Redis 配置无效: %s", safeInstallError(err, req.Cache.RedisPassword))
		}
		defer client.Close()
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pingCancel()
		if err := client.Ping(pingCtx); err != nil {
			return fmt.Errorf("Redis 连接失败: %s", safeInstallError(err, req.Cache.RedisPassword))
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Admin.Password), 12)
	if err != nil {
		return errors.New("管理员密码保护失败")
	}
	user, err := db.CreateAdminUser(ctx, req.Admin.Username, string(hash), false)
	if err != nil {
		return fmt.Errorf("创建管理员账号失败: %w", err)
	}
	if err := writeInstallConfig(req); err != nil {
		_ = db.DeleteAdminUser(ctx, user.ID)
		return fmt.Errorf("配置写入失败，请检查程序目录权限: %w", err)
	}
	return nil
}

func installEnvValues(req *installRequest) map[string]string {
	return map[string]string{
		"AXISRELAY_PORT": strconv.Itoa(req.Port), "AXISRELAY_BIND": strings.TrimSpace(req.BindAddress), "AXISRELAY_PUBLIC_BASE_URL": strings.TrimSpace(req.PublicBaseURL),
		"AXISRELAY_DATABASE_DRIVER": req.Database.Driver, "AXISRELAY_DATABASE_HOST": strings.TrimSpace(req.Database.Host), "AXISRELAY_DATABASE_PORT": strconv.Itoa(req.Database.Port), "AXISRELAY_DATABASE_USER": strings.TrimSpace(req.Database.User), "AXISRELAY_DATABASE_PASSWORD": req.Database.Password, "AXISRELAY_DATABASE_NAME": strings.TrimSpace(req.Database.Name),
		"AXISRELAY_CACHE_DRIVER": req.Cache.Driver, "AXISRELAY_REDIS_ADDR": strings.TrimSpace(req.Cache.RedisAddr), "AXISRELAY_REDIS_USERNAME": strings.TrimSpace(req.Cache.RedisUsername), "AXISRELAY_REDIS_PASSWORD": req.Cache.RedisPassword, "AXISRELAY_REDIS_DB": strconv.Itoa(req.Cache.RedisDB), "AXISRELAY_REDIS_TLS": strconv.FormatBool(req.Cache.RedisTLS),
	}
}

func writeInstallConfig(req *installRequest) error {
	if info, err := os.Lstat(".env"); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("拒绝写入符号链接 .env")
	}
	values := map[string]string{}
	if existing, err := godotenv.Read(".env"); err == nil {
		for key, value := range existing {
			values[key] = value
		}
	}
	delete(values, "AXISRELAY_ADMIN_SECRET")
	for key, value := range installEnvValues(req) {
		values[key] = value
	}
	content, err := godotenv.Marshal(values)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(".", ".axisrelay-env-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(content + "\n"); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, ".env"); err != nil {
		return err
	}
	if err := os.Chmod(".env", 0600); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(installLockPath), 0750); err != nil {
		return err
	}
	lockTmp, err := os.CreateTemp(filepath.Dir(installLockPath), ".install-lock-*")
	if err != nil {
		return err
	}
	lockName := lockTmp.Name()
	defer os.Remove(lockName)
	_ = lockTmp.Chmod(0600)
	_, _ = fmt.Fprintf(lockTmp, "installed_at=%s\n", time.Now().UTC().Format(time.RFC3339))
	if err := lockTmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(lockName, installLockPath); err != nil {
		return err
	}
	// The installer form is the explicit source of truth for this first boot.
	// Override values that may have been loaded from an incomplete .env earlier
	// in the same process; no secret values are logged.
	for key, value := range installEnvValues(req) {
		_ = os.Setenv(key, value)
	}
	_ = os.Unsetenv("AXISRELAY_ADMIN_SECRET")
	return nil
}

func runFirstInstall() error {
	webFS, err := fs.Sub(frontendFS, "frontend/dist/web")
	if err != nil {
		return fmt.Errorf("加载安装页面失败: %w", err)
	}
	fileServer := http.FileServer(http.FS(webFS))
	mux := http.NewServeMux()
	var installing atomic.Bool
	completed := make(chan struct{}, 1)
	mux.HandleFunc("/api/install/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{"installed": false, "ready": true, "default_port": installerPort()})
	})
	mux.HandleFunc("/api/install/complete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !installClientAllowed(r) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "当前来源不允许执行首次安装；请从本机、内网访问，或临时设置 AXISRELAY_INSTALL_ALLOWED_CIDR 放行你的 IP"})
			return
		}
		if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
			scheme := "http"
			if strings.EqualFold(strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]), "https") {
				scheme = "https"
			}
			if !strings.EqualFold(origin, scheme+"://"+r.Host) {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "安装请求来源不受信任"})
				return
			}
		}
		if installationLocked() {
			http.Error(w, `{"error":"系统已经安装"}`, http.StatusConflict)
			return
		}
		if !installing.CompareAndSwap(false, true) {
			http.Error(w, `{"error":"安装正在进行"}`, http.StatusConflict)
			return
		}
		defer installing.Store(false)
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		var req installRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "安装参数格式无效"})
			return
		}
		if err := validateInstallRequest(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if err := verifyAndInitialize(&req); err != nil {
			log.Printf("安装验证失败: %s", safeInstallError(err, req.Database.Password, req.Cache.RedisPassword))
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": safeInstallError(err, req.Database.Password, req.Cache.RedisPassword)})
			return
		}
		redirectURL := strings.TrimRight(strings.TrimSpace(req.PublicBaseURL), "/")
		if redirectURL == "" {
			redirectURL = "/"
		} else {
			redirectURL += "/"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "message": "安装完成，服务正在启动", "redirect_url": redirectURL})
		select {
		case completed <- struct{}{}:
		default:
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/install", http.StatusTemporaryRedirect)
			return
		}
		path := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		if path != "." {
			if _, statErr := fs.Stat(webFS, path); statErr == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		if strings.HasPrefix(path, "assets/") {
			http.NotFound(w, r)
			return
		}
		index, readErr := fs.ReadFile(webFS, "index.html")
		if readErr != nil {
			http.Error(w, "installer unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(index)
	})
	address := net.JoinHostPort("0.0.0.0", strconv.Itoa(installerPort()))
	secureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		// The bundled Web entry contains a small inline theme bootstrap. No user
		// content is rendered into it; allow that script while keeping all network,
		// framing and form targets same-origin during installation.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		mux.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: address, Handler: secureHandler, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 10 * time.Minute, IdleTimeout: 60 * time.Second}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	log.Printf("首次安装页面已启动: http://127.0.0.1:%d/install（宝塔反向代理可指向此端口）", installerPort())
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-completed:
		time.Sleep(350 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return err
		}
		<-errCh
	}
	return nil
}

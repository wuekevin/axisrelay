package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wuekevin/axisrelay/database"
	"golang.org/x/crypto/bcrypt"
)

const adminSessionCookie = "axisrelay_admin_session"
const adminSessionCookiePath = "/api/admin"

var adminLoginAttempts = struct {
	sync.Mutex
	items map[string][]time.Time
}{items: make(map[string][]time.Time)}

func adminLoginRateLimited(key string) bool {
	now := time.Now()
	cutoff := now.Add(-15 * time.Minute)
	adminLoginAttempts.Lock()
	defer adminLoginAttempts.Unlock()
	if len(adminLoginAttempts.items) > 10000 {
		for itemKey, itemEntries := range adminLoginAttempts.items {
			active := itemEntries[:0]
			for _, entry := range itemEntries {
				if entry.After(cutoff) {
					active = append(active, entry)
				}
			}
			if len(active) == 0 {
				delete(adminLoginAttempts.items, itemKey)
			} else {
				adminLoginAttempts.items[itemKey] = active
			}
		}
	}
	entries := adminLoginAttempts.items[key]
	kept := entries[:0]
	for _, entry := range entries {
		if entry.After(cutoff) {
			kept = append(kept, entry)
		}
	}
	if len(kept) >= 10 {
		adminLoginAttempts.items[key] = kept
		return true
	}
	adminLoginAttempts.items[key] = kept
	return false
}

func recordAdminLoginFailure(key string) {
	adminLoginAttempts.Lock()
	adminLoginAttempts.items[key] = append(adminLoginAttempts.items[key], time.Now())
	adminLoginAttempts.Unlock()
}

func clearAdminLoginFailures(key string) {
	adminLoginAttempts.Lock()
	delete(adminLoginAttempts.items, key)
	adminLoginAttempts.Unlock()
}

func requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]), "https")
}

func setAdminSessionCookie(c *gin.Context, token string, expires time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{Name: adminSessionCookie, Value: token, Path: adminSessionCookiePath, HttpOnly: true, Secure: requestIsHTTPS(c.Request), SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: int(time.Until(expires).Seconds())})
}

func clearAdminSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: adminSessionCookie, Value: "", Path: adminSessionCookiePath, HttpOnly: true, Secure: requestIsHTTPS(c.Request), SameSite: http.SameSiteStrictMode, Expires: time.Unix(1, 0), MaxAge: -1})
}

func disableAdminAuthCaching(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}

func newAdminSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (h *Handler) currentAdmin(c *gin.Context) (*database.AdminUser, error) {
	cookie, err := c.Request.Cookie(adminSessionCookie)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return nil, database.ErrAdminAuthNotFound
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	return h.db.AdminUserBySession(ctx, database.HashAdminSessionToken(cookie.Value), time.Now())
}

func (h *Handler) GetAdminAuthStatus(c *gin.Context) {
	disableAdminAuthCaching(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	count, err := h.db.AdminUserCount(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "无法读取管理员账号状态"})
		return
	}
	user, authErr := h.currentAdmin(c)
	if authErr == nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": true, "needs_setup": false, "user": gin.H{"username": user.Username, "must_change_password": user.MustChangePassword}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"authenticated": false, "needs_setup": count == 0})
}

func (h *Handler) PostAdminLogin(c *gin.Context) {
	disableAdminAuthCaching(c)
	if !adminMutationOriginAllowed(c.Request) {
		c.JSON(http.StatusForbidden, gin.H{"error": "登录请求来源不受信任"})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Remember bool   `json:"remember"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入账号和密码"})
		return
	}
	loginKey := c.ClientIP() + "|" + database.NormalizeAdminUsername(body.Username)
	if adminLoginRateLimited(loginKey) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "登录尝试过于频繁，请稍后再试"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	user, err := h.db.FindAdminUserByUsername(ctx, body.Username)
	if err != nil || user.Status != "active" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)) != nil {
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$q4bZxQh2KjT5UF3z5mfcZuJfTxJ8M6ChQy.YwDPoFfr9hKye2pA/G"), []byte(body.Password))
		recordAdminLoginFailure(loginKey)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}
	clearAdminLoginFailures(loginKey)
	token, err := newAdminSessionToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建安全会话"})
		return
	}
	duration := 12 * time.Hour
	if body.Remember {
		duration = 30 * 24 * time.Hour
	}
	expires := time.Now().Add(duration)
	if err := h.db.CreateAdminSession(ctx, user.ID, database.HashAdminSessionToken(token), c.ClientIP(), c.GetHeader("User-Agent"), expires); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建安全会话"})
		return
	}
	_ = h.db.MarkAdminLogin(ctx, user.ID)
	setAdminSessionCookie(c, token, expires)
	c.JSON(http.StatusOK, gin.H{"ok": true, "username": user.Username, "must_change_password": user.MustChangePassword})
}

func (h *Handler) PostAdminLogout(c *gin.Context) {
	disableAdminAuthCaching(c)
	if !adminMutationOriginAllowed(c.Request) {
		c.JSON(http.StatusForbidden, gin.H{"error": "退出请求来源不受信任"})
		return
	}
	if cookie, err := c.Request.Cookie(adminSessionCookie); err == nil && cookie.Value != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		_ = h.db.RevokeAdminSession(ctx, database.HashAdminSessionToken(cookie.Value))
	}
	clearAdminSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) PostAdminChangePassword(c *gin.Context) {
	disableAdminAuthCaching(c)
	if !adminMutationOriginAllowed(c.Request) {
		c.JSON(http.StatusForbidden, gin.H{"error": "密码修改请求来源不受信任"})
		return
	}
	user, err := h.currentAdmin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录管理员账号"})
		return
	}
	var body struct {
		Password string `json:"password"`
		Confirm  string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式无效"})
		return
	}
	if len(body.Password) < 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管理员密码至少 12 位"})
		return
	}
	if len(body.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管理员密码不能超过 72 字节"})
		return
	}
	if body.Password != body.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)) == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码不能与原密码相同"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法保护管理员密码"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	if err := h.db.ChangeAdminPassword(ctx, user.ID, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码更新失败"})
		return
	}
	token, err := newAdminSessionToken()
	if err != nil {
		clearAdminSessionCookie(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码已更新，请重新登录"})
		return
	}
	expires := time.Now().Add(12 * time.Hour)
	if err := h.db.CreateAdminSession(ctx, user.ID, database.HashAdminSessionToken(token), c.ClientIP(), c.GetHeader("User-Agent"), expires); err != nil {
		clearAdminSessionCookie(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码已更新，请重新登录"})
		return
	}
	setAdminSessionCookie(c, token, expires)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) PostAdminSetup(c *gin.Context) {
	disableAdminAuthCaching(c)
	if !adminMutationOriginAllowed(c.Request) {
		c.JSON(http.StatusForbidden, gin.H{"error": "初始化请求来源不受信任"})
		return
	}
	if !bootstrapAllowClientIP(c.ClientIP()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "首次管理员创建仅允许从本机或可信私有网络执行"})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Confirm  string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式无效"})
		return
	}
	username := strings.TrimSpace(body.Username)
	if !database.ValidAdminUsername(username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管理员账号需为 3 至 64 位字母、数字、点、下划线或连字符"})
		return
	}
	if len(body.Password) < 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管理员密码至少 12 位"})
		return
	}
	if len(body.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管理员密码不能超过 72 字节"})
		return
	}
	if body.Password != body.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	count, err := h.db.AdminUserCount(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "无法确认管理员状态"})
		return
	}
	if count != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "管理员账号已经创建"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法保护管理员密码"})
		return
	}
	user, err := h.db.CreateAdminUser(ctx, username, string(hash), false)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "管理员创建失败"})
		return
	}
	token, err := newAdminSessionToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建安全会话"})
		return
	}
	expires := time.Now().Add(12 * time.Hour)
	if err := h.db.CreateAdminSession(ctx, user.ID, database.HashAdminSessionToken(token), c.ClientIP(), c.GetHeader("User-Agent"), expires); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建安全会话"})
		return
	}
	setAdminSessionCookie(c, token, expires)
	c.JSON(http.StatusCreated, gin.H{"ok": true, "username": user.Username})
}

func adminMutationOriginAllowed(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return true
	}
	if site := strings.ToLower(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site"))); site != "" && site != "same-origin" && site != "none" {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	scheme := "http"
	if requestIsHTTPS(r) {
		scheme = "https"
	}
	return strings.EqualFold(origin, scheme+"://"+r.Host)
}

func isAdminAuthNotFound(err error) bool { return errors.Is(err, database.ErrAdminAuthNotFound) }

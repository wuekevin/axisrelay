package portal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/wuekevin/axisrelay/database"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName  = "axisrelay_session"
	defaultSessionTTL  = 7 * 24 * time.Hour
	rememberSessionTTL = 30 * 24 * time.Hour
	verificationTTL    = 24 * time.Hour
	passwordResetTTL   = 30 * time.Minute
	maxUserAPIKeys     = 10
)

var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("AxisRelay-Invalid-Password-9"), 12)
	if err != nil {
		panic(err)
	}
	return hash
}()

type Handler struct {
	db            *database.DB
	limiter       *attemptLimiter
	signupEnabled bool
	defaultQuota  float64
	mailer        emailSender
	mailConfig    smtpConfig
}

type authContext struct {
	User      *database.PortalUser
	SessionID int64
	TokenHash string
}

const authContextKey = "axisrelay_portal_auth"

func NewHandler(db *database.DB) *Handler {
	handler := &Handler{
		db:            db,
		limiter:       newAttemptLimiter(),
		signupEnabled: envBoolDefault("AXISRELAY_PUBLIC_SIGNUP_ENABLED", true),
		defaultQuota:  envFloatDefault("AXISRELAY_PUBLIC_USER_DEFAULT_QUOTA_USD", 1),
	}
	config, err := smtpConfigFromEnv()
	if err != nil {
		if handler.signupEnabled {
			log.Printf("公共注册邮件服务尚未就绪: %v", err)
		}
		return handler
	}
	handler.mailConfig = config
	handler.mailer = &smtpSender{config: config}
	return handler
}

func newHandlerWithMailer(db *database.DB, sender emailSender, config smtpConfig) *Handler {
	return &Handler{db: db, limiter: newAttemptLimiter(), signupEnabled: true, defaultQuota: 1, mailer: sender, mailConfig: config}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	auth.Use(noStore(), h.sameOrigin())
	auth.GET("/session", h.getSession)
	auth.POST("/register", h.register)
	auth.POST("/login", h.login)
	auth.POST("/logout", h.logout)
	auth.POST("/email/verify", h.verifyEmail)
	auth.POST("/email/resend", h.resendVerification)
	auth.POST("/password/forgot", h.forgotPassword)
	auth.POST("/password/reset", h.resetPassword)

	user := r.Group("/api/user")
	user.Use(noStore(), h.sameOrigin(), h.requireSession())
	user.GET("/profile", h.getProfile)
	user.PATCH("/profile", h.updateProfile)
	user.PUT("/password", h.changePassword)
	user.GET("/sessions", h.listSessions)
	user.DELETE("/sessions/:id", h.revokeSession)
	user.POST("/sessions/revoke-others", h.revokeOtherSessions)
	user.GET("/api-keys", h.listAPIKeys)
	user.POST("/api-keys", h.createAPIKey)
	user.PATCH("/api-keys/:id", h.renameAPIKey)
	user.DELETE("/api-keys/:id", h.revokeAPIKey)
	user.GET("/api-keys/:id/usage", h.getAPIKeyUsage)
	user.GET("/subscription", h.getSubscription)
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	AcceptTerms bool   `json:"accept_terms"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func (h *Handler) register(c *gin.Context) {
	if !h.signupEnabled {
		writeError(c, http.StatusForbidden, "signup_disabled", "当前站点暂未开放注册")
		return
	}
	if h.mailer == nil {
		writeError(c, http.StatusServiceUnavailable, "email_service_unavailable", "邮件服务尚未配置，暂时无法注册，请稍后再试")
		return
	}
	if !h.limiter.Allow("register:"+c.ClientIP(), 5, 15*time.Minute) {
		c.Header("Retry-After", "900")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "注册请求过于频繁，请稍后再试")
		return
	}
	var request registerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	email, err := normalizeEmail(request.Email)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_email", "请输入有效的邮箱地址")
		return
	}
	if !h.limiter.Allow("register:email:"+email, 3, time.Hour) {
		c.Header("Retry-After", "3600")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "该邮箱的注册请求过于频繁，请稍后再试")
		return
	}
	if message := validatePassword(request.Password); message != "" {
		writeError(c, http.StatusBadRequest, "weak_password", message)
		return
	}
	displayName := strings.TrimSpace(request.DisplayName)
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}
	if utf8.RuneCountInString(displayName) < 2 || utf8.RuneCountInString(displayName) > 64 {
		writeError(c, http.StatusBadRequest, "invalid_display_name", "昵称长度需要在 2 到 64 个字符之间")
		return
	}
	if !request.AcceptTerms {
		writeError(c, http.StatusBadRequest, "terms_required", "请先同意服务条款和隐私政策")
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 12)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法创建账号")
		return
	}
	verificationToken, verificationHash, err := newSecret(32)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法创建账号")
		return
	}
	publicID, _, err := newSecret(16)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法创建账号")
		return
	}
	publicID = "usr_" + publicID
	if len(publicID) > 26 {
		publicID = publicID[:26]
	}
	user, err := h.db.CreatePendingPortalUser(c.Request.Context(), database.CreatePendingPortalUserInput{
		PublicID:        publicID,
		Email:           email,
		EmailNormalized: email,
		PasswordHash:    string(passwordHash),
		DisplayName:     displayName,
		Locale:          preferredLocale(c),
		Timezone:        "Asia/Shanghai",
		TokenHash:       verificationHash,
		TokenExpires:    time.Now().UTC().Add(verificationTTL),
	})
	if errors.Is(err, database.ErrPortalEmailExists) {
		writeError(c, http.StatusConflict, "email_exists", "该邮箱已注册，请直接登录")
		return
	}
	if err != nil || user == nil {
		writeError(c, http.StatusInternalServerError, "server_error", "账号创建失败，请稍后重试")
		return
	}
	mailContext, cancel := contextWithMailTimeout(c.Request.Context())
	defer cancel()
	if err := h.mailer.Send(mailContext, verificationMessage(h.mailConfig, email, verificationToken)); err != nil {
		log.Printf("注册验证邮件投递失败: %v", err)
		writeError(c, http.StatusBadGateway, "email_delivery_failed", "账号已创建，但验证邮件暂未送达，请稍后重新发送")
		return
	}
	h.limiter.Reset("register:" + c.ClientIP())
	c.JSON(http.StatusAccepted, gin.H{"authenticated": false, "pending_verification": true, "email": user.Email})
}

func (h *Handler) login(c *gin.Context) {
	key := "login:" + c.ClientIP()
	if !h.limiter.Allow(key, 10, 15*time.Minute) {
		c.Header("Retry-After", "900")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "登录尝试过于频繁，请 15 分钟后再试")
		return
	}
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	email, err := normalizeEmail(request.Email)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "邮箱或密码不正确")
		return
	}
	user, passwordHash, lookupErr := h.db.GetPortalUserCredentialsByEmail(c.Request.Context(), email)
	if lookupErr != nil || user == nil {
		// Keep the missing-account path close to the wrong-password path so the
		// endpoint does not disclose registered addresses through response time.
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(request.Password))
		h.db.RecordPortalLoginFailure(c.Request.Context(), nil, email, "invalid_credentials", trimString(c.ClientIP(), 45), trimString(c.Request.UserAgent(), 512))
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "邮箱或密码不正确")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)) != nil {
		h.db.RecordPortalLoginFailure(c.Request.Context(), &user.ID, email, "invalid_credentials", trimString(c.ClientIP(), 45), trimString(c.Request.UserAgent(), 512))
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "邮箱或密码不正确")
		return
	}
	if user.Status == "pending_verification" || user.EmailVerifiedAt == nil {
		h.db.RecordPortalLoginFailure(c.Request.Context(), &user.ID, email, "email_unverified", trimString(c.ClientIP(), 45), trimString(c.Request.UserAgent(), 512))
		writeError(c, http.StatusForbidden, "email_unverified", "请先完成邮箱验证后再登录")
		return
	}
	if user.Status != "active" {
		h.db.RecordPortalLoginFailure(c.Request.Context(), &user.ID, email, "account_inactive", trimString(c.ClientIP(), 45), trimString(c.Request.UserAgent(), 512))
		writeError(c, http.StatusForbidden, "account_inactive", "账号当前不可用，请稍后再试")
		return
	}
	token, tokenHash, err := newSecret(32)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法创建会话")
		return
	}
	ttl := defaultSessionTTL
	if request.Remember {
		ttl = rememberSessionTTL
	}
	expiresAt := time.Now().UTC().Add(ttl)
	if _, err := h.db.CreatePortalSession(c.Request.Context(), user.ID, email, tokenHash, trimString(c.ClientIP(), 45), trimString(c.Request.UserAgent(), 512), expiresAt); err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "登录失败，请稍后重试")
		return
	}
	h.setSessionCookie(c, token, expiresAt)
	h.limiter.Reset(key)
	c.JSON(http.StatusOK, gin.H{"authenticated": true, "user": user})
}

type emailRequest struct {
	Email string `json:"email"`
}

type tokenRequest struct {
	Token string `json:"token"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) verifyEmail(c *gin.Context) {
	var request tokenRequest
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Token) == "" {
		writeError(c, http.StatusBadRequest, "invalid_token", "验证链接无效或已过期")
		return
	}
	if !h.limiter.Allow("verify:"+c.ClientIP(), 12, 15*time.Minute) {
		c.Header("Retry-After", "900")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "验证尝试过于频繁，请稍后再试")
		return
	}
	tokenHash := hashSecret(strings.TrimSpace(request.Token))
	if !h.limiter.Allow("verify:token:"+tokenHash, 5, 15*time.Minute) {
		c.Header("Retry-After", "900")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "验证尝试过于频繁，请稍后再试")
		return
	}
	sessionToken, sessionHash, err := newSecret(32)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法完成验证")
		return
	}
	expiresAt := time.Now().UTC().Add(defaultSessionTTL)
	user, err := h.db.ConsumePortalEmailVerification(c.Request.Context(), tokenHash, sessionHash, trimString(c.ClientIP(), 45), trimString(c.Request.UserAgent(), 512), expiresAt)
	if errors.Is(err, database.ErrPortalTokenInvalid) {
		writeError(c, http.StatusBadRequest, "invalid_token", "验证链接无效或已过期")
		return
	}
	if err != nil || user == nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法完成验证")
		return
	}
	h.setSessionCookie(c, sessionToken, expiresAt)
	h.limiter.Reset("verify:" + c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"authenticated": true, "user": user})
}

func (h *Handler) resendVerification(c *gin.Context) {
	if h.mailer == nil {
		writeError(c, http.StatusServiceUnavailable, "email_service_unavailable", "邮件服务暂时不可用，请稍后再试")
		return
	}
	var request emailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	email, err := normalizeEmail(request.Email)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_email", "请输入有效的邮箱地址")
		return
	}
	if !h.allowEmailAction("resend", c.ClientIP(), email, 5, time.Hour) {
		c.Header("Retry-After", "3600")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "发送请求过于频繁，请稍后再试")
		return
	}
	user, _, lookupErr := h.db.GetPortalUserCredentialsByEmail(c.Request.Context(), email)
	if lookupErr != nil || user == nil || user.Status != "pending_verification" || user.EmailVerifiedAt != nil {
		c.JSON(http.StatusAccepted, gin.H{"message": "如果该邮箱正在等待验证，我们会发送一封新邮件"})
		return
	}
	token, tokenHash, err := newSecret(32)
	if err != nil || h.db.ReplacePortalEmailToken(c.Request.Context(), user.ID, user.Email, "verify_email", tokenHash, time.Now().UTC().Add(verificationTTL)) != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法发送验证邮件")
		return
	}
	mailContext, cancel := contextWithMailTimeout(c.Request.Context())
	defer cancel()
	if err := h.mailer.Send(mailContext, verificationMessage(h.mailConfig, user.Email, token)); err != nil {
		log.Printf("验证邮件重发失败: %v", err)
		writeError(c, http.StatusBadGateway, "email_delivery_failed", "验证邮件暂未送达，请稍后再试")
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "如果该邮箱正在等待验证，我们会发送一封新邮件"})
}

func (h *Handler) forgotPassword(c *gin.Context) {
	const genericMessage = "如果该邮箱已注册，我们会发送一封密码重置邮件"
	var request emailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	email, err := normalizeEmail(request.Email)
	if err != nil {
		c.JSON(http.StatusAccepted, gin.H{"message": genericMessage})
		return
	}
	if !h.allowEmailAction("forgot", c.ClientIP(), email, 5, time.Hour) {
		c.Header("Retry-After", "3600")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "请求过于频繁，请稍后再试")
		return
	}
	if h.mailer != nil {
		user, _, lookupErr := h.db.GetPortalUserCredentialsByEmail(c.Request.Context(), email)
		if lookupErr == nil && user != nil && user.Status == "active" && user.EmailVerifiedAt != nil {
			token, tokenHash, secretErr := newSecret(32)
			if secretErr == nil && h.db.ReplacePortalEmailToken(c.Request.Context(), user.ID, user.Email, "reset_password", tokenHash, time.Now().UTC().Add(passwordResetTTL)) == nil {
				mailContext, cancel := contextWithMailTimeout(c.Request.Context())
				defer cancel()
				if sendErr := h.mailer.Send(mailContext, passwordResetMessage(h.mailConfig, user.Email, token)); sendErr != nil {
					log.Printf("密码重置邮件投递失败: %v", sendErr)
				}
			}
		}
	}
	c.JSON(http.StatusAccepted, gin.H{"message": genericMessage})
}

func (h *Handler) resetPassword(c *gin.Context) {
	if !h.limiter.Allow("password-reset:"+c.ClientIP(), 10, 15*time.Minute) {
		c.Header("Retry-After", "900")
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "重置尝试过于频繁，请稍后再试")
		return
	}
	var request resetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Token) == "" {
		writeError(c, http.StatusBadRequest, "invalid_token", "重置链接无效或已过期")
		return
	}
	if message := validatePassword(request.NewPassword); message != "" {
		writeError(c, http.StatusBadRequest, "weak_password", message)
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), 12)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法重置密码")
		return
	}
	if err = h.db.ConsumePortalPasswordReset(c.Request.Context(), hashSecret(strings.TrimSpace(request.Token)), string(passwordHash)); errors.Is(err, database.ErrPortalTokenInvalid) {
		writeError(c, http.StatusBadRequest, "invalid_token", "重置链接无效或已过期")
		return
	} else if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "暂时无法重置密码")
		return
	}
	h.clearSessionCookie(c)
	h.limiter.Reset("password-reset:" + c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "密码已重置，请使用新密码登录"})
}

func (h *Handler) allowEmailAction(action, ip, email string, limit int, window time.Duration) bool {
	return h.limiter.Allow(action+":ip:"+ip, limit, window) && h.limiter.Allow(action+":email:"+email, limit, window)
}

func contextWithMailTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 12*time.Second)
}

func (h *Handler) logout(c *gin.Context) {
	if token, err := c.Cookie(sessionCookieName); err == nil && token != "" {
		_ = h.db.RevokePortalSessionByHash(c.Request.Context(), hashSecret(token))
	}
	h.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) getSession(c *gin.Context) {
	auth, err := h.resolveSession(c)
	if err != nil || auth == nil {
		if token, cookieErr := c.Cookie(sessionCookieName); cookieErr == nil && token != "" {
			h.clearSessionCookie(c)
		}
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"authenticated": true, "user": auth.User})
}

func (h *Handler) getProfile(c *gin.Context) {
	auth := mustAuth(c)
	subscription, err := h.db.GetPortalSubscription(c.Request.Context(), auth.User.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法读取用户资料")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": auth.User, "subscription": subscription})
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Phone       string `json:"phone"`
	Locale      string `json:"locale"`
	Timezone    string `json:"timezone"`
}

func (h *Handler) updateProfile(c *gin.Context) {
	auth := mustAuth(c)
	var request updateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Locale = normalizeLocale(request.Locale)
	request.Timezone = strings.TrimSpace(request.Timezone)
	if utf8.RuneCountInString(request.DisplayName) < 2 || utf8.RuneCountInString(request.DisplayName) > 64 {
		writeError(c, http.StatusBadRequest, "invalid_display_name", "昵称长度需要在 2 到 64 个字符之间")
		return
	}
	if len(request.Phone) > 32 {
		writeError(c, http.StatusBadRequest, "invalid_phone", "手机号格式不正确")
		return
	}
	if request.Timezone == "" || len(request.Timezone) > 64 {
		writeError(c, http.StatusBadRequest, "invalid_timezone", "请选择有效时区")
		return
	}
	if _, err := time.LoadLocation(request.Timezone); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_timezone", "请选择有效时区")
		return
	}
	if err := h.db.UpdatePortalProfile(c.Request.Context(), auth.User.ID, request.DisplayName, request.Phone, request.Locale, request.Timezone); err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "资料保存失败")
		return
	}
	user := h.db.GetPortalUserByID(c.Request.Context(), auth.User.ID)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) changePassword(c *gin.Context) {
	auth := mustAuth(c)
	var request changePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	if message := validatePassword(request.NewPassword); message != "" {
		writeError(c, http.StatusBadRequest, "weak_password", message)
		return
	}
	currentHash, err := h.db.GetPortalPasswordHash(c.Request.Context(), auth.User.ID)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(request.CurrentPassword)) != nil {
		writeError(c, http.StatusUnauthorized, "invalid_password", "当前密码不正确")
		return
	}
	if request.CurrentPassword == request.NewPassword {
		writeError(c, http.StatusBadRequest, "password_unchanged", "新密码不能与当前密码相同")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), 12)
	if err != nil || h.db.UpdatePortalPassword(c.Request.Context(), auth.User.ID, string(hash), auth.SessionID) != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "密码更新失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码已更新，其他设备已退出登录"})
}

func (h *Handler) listSessions(c *gin.Context) {
	auth := mustAuth(c)
	sessions, err := h.db.ListPortalSessions(c.Request.Context(), auth.User.ID, auth.SessionID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法读取登录设备")
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *Handler) revokeSession(c *gin.Context) {
	auth := mustAuth(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, http.StatusBadRequest, "invalid_session", "会话编号无效")
		return
	}
	if err := h.db.RevokePortalSession(c.Request.Context(), auth.User.ID, id); errors.Is(err, sql.ErrNoRows) {
		writeError(c, http.StatusNotFound, "session_not_found", "登录设备不存在")
		return
	} else if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法退出该设备")
		return
	}
	if id == auth.SessionID {
		h.clearSessionCookie(c)
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) revokeOtherSessions(c *gin.Context) {
	auth := mustAuth(c)
	if err := h.db.RevokeOtherPortalSessions(c.Request.Context(), auth.User.ID, auth.SessionID); err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法退出其他设备")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "其他设备已退出登录"})
}

func (h *Handler) listAPIKeys(c *gin.Context) {
	auth := mustAuth(c)
	keys, err := h.db.ListPortalAPIKeys(c.Request.Context(), auth.User.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法读取 API Key")
		return
	}
	c.JSON(http.StatusOK, gin.H{"api_keys": keys, "limit": maxUserAPIKeys})
}

type apiKeyRequest struct {
	Name string `json:"name"`
}

func (h *Handler) createAPIKey(c *gin.Context) {
	auth := mustAuth(c)
	var request apiKeyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	name := strings.TrimSpace(request.Name)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 64 {
		writeError(c, http.StatusBadRequest, "invalid_name", "Key 名称长度需要在 2 到 64 个字符之间")
		return
	}
	randomPart, _, err := newSecret(32)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法创建 API Key")
		return
	}
	rawKey := "sk-axis-" + randomPart
	prefix := rawKey
	if len(prefix) > 18 {
		prefix = prefix[:18]
	}
	item, err := h.db.CreatePortalAPIKey(c.Request.Context(), auth.User.ID, name, rawKey, prefix, hashSecret(rawKey), h.defaultQuota, maxUserAPIKeys)
	if errors.Is(err, database.ErrPortalAPIKeyLimit) {
		writeError(c, http.StatusConflict, "api_key_limit", "每个账号最多创建 10 个 API Key")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法创建 API Key")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"api_key": item, "secret": rawKey})
}

func (h *Handler) renameAPIKey(c *gin.Context) {
	auth := mustAuth(c)
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_api_key", "API Key 编号无效")
		return
	}
	var request apiKeyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	name := strings.TrimSpace(request.Name)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 64 {
		writeError(c, http.StatusBadRequest, "invalid_name", "Key 名称长度需要在 2 到 64 个字符之间")
		return
	}
	if err := h.db.RenamePortalAPIKey(c.Request.Context(), auth.User.ID, id, name); errors.Is(err, sql.ErrNoRows) {
		writeError(c, http.StatusNotFound, "api_key_not_found", "API Key 不存在")
		return
	} else if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "API Key 更新失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API Key 名称已更新"})
}

func (h *Handler) revokeAPIKey(c *gin.Context) {
	auth := mustAuth(c)
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_api_key", "API Key 编号无效")
		return
	}
	if err := h.db.RevokePortalAPIKey(c.Request.Context(), auth.User.ID, id); errors.Is(err, sql.ErrNoRows) {
		writeError(c, http.StatusNotFound, "api_key_not_found", "API Key 不存在")
		return
	} else if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "API Key 撤销失败")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) getAPIKeyUsage(c *gin.Context) {
	auth := mustAuth(c)
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_api_key", "API Key 编号无效")
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	switch days {
	case 1, 7, 30, 90:
	default:
		days = 30
	}
	end := time.Now().UTC()
	start := end.Add(-time.Duration(days) * 24 * time.Hour)
	report, err := h.db.GetPortalAPIKeyUsage(c.Request.Context(), auth.User.ID, id, start, end)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(c, http.StatusNotFound, "api_key_not_found", "API Key 不存在")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法读取用量")
		return
	}
	c.JSON(http.StatusOK, gin.H{"range_days": days, "usage": report})
}

func (h *Handler) getSubscription(c *gin.Context) {
	auth := mustAuth(c)
	subscription, err := h.db.GetPortalSubscription(c.Request.Context(), auth.User.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", "无法读取订阅信息")
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscription": subscription})
}

func (h *Handler) requireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, err := h.resolveSession(c)
		if err != nil || auth == nil {
			h.clearSessionCookie(c)
			writeError(c, http.StatusUnauthorized, "authentication_required", "请先登录")
			c.Abort()
			return
		}
		c.Set(authContextKey, auth)
		c.Next()
	}
}

func (h *Handler) resolveSession(c *gin.Context) (*authContext, error) {
	token, err := c.Cookie(sessionCookieName)
	if err != nil || token == "" {
		return nil, sql.ErrNoRows
	}
	tokenHash := hashSecret(token)
	user, sessionID, err := h.db.GetPortalUserBySession(c.Request.Context(), tokenHash, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return &authContext{User: user, SessionID: sessionID, TokenHash: tokenHash}, nil
}

func (h *Handler) sameOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		if strings.EqualFold(c.GetHeader("Sec-Fetch-Site"), "cross-site") {
			writeError(c, http.StatusForbidden, "cross_site_request", "已拒绝跨站请求")
			c.Abort()
			return
		}
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin != "" {
			parsed, err := url.Parse(origin)
			if err != nil || !strings.EqualFold(parsed.Host, c.Request.Host) {
				writeError(c, http.StatusForbidden, "origin_mismatch", "请求来源不受信任")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func (h *Handler) setSessionCookie(c *gin.Context, token string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, token, int(time.Until(expiresAt).Seconds()), "/", "", requestIsHTTPS(c), true)
}

func (h *Handler) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", requestIsHTTPS(c), true)
}

func requestIsHTTPS(c *gin.Context) bool {
	return c.Request.TLS != nil || strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https")
}

func mustAuth(c *gin.Context) *authContext {
	value, _ := c.Get(authContextKey)
	auth, _ := value.(*authContext)
	return auth
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func noStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.Next()
	}
}

func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) < 3 || len(value) > 320 || strings.ContainsAny(value, "\r\n") {
		return "", errors.New("invalid email")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || strings.ToLower(address.Address) != value || !strings.Contains(value, "@") {
		return "", errors.New("invalid email")
	}
	return value, nil
}

func validatePassword(value string) string {
	if len(value) < 10 || len(value) > 128 {
		return "密码长度需要在 10 到 128 个字符之间"
	}
	var lower, upper, digit bool
	for _, r := range value {
		lower = lower || unicode.IsLower(r)
		upper = upper || unicode.IsUpper(r)
		digit = digit || unicode.IsDigit(r)
	}
	if !lower || !upper || !digit {
		return "密码必须同时包含大写字母、小写字母和数字"
	}
	return ""
}

func newSecret(byteLength int) (string, string, error) {
	buffer := make([]byte, byteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(buffer)
	return raw, hashSecret(raw), nil
}

func hashSecret(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func preferredLocale(c *gin.Context) string {
	value := strings.ToLower(c.GetHeader("Accept-Language"))
	if strings.HasPrefix(value, "en") {
		return "en"
	}
	if strings.HasPrefix(value, "zh-tw") || strings.HasPrefix(value, "zh-hk") {
		return "zh-TW"
	}
	return "zh-CN"
}

func normalizeLocale(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "en", "en-us":
		return "en"
	case "zh-tw", "zh-hk":
		return "zh-TW"
	default:
		return "zh-CN"
	}
}

func trimString(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func parsePositiveID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func envBoolDefault(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloatDefault(name string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 || parsed > 1000000 {
		return fallback
	}
	return parsed
}

type attemptLimiter struct {
	mu        sync.Mutex
	entries   map[string][]time.Time
	lastSweep time.Time
}

func newAttemptLimiter() *attemptLimiter {
	return &attemptLimiter{entries: make(map[string][]time.Time)}
}

func (l *attemptLimiter) Allow(key string, limit int, window time.Duration) bool {
	now := time.Now()
	cutoff := now.Add(-window)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lastSweep.IsZero() || now.Sub(l.lastSweep) >= time.Minute {
		for entryKey, attempts := range l.entries {
			latest := time.Time{}
			for _, attempt := range attempts {
				if attempt.After(latest) {
					latest = attempt
				}
			}
			if latest.IsZero() || !latest.After(cutoff) {
				delete(l.entries, entryKey)
			}
		}
		l.lastSweep = now
	}
	entries := l.entries[key]
	kept := entries[:0]
	for _, item := range entries {
		if item.After(cutoff) {
			kept = append(kept, item)
		}
	}
	if len(kept) >= limit {
		l.entries[key] = kept
		return false
	}
	l.entries[key] = append(kept, now)
	return true
}

func (l *attemptLimiter) Reset(key string) {
	l.mu.Lock()
	delete(l.entries, key)
	l.mu.Unlock()
}

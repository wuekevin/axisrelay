package admin

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/wuekevin/axisrelay/database"
	"golang.org/x/crypto/bcrypt"
)

// MigrateLegacyAdminSecret converts a pre-account deployment once. The old
// secret becomes the initial password of username "admin" and is then removed
// from system settings. Existing account-based installations are untouched.
func MigrateLegacyAdminSecret(db *database.DB, envSecret string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	count, err := db.AdminUserCount(ctx)
	if err != nil || count > 0 {
		return err
	}
	settings, err := db.GetSystemSettings(ctx)
	if err != nil {
		return err
	}
	secret := strings.TrimSpace(envSecret)
	if secret == "" && settings != nil {
		secret = strings.TrimSpace(settings.AdminSecret)
	}
	if secret == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), 12)
	if err != nil {
		return err
	}
	if _, err := db.CreateAdminUser(ctx, "admin", string(hash), true); err != nil {
		return err
	}
	if settings != nil && settings.AdminSecret != "" {
		settings.AdminSecret = ""
		if err := db.UpdateSystemSettings(ctx, settings); err != nil {
			return err
		}
	}
	log.Println("旧管理密钥已安全迁移为 admin 账号初始密码；请登录后尽快修改密码")
	return nil
}

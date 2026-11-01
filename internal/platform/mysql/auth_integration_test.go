package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
	platformmysql "github.com/wuekevin/axisrelay/internal/platform/mysql"
)

func TestConfigConnectsWithCachingSHA2Password(t *testing.T) {
	rawRootDSN := strings.TrimSpace(os.Getenv("AXISRELAY_TEST_MYSQL_ROOT_DSN"))
	if rawRootDSN == "" {
		t.Skip("AXISRELAY_TEST_MYSQL_ROOT_DSN is not set")
	}

	rootCfg, err := drivermysql.ParseDSN(rawRootDSN)
	if err != nil {
		t.Fatalf("parse MySQL test configuration: %v", err)
	}
	rootCfg.DBName = ""
	rootDB, err := sql.Open("mysql", rootCfg.FormatDSN())
	if err != nil {
		t.Fatalf("open MySQL test administrator connection: %v", err)
	}
	t.Cleanup(func() { _ = rootDB.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := rootDB.PingContext(ctx); err != nil {
		t.Fatalf("ping MySQL test administrator connection: %v", err)
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	schema := "axisrelay_auth_" + suffix
	user := "axis_auth_" + suffix
	password := "AxisRelayCachingSHA2Test_42!"
	account := fmt.Sprintf("'%s'@'%%'", user)

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_, _ = rootDB.ExecContext(cleanupCtx, "DROP USER IF EXISTS "+account)
		_, _ = rootDB.ExecContext(cleanupCtx, "DROP DATABASE IF EXISTS `"+schema+"`")
	})

	if _, err := rootDB.ExecContext(ctx, "CREATE DATABASE `"+schema+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci"); err != nil {
		t.Fatalf("create isolated authentication test database: %v", err)
	}
	if _, err := rootDB.ExecContext(ctx, "CREATE USER "+account+" IDENTIFIED WITH caching_sha2_password BY '"+password+"'"); err != nil {
		t.Fatalf("create isolated caching_sha2_password test user: %v", err)
	}
	if _, err := rootDB.ExecContext(ctx, "GRANT ALL PRIVILEGES ON `"+schema+"`.* TO "+account); err != nil {
		t.Fatalf("grant isolated authentication test privileges: %v", err)
	}

	host, portText, err := net.SplitHostPort(rootCfg.Addr)
	if err != nil {
		t.Fatalf("parse MySQL test address: %v", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("parse MySQL test port: %v", err)
	}

	db, err := platformmysql.Open(ctx, platformmysql.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: schema,
	})
	if err != nil {
		t.Fatalf("connect with production MySQL configuration and caching_sha2_password: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var currentUser string
	if err := db.SQL().QueryRowContext(ctx, "SELECT CURRENT_USER()").Scan(&currentUser); err != nil {
		t.Fatalf("query through caching_sha2_password connection: %v", err)
	}
	if currentUser != user+"@%" {
		t.Fatalf("CURRENT_USER() = %q, want %q", currentUser, user+"@%")
	}
}

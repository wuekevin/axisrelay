package mysql

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
)

const (
	DefaultPort      = 3306
	DefaultCharset   = "utf8mb4"
	DefaultCollation = "utf8mb4_0900_ai_ci"
)

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectTimeout  time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	HealthTimeout   time.Duration
}

func (c Config) Normalize() Config {
	c.Host = strings.TrimSpace(c.Host)
	c.User = strings.TrimSpace(c.User)
	c.Database = strings.TrimSpace(c.Database)
	c.Charset = strings.ToLower(strings.TrimSpace(c.Charset))

	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.Charset == "" {
		c.Charset = DefaultCharset
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 50
	}
	if c.MaxIdleConns < 0 {
		c.MaxIdleConns = 0
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = min(25, c.MaxOpenConns)
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		c.MaxIdleConns = c.MaxOpenConns
	}
	if c.ConnMaxLifetime <= 0 {
		c.ConnMaxLifetime = 30 * time.Minute
	}
	if c.ConnMaxIdleTime <= 0 {
		c.ConnMaxIdleTime = 5 * time.Minute
	}
	if c.ConnectTimeout <= 0 {
		c.ConnectTimeout = 5 * time.Second
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 30 * time.Second
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 30 * time.Second
	}
	if c.HealthTimeout <= 0 {
		c.HealthTimeout = 5 * time.Second
	}
	return c
}

func (c Config) Validate() error {
	c = c.Normalize()
	switch {
	case c.Host == "":
		return fmt.Errorf("mysql host is required")
	case c.Port < 1 || c.Port > 65535:
		return fmt.Errorf("mysql port must be between 1 and 65535")
	case c.User == "":
		return fmt.Errorf("mysql user is required")
	case c.Database == "":
		return fmt.Errorf("mysql database is required")
	case c.Charset != DefaultCharset:
		return fmt.Errorf("mysql charset must be %s", DefaultCharset)
	case c.MaxOpenConns < 1:
		return fmt.Errorf("mysql max open connections must be positive")
	case c.MaxIdleConns < 0 || c.MaxIdleConns > c.MaxOpenConns:
		return fmt.Errorf("mysql max idle connections must be between 0 and max open connections")
	}
	return nil
}

func (c Config) Address() string {
	c = c.Normalize()
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c Config) DSN() (string, error) {
	c = c.Normalize()
	if err := c.Validate(); err != nil {
		return "", err
	}

	driverConfig := drivermysql.Config{
		User:         c.User,
		Passwd:       c.Password,
		Net:          "tcp",
		Addr:         c.Address(),
		DBName:       c.Database,
		ParseTime:    true,
		Loc:          time.UTC,
		Collation:    DefaultCollation,
		Timeout:      c.ConnectTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		Params: map[string]string{
			"charset":   DefaultCharset,
			"time_zone": "'+00:00'",
		},
	}
	return driverConfig.FormatDSN(), nil
}

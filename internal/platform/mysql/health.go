package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type Health struct {
	Version      string
	Major        int
	Minor        int
	Patch        int
	CharacterSet string
	Collation    string
}

func CheckHealth(ctx context.Context, db *sql.DB) (Health, error) {
	if db == nil {
		return Health{}, fmt.Errorf("mysql database is nil")
	}

	var health Health
	if err := db.QueryRowContext(
		ctx,
		"SELECT VERSION(), @@character_set_connection, @@collation_connection",
	).Scan(&health.Version, &health.CharacterSet, &health.Collation); err != nil {
		return Health{}, fmt.Errorf("query mysql health: %w", err)
	}
	return validateHealth(health)
}

func validateHealth(health Health) (Health, error) {
	major, minor, patch, err := parseMySQLVersion(health.Version)
	if err != nil {
		return Health{}, err
	}
	health.Major = major
	health.Minor = minor
	health.Patch = patch

	if major < 8 {
		return Health{}, fmt.Errorf("mysql 8.0 or newer is required, got %q", health.Version)
	}
	if !strings.EqualFold(strings.TrimSpace(health.CharacterSet), DefaultCharset) {
		return Health{}, fmt.Errorf(
			"mysql connection charset must be %s, got %q",
			DefaultCharset,
			health.CharacterSet,
		)
	}
	if collation := strings.ToLower(strings.TrimSpace(health.Collation)); collation != "" &&
		!strings.HasPrefix(collation, DefaultCharset+"_") {
		return Health{}, fmt.Errorf(
			"mysql connection collation must use %s, got %q",
			DefaultCharset,
			health.Collation,
		)
	}

	return health, nil
}

func parseMySQLVersion(version string) (major, minor, patch int, err error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return 0, 0, 0, fmt.Errorf("mysql version is empty")
	}
	if strings.Contains(strings.ToLower(version), "mariadb") {
		return 0, 0, 0, fmt.Errorf("MariaDB is not supported: %q", version)
	}

	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, 0, fmt.Errorf("invalid mysql version %q", version)
	}

	parsePart := func(value string) (int, error) {
		var digits strings.Builder
		for _, r := range value {
			if r < '0' || r > '9' {
				break
			}
			digits.WriteRune(r)
		}
		if digits.Len() == 0 {
			return 0, fmt.Errorf("invalid mysql version %q", version)
		}
		n, parseErr := strconv.Atoi(digits.String())
		if parseErr != nil {
			return 0, fmt.Errorf("invalid mysql version %q: %w", version, parseErr)
		}
		return n, nil
	}

	if major, err = parsePart(parts[0]); err != nil {
		return 0, 0, 0, err
	}
	if minor, err = parsePart(parts[1]); err != nil {
		return 0, 0, 0, err
	}
	if len(parts) >= 3 {
		if patch, err = parsePart(parts[2]); err != nil {
			return 0, 0, 0, err
		}
	}
	return major, minor, patch, nil
}

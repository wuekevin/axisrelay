package mysql

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var migrationFilenamePattern = regexp.MustCompile(`^(\d{6})_([a-z0-9_]+)\.sql$`)

type SQLMigrationFile struct {
	Version  uint64
	Name     string
	Path     string
	Checksum string
	SQL      string
}

func LoadSQLMigrationFiles(dir string) ([]SQLMigrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read mysql migration directory: %w", err)
	}

	files := make([]SQLMigrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".sql") {
			continue
		}
		match := migrationFilenamePattern.FindStringSubmatch(name)
		if match == nil {
			return nil, fmt.Errorf("invalid mysql migration filename %q", name)
		}
		version, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil || version == 0 {
			return nil, fmt.Errorf("invalid mysql migration version in %q", name)
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read mysql migration %q: %w", name, err)
		}
		sqlText := strings.TrimSpace(string(data))
		if sqlText == "" {
			return nil, fmt.Errorf("mysql migration %q is empty", name)
		}
		digest := sha256.Sum256(data)
		files = append(files, SQLMigrationFile{
			Version:  version,
			Name:     match[2],
			Path:     path,
			Checksum: hex.EncodeToString(digest[:]),
			SQL:      sqlText,
		})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Version < files[j].Version })
	seenVersions := make(map[uint64]string, len(files))
	seenNames := make(map[string]uint64, len(files))
	for _, file := range files {
		if existing, ok := seenVersions[file.Version]; ok {
			return nil, fmt.Errorf("duplicate mysql migration version %d: %q and %q", file.Version, existing, file.Name)
		}
		if existing, ok := seenNames[file.Name]; ok {
			return nil, fmt.Errorf("duplicate mysql migration name %q: versions %d and %d", file.Name, existing, file.Version)
		}
		seenVersions[file.Version] = file.Name
		seenNames[file.Name] = file.Version
	}
	return files, nil
}

func BuildSQLMigrations(dir string) ([]Migration, error) {
	files, err := LoadSQLMigrationFiles(dir)
	if err != nil {
		return nil, err
	}
	migrations := make([]Migration, 0, len(files))
	for _, file := range files {
		file := file
		statements := splitSQLStatements(file.SQL)
		migrations = append(migrations, Migration{
			Version:  file.Version,
			Name:     file.Name,
			Checksum: file.Checksum,
			Up: func(ctx context.Context, db MigrationDB) error {
				for _, statement := range statements {
					if _, err := db.ExecContext(ctx, statement); err != nil {
						return fmt.Errorf("execute %s statement: %w", filepath.Base(file.Path), err)
					}
				}
				return nil
			},
		})
	}
	if _, err := prepareMigrations(migrations); err != nil {
		return nil, err
	}
	return migrations, nil
}

func splitSQLStatements(sqlText string) []string {
	statements := make([]string, 0, 16)
	var builder strings.Builder
	var quote byte
	lineComment := false
	blockComment := false

	flush := func() {
		statement := strings.TrimSpace(builder.String())
		builder.Reset()
		if statement != "" {
			statements = append(statements, statement)
		}
	}

	for i := 0; i < len(sqlText); i++ {
		ch := sqlText[i]
		if lineComment {
			builder.WriteByte(ch)
			if ch == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			builder.WriteByte(ch)
			if ch == '*' && i+1 < len(sqlText) && sqlText[i+1] == '/' {
				builder.WriteByte('/')
				i++
				blockComment = false
			}
			continue
		}
		if quote != 0 {
			builder.WriteByte(ch)
			if ch == '\\' && i+1 < len(sqlText) {
				builder.WriteByte(sqlText[i+1])
				i++
				continue
			}
			if ch == quote {
				if i+1 < len(sqlText) && sqlText[i+1] == quote {
					builder.WriteByte(sqlText[i+1])
					i++
					continue
				}
				quote = 0
			}
			continue
		}

		if ch == '-' && i+1 < len(sqlText) && sqlText[i+1] == '-' {
			builder.WriteString("--")
			i++
			lineComment = true
			continue
		}
		if ch == '#' {
			builder.WriteByte(ch)
			lineComment = true
			continue
		}
		if ch == '/' && i+1 < len(sqlText) && sqlText[i+1] == '*' {
			builder.WriteString("/*")
			i++
			blockComment = true
			continue
		}

		switch ch {
		case '\'', '"', 0x60:
			quote = ch
			builder.WriteByte(ch)
		case ';':
			flush()
		default:
			builder.WriteByte(ch)
		}
	}
	flush()
	return statements
}

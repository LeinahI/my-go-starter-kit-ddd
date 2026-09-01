package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const migrationTemplate = `-- +goose Up
SELECT 'up SQL query';

-- +goose Down
SELECT 'down SQL query';
`

func createMigration(dir, name string) error {
	now := time.Now().UTC()
	date := now.Format("20060102")
	clock := now.Format("150405")
	filename := fmt.Sprintf("%s_%s_%s.sql", date, clock, snakeCase(name))
	path := filepath.Join(dir, filename)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("migration already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.WriteFile(path, []byte(migrationTemplate), 0o644); err != nil {
		return fmt.Errorf("write migration: %w", err)
	}
	fmt.Printf("Created new file: %s\n", path)
	return nil
}

func snakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "migration"
	}
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		if r == '-' || r == ' ' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

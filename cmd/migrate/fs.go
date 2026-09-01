package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type migrationDirFS struct {
	dir string
}

func (f migrationDirFS) Open(name string) (fs.File, error) {
	if strings.Contains(name, "..") {
		return nil, fmt.Errorf("invalid migration path: %s", name)
	}
	actual, err := f.resolveActual(name)
	if err != nil {
		return nil, err
	}
	return os.Open(filepath.Join(f.dir, actual))
}

func (f migrationDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name != "." {
		return nil, fs.ErrNotExist
	}
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return nil, err
	}
	out := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !isMigrationFile(entry.Name()) {
			continue
		}
		out = append(out, gooseDirEntry{DirEntry: entry, name: gooseFilename(entry.Name())})
	}
	return out, nil
}

func (f migrationDirFS) Glob(pattern string) ([]string, error) {
	entries, err := f.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, entry := range entries {
		ok, err := filepath.Match(pattern, entry.Name())
		if err != nil {
			return nil, err
		}
		if ok {
			matches = append(matches, entry.Name())
		}
	}
	return matches, nil
}

func (f migrationDirFS) resolveActual(gooseName string) (string, error) {
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !isMigrationFile(entry.Name()) {
			continue
		}
		if gooseFilename(entry.Name()) == gooseName || entry.Name() == gooseName {
			return entry.Name(), nil
		}
	}
	return "", fs.ErrNotExist
}

type gooseDirEntry struct {
	fs.DirEntry
	name string
}

func (e gooseDirEntry) Name() string { return e.name }

func isMigrationFile(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".sql" || (ext == ".go" && !strings.HasSuffix(name, "_test.go"))
}

func gooseFilename(actual string) string {
	ext := filepath.Ext(actual)
	base := strings.TrimSuffix(actual, ext)
	parts := strings.Split(base, "_")
	if len(parts) >= 3 && len(parts[0]) == 8 && len(parts[1]) == 6 && isDigits(parts[0]) && isDigits(parts[1]) {
		return parts[0] + parts[1] + "_" + strings.Join(parts[2:], "_") + ext
	}
	return actual
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type migrationFile struct {
	ULID string
	Name string
	Up   string // path to .up.sql
	Down string // path to .down.sql
}

var migrationFilePattern = regexp.MustCompile(`^([0-9A-HJKMNP-TV-Z]{26})_(.+)\.(up|down)\.sql$`)

func DiscoverMigrations(migrationsPath string) ([]migrationFile, error) {
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationsMap := make(map[string]*migrationFile)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		ulid, name, typ := matches[1], matches[2], matches[3]
		key := ulid + "_" + name
		mf, ok := migrationsMap[key]
		if !ok {
			mf = &migrationFile{ULID: ulid, Name: name}
			migrationsMap[key] = mf
		}
		path := filepath.Join(migrationsPath, entry.Name())
		if typ == "up" {
			mf.Up = path
		} else if typ == "down" {
			mf.Down = path
		}
	}

	var migrations []migrationFile
	for _, mf := range migrationsMap {
		if mf.Up == "" || mf.Down == "" {
			return nil, fmt.Errorf("migration pair missing for: %s_%s", mf.ULID, mf.Name)
		}
		migrations = append(migrations, *mf)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ULID < migrations[j].ULID
	})

	return migrations, nil
}

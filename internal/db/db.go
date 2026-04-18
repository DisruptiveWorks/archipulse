// Package db handles PostgreSQL connection and migration runner.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

// Connect opens and verifies a PostgreSQL connection using DATABASE_URL from the environment.
func Connect() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

// Migrate runs all pending SQL migrations from the given directory.
// Migration files must be named NNN_*.sql where NNN is a zero-padded integer version.
// A PostgreSQL session-level advisory lock (key 0xArchipulse) serialises concurrent
// callers (e.g. parallel test packages sharing the same database).
func Migrate(db *sql.DB, migrationsDir string) error {
	// Acquire an exclusive advisory lock for the duration of migration.
	// pg_advisory_lock blocks until acquired; pg_advisory_unlock releases it.
	const lockKey = 0x417263 // "Arc" in hex — arbitrary but stable
	if _, err := db.Exec(`SELECT pg_advisory_lock($1)`, lockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() { _, _ = db.Exec(`SELECT pg_advisory_unlock($1)`, lockKey) }()

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, f := range files {
		ver, err := versionFromFilename(filepath.Base(f))
		if err != nil {
			return err
		}
		if applied[ver] {
			continue
		}
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, ver); err != nil {
			return fmt.Errorf("record migration %d: %w", ver, err)
		}
		fmt.Printf("applied migration %03d: %s\n", ver, filepath.Base(f))
	}
	return nil
}

func appliedVersions(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func versionFromFilename(name string) (int, error) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("migration filename must be NNN_description.sql, got: %s", name)
	}
	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("migration filename must start with integer version, got: %s", name)
	}
	return v, nil
}

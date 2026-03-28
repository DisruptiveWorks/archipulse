package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"

	"github.com/DisruptiveWorks/archipulse/internal/api"
	"github.com/DisruptiveWorks/archipulse/internal/db"
)

func main() {
	// Load .env if present — errors are intentionally ignored (file is optional).
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: archipulse <command>")
		fmt.Fprintln(os.Stderr, "commands: serve, migrate")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "migrate":
		if err := runMigrate(); err != nil {
			fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
			os.Exit(1)
		}
	case "serve":
		if err := runServe(); err != nil {
			fmt.Fprintf(os.Stderr, "serve: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runServe() error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	fmt.Printf("listening on %s\n", addr)
	return http.ListenAndServe(addr, api.NewRouter(conn))
}

func runMigrate() error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	migrationsDir := migrationsPath()
	fmt.Printf("running migrations from %s\n", migrationsDir)
	return db.Migrate(conn, migrationsDir)
}

// migrationsPath returns the path to the migrations directory relative to the binary location.
// Falls back to a path relative to the source file for `go run` usage.
func migrationsPath() string {
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "migrations")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// go run: derive from source location
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
}

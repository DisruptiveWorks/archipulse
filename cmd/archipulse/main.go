package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"

	"github.com/DisruptiveWorks/archipulse/internal/api"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/cli"
	"github.com/DisruptiveWorks/archipulse/internal/db"
	"github.com/DisruptiveWorks/archipulse/internal/mcpserver"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// version is set at build time via -ldflags="-X main.version=x.y.z".
var version = "dev"

func main() {
	// Load .env if present — errors are intentionally ignored (file is optional).
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "migrate":
		if err := runMigrate(); err != nil {
			fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
			os.Exit(1)
		}
	case "seed":
		if err := runSeed(); err != nil {
			fmt.Fprintf(os.Stderr, "seed: %v\n", err)
			os.Exit(1)
		}
	case "serve":
		if err := runServe(); err != nil {
			fmt.Fprintf(os.Stderr, "serve: %v\n", err)
			os.Exit(1)
		}
	case "login":
		if err := cli.RunLogin(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "login: %v\n", err)
			os.Exit(1)
		}
	case "context":
		if err := cli.RunContext(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "context: %v\n", err)
			os.Exit(1)
		}
	case "workspace":
		if err := cli.RunWorkspace(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "workspace: %v\n", err)
			os.Exit(1)
		}
	case "element":
		if err := cli.RunElement(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "element: %v\n", err)
			os.Exit(1)
		}
	case "relationship":
		if err := cli.RunRelationship(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "relationship: %v\n", err)
			os.Exit(1)
		}
	case "diagram":
		if err := cli.RunDiagram(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "diagram: %v\n", err)
			os.Exit(1)
		}
	case "import":
		if err := cli.RunImport(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "import: %v\n", err)
			os.Exit(1)
		}
	case "export":
		if err := cli.RunExport(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "export: %v\n", err)
			os.Exit(1)
		}
	case "mcp":
		var contextOverride string
		if len(os.Args) >= 4 && os.Args[2] == "--context" {
			contextOverride = os.Args[3]
		}
		if err := mcpserver.Serve(contextOverride); err != nil {
			fmt.Fprintf(os.Stderr, "mcp: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: archipulse <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "server commands:")
	fmt.Fprintln(os.Stderr, "  serve      start the API server")
	fmt.Fprintln(os.Stderr, "  migrate    run database migrations")
	fmt.Fprintln(os.Stderr, "  seed       load demo data")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "client commands:")
	fmt.Fprintln(os.Stderr, "  login      authenticate and save credentials")
	fmt.Fprintln(os.Stderr, "  context    manage named server contexts")
	fmt.Fprintln(os.Stderr, "  workspace  list and inspect workspaces")
	fmt.Fprintln(os.Stderr, "  element      list and inspect elements")
	fmt.Fprintln(os.Stderr, "  relationship list and inspect relationships")
	fmt.Fprintln(os.Stderr, "  diagram      list and inspect diagrams")
	fmt.Fprintln(os.Stderr, "  import       import an AOEF (.xml) file into a workspace")
	fmt.Fprintln(os.Stderr, "  export       export a workspace as AOEF (.xml)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "agent commands:")
	fmt.Fprintln(os.Stderr, "  mcp          start the MCP server (stdio transport)")
}

func runServe() error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}

	// Initialise auth.
	cfg, err := auth.ConfigFromEnv()
	if err != nil {
		return fmt.Errorf("auth config: %w", err)
	}
	svc, err := auth.NewService(conn, cfg)
	if err != nil {
		return fmt.Errorf("auth service: %w", err)
	}
	if err := auth.Bootstrap(svc); err != nil {
		return fmt.Errorf("auth bootstrap: %w", err)
	}

	// Optional OIDC provider.
	oidcProvider, err := auth.NewOIDCProvider(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("oidc provider: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	fmt.Printf("listening on %s\n", addr)
	return http.ListenAndServe(addr, api.NewRouter(conn, svc, oidcProvider, version, staticFiles))
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

func runSeed() error {
	conn, err := db.Connect()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	store := workspace.NewStore(conn)
	wss, err := store.List()
	if err != nil {
		return fmt.Errorf("list workspaces: %w", err)
	}
	if len(wss) > 0 {
		fmt.Println("seed: workspaces already exist, skipping")
		return nil
	}

	modelPath := filepath.Join(examplesPath(), "archisurance-extended.xml")
	f, err := os.Open(modelPath)
	if err != nil {
		return fmt.Errorf("open seed model %s: %w", modelPath, err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAOEF(f)
	if err != nil {
		return fmt.Errorf("parse seed model: %w", err)
	}
	if err := parser.Validate(m); err != nil {
		return fmt.Errorf("validate seed model: %w", err)
	}

	ws, err := store.Create("ArchiSurance Demo", "as-is", "Pre-loaded demonstration model from The Open Group ArchiSurance case study.")
	if err != nil {
		return fmt.Errorf("create seed workspace: %w", err)
	}

	result, err := api.ImportModel(conn, ws.ID, m)
	if err != nil {
		return fmt.Errorf("import seed model: %w", err)
	}

	fmt.Printf("seed: workspace %q created — %d elements, %d relationships, %d diagrams\n",
		ws.Name, result.Elements, result.Relationships, result.Diagrams)
	return nil
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

// examplesPath returns the path to the examples directory relative to the binary location.
func examplesPath() string {
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "examples")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "examples")
}

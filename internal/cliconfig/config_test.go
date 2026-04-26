package cliconfig_test

import (
	"path/filepath"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/cliconfig"
)

func TestUpsertAndSaveLoad(t *testing.T) {
	dir := t.TempDir()
	overrideConfigPath(t, filepath.Join(dir, ".archipulse", "config.yaml"))

	f := &cliconfig.File{}
	f.UpsertContext("local", "http://localhost:8080", "tok-local", "", true)
	f.UpsertContext("prod", "https://api.example.com", "tok-prod", "ws-uuid", false)

	if err := f.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := cliconfig.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.CurrentContext != "local" {
		t.Errorf("CurrentContext = %q, want %q", loaded.CurrentContext, "local")
	}
	if len(loaded.Contexts) != 2 {
		t.Errorf("len(Contexts) = %d, want 2", len(loaded.Contexts))
	}
}

func TestUpsertUpdatesExisting(t *testing.T) {
	f := &cliconfig.File{}
	f.UpsertContext("local", "http://localhost:8080", "old-token", "", true)
	f.UpsertContext("local", "", "new-token", "", false)

	if f.Contexts[0].Token != "new-token" {
		t.Errorf("token not updated: got %q", f.Contexts[0].Token)
	}
	if f.Contexts[0].Server != "http://localhost:8080" {
		t.Errorf("server unexpectedly cleared: got %q", f.Contexts[0].Server)
	}
}

func TestRemoveContext(t *testing.T) {
	f := &cliconfig.File{}
	f.UpsertContext("a", "http://a", "tok-a", "", true)
	f.UpsertContext("b", "http://b", "tok-b", "", false)

	if err := f.RemoveContext("a"); err == nil {
		t.Error("expected error removing current context")
	}
	if err := f.RemoveContext("b"); err != nil {
		t.Errorf("RemoveContext b: %v", err)
	}
	if len(f.Contexts) != 1 {
		t.Errorf("len(Contexts) = %d, want 1", len(f.Contexts))
	}
	if err := f.RemoveContext("nonexistent"); err == nil {
		t.Error("expected error for nonexistent context")
	}
}

func TestUseContext(t *testing.T) {
	f := &cliconfig.File{}
	f.UpsertContext("a", "http://a", "tok-a", "", true)
	f.UpsertContext("b", "http://b", "tok-b", "", false)

	if err := f.UseContext("b"); err != nil {
		t.Fatalf("UseContext: %v", err)
	}
	if f.CurrentContext != "b" {
		t.Errorf("CurrentContext = %q, want %q", f.CurrentContext, "b")
	}
	if err := f.UseContext("nonexistent"); err == nil {
		t.Error("expected error for nonexistent context")
	}
}

func TestResolvePrecedence(t *testing.T) {
	dir := t.TempDir()
	overrideConfigPath(t, filepath.Join(dir, ".archipulse", "config.yaml"))

	// Seed config file with a context.
	f := &cliconfig.File{}
	f.UpsertContext("default", "http://from-file", "file-token", "file-ws", true)
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}

	t.Run("file wins over hardcoded default", func(t *testing.T) {
		r := cliconfig.Resolve(cliconfig.ResolveOptions{})
		if r.Server != "http://from-file" {
			t.Errorf("Server = %q, want %q", r.Server, "http://from-file")
		}
		if r.Token != "file-token" {
			t.Errorf("Token = %q, want %q", r.Token, "file-token")
		}
	})

	t.Run("env var wins over file", func(t *testing.T) {
		t.Setenv("ARCHIPULSE_SERVER", "http://from-env")
		t.Setenv("ARCHIPULSE_TOKEN", "env-token")
		r := cliconfig.Resolve(cliconfig.ResolveOptions{})
		if r.Server != "http://from-env" {
			t.Errorf("Server = %q, want %q", r.Server, "http://from-env")
		}
		if r.Token != "env-token" {
			t.Errorf("Token = %q, want %q", r.Token, "env-token")
		}
	})

	t.Run("flag wins over env var", func(t *testing.T) {
		t.Setenv("ARCHIPULSE_SERVER", "http://from-env")
		r := cliconfig.Resolve(cliconfig.ResolveOptions{Server: "http://from-flag"})
		if r.Server != "http://from-flag" {
			t.Errorf("Server = %q, want %q", r.Server, "http://from-flag")
		}
	})
}

// overrideConfigPath redirects the config file location for testing by
// temporarily setting HOME to a temp directory.
func overrideConfigPath(t *testing.T, path string) {
	t.Helper()
	home := filepath.Dir(filepath.Dir(path))
	t.Setenv("HOME", home)
}

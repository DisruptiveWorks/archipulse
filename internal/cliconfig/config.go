// Package cliconfig manages the client-side configuration file used by the
// ArchiPulse CLI and MCP server subcommands (~/.archipulse/config.yaml).
package cliconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultServer = "http://localhost:8080"

// File is the in-memory representation of ~/.archipulse/config.yaml.
type File struct {
	CurrentContext string    `yaml:"current-context"`
	Contexts       []Context `yaml:"contexts"`
}

// Context holds connection details for one named ArchiPulse server.
type Context struct {
	Name      string `yaml:"name"`
	Server    string `yaml:"server"`
	Token     string `yaml:"token"`
	Workspace string `yaml:"workspace,omitempty"`
}

// Resolved is the final set of connection params after applying the
// precedence chain: flags > env vars > config file > defaults.
type Resolved struct {
	Server    string
	Token     string
	Workspace string
}

// ResolveOptions are overrides provided at call time (e.g. from CLI flags).
// An empty string means "not set" — fall through to the next source.
type ResolveOptions struct {
	Server    string
	Token     string
	Workspace string
	Context   string // override which context to use
}

// Resolve returns the active connection params using the precedence chain.
// It never returns an error — missing config files are silently ignored.
func Resolve(opts ResolveOptions) Resolved {
	r := Resolved{
		Server:    defaultServer,
		Token:     os.Getenv("ARCHIPULSE_TOKEN"),
		Workspace: os.Getenv("ARCHIPULSE_WORKSPACE"),
	}
	if v := os.Getenv("ARCHIPULSE_SERVER"); v != "" {
		r.Server = v
	}

	// Overlay config file values (lower priority than env vars).
	if f, err := Load(); err == nil {
		ctx := f.active(opts.Context)
		if ctx != nil {
			if ctx.Server != "" && r.Server == defaultServer && os.Getenv("ARCHIPULSE_SERVER") == "" {
				r.Server = ctx.Server
			}
			if ctx.Token != "" && r.Token == "" {
				r.Token = ctx.Token
			}
			if ctx.Workspace != "" && r.Workspace == "" {
				r.Workspace = ctx.Workspace
			}
		}
	}

	// Flags override everything.
	if opts.Server != "" {
		r.Server = opts.Server
	}
	if opts.Token != "" {
		r.Token = opts.Token
	}
	if opts.Workspace != "" {
		r.Workspace = opts.Workspace
	}

	return r
}

// Load reads the config file. Returns an empty File (no error) if the file
// does not exist yet.
func Load() (*File, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &File{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &f, nil
}

// Save writes the file back to disk, creating the directory if needed.
func (f *File) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// UpsertContext adds or updates a named context and optionally sets it as
// current. Token is the only field required after initial creation.
func (f *File) UpsertContext(name, server, token, workspace string, setCurrent bool) {
	for i, c := range f.Contexts {
		if c.Name == name {
			if server != "" {
				f.Contexts[i].Server = server
			}
			if token != "" {
				f.Contexts[i].Token = token
			}
			if workspace != "" {
				f.Contexts[i].Workspace = workspace
			}
			if setCurrent {
				f.CurrentContext = name
			}
			return
		}
	}
	f.Contexts = append(f.Contexts, Context{
		Name:      name,
		Server:    server,
		Token:     token,
		Workspace: workspace,
	})
	if setCurrent || f.CurrentContext == "" {
		f.CurrentContext = name
	}
}

// RemoveContext deletes a context by name. Returns an error if it was the
// current context, so the caller can prompt the user to switch first.
func (f *File) RemoveContext(name string) error {
	if f.CurrentContext == name {
		return fmt.Errorf("context %q is current — switch to another context first", name)
	}
	out := f.Contexts[:0]
	for _, c := range f.Contexts {
		if c.Name != name {
			out = append(out, c)
		}
	}
	if len(out) == len(f.Contexts) {
		return fmt.Errorf("context %q not found", name)
	}
	f.Contexts = out
	return nil
}

// UseContext sets the active context. Returns an error if the name is not found.
func (f *File) UseContext(name string) error {
	for _, c := range f.Contexts {
		if c.Name == name {
			f.CurrentContext = name
			return nil
		}
	}
	return fmt.Errorf("context %q not found", name)
}

// active returns the Context to use. If override is non-empty it is used
// instead of CurrentContext. Returns nil if no matching context exists.
func (f *File) active(override string) *Context {
	name := f.CurrentContext
	if override != "" {
		name = override
	}
	for i, c := range f.Contexts {
		if c.Name == name {
			return &f.Contexts[i]
		}
	}
	return nil
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".archipulse", "config.yaml"), nil
}

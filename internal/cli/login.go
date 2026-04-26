package cli

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/DisruptiveWorks/archipulse/internal/cliconfig"
)

// RunLogin handles the `archipulse login` subcommand.
// It prompts for credentials interactively when --email / --password are omitted.
func RunLogin(args []string) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	serverFlag := fs.String("server", "", "ArchiPulse server URL (default from config or http://localhost:8080)")
	contextFlag := fs.String("context", "", "Config context to save credentials under (default: current context or \"default\")")
	emailFlag := fs.String("email", "", "Email address")
	passwordFlag := fs.String("password", "", "Password (omit to be prompted securely)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Resolve server: flag > env > config > hardcoded default.
	resolved := cliconfig.Resolve(cliconfig.ResolveOptions{Server: *serverFlag})
	server := resolved.Server
	if *serverFlag != "" {
		server = *serverFlag
	}

	// Context name: flag > current context > "default".
	contextName := *contextFlag
	if contextName == "" {
		if f, err := cliconfig.Load(); err == nil && f.CurrentContext != "" {
			contextName = f.CurrentContext
		} else {
			contextName = "default"
		}
	}

	email := *emailFlag
	password := *passwordFlag

	r := bufio.NewReader(os.Stdin)

	if email == "" {
		fmt.Print("Email: ")
		line, _ := r.ReadString('\n')
		email = strings.TrimSpace(line)
	}
	if password == "" {
		fmt.Print("Password: ")
		if term.IsTerminal(int(os.Stdin.Fd())) {
			raw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			password = string(raw)
		} else {
			line, _ := r.ReadString('\n')
			password = strings.TrimSpace(line)
		}
	}

	client := NewClient(server, "")
	resp, err := client.Do(http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return fmt.Errorf("connect to %s: %w", server, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid credentials")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed (HTTP %d)", resp.StatusCode)
	}

	token := ""
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "ap_session" {
			token = cookie.Value
			break
		}
	}
	if token == "" {
		return fmt.Errorf("server did not return a session cookie")
	}

	f, err := cliconfig.Load()
	if err != nil {
		return err
	}
	f.UpsertContext(contextName, server, token, "", true)
	if err := f.Save(); err != nil {
		return err
	}

	fmt.Printf("Logged in as %s\n", email)
	fmt.Printf("Context %q saved → %s\n", contextName, server)
	return nil
}

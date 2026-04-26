package cli

import (
	"flag"
	"fmt"

	"github.com/DisruptiveWorks/archipulse/internal/cliconfig"
)

// RunContext handles the `archipulse context` subcommand.
func RunContext(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: archipulse context <list|use|add|remove>")
	}
	switch args[0] {
	case "list":
		return contextList()
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: archipulse context use <name>")
		}
		return contextUse(args[1])
	case "add":
		return contextAdd(args[1:])
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: archipulse context remove <name>")
		}
		return contextRemove(args[1])
	default:
		return fmt.Errorf("unknown subcommand %q — use list, use, add, or remove", args[0])
	}
}

func contextList() error {
	f, err := cliconfig.Load()
	if err != nil {
		return err
	}
	if len(f.Contexts) == 0 {
		fmt.Println("No contexts configured. Run 'archipulse login' to get started.")
		return nil
	}
	for _, c := range f.Contexts {
		marker := "  "
		if c.Name == f.CurrentContext {
			marker = "* "
		}
		status := "not authenticated"
		if c.Token != "" {
			status = "authenticated"
		}
		fmt.Printf("%s%-20s  %-40s  %s\n", marker, c.Name, c.Server, status)
	}
	return nil
}

func contextUse(name string) error {
	f, err := cliconfig.Load()
	if err != nil {
		return err
	}
	if err := f.UseContext(name); err != nil {
		return err
	}
	if err := f.Save(); err != nil {
		return err
	}
	fmt.Printf("Switched to context %q\n", name)
	return nil
}

func contextAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: archipulse context add <name> [--server URL]")
	}
	name := args[0]

	fs := flag.NewFlagSet("context add", flag.ContinueOnError)
	server := fs.String("server", "http://localhost:8080", "ArchiPulse server URL")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	f, err := cliconfig.Load()
	if err != nil {
		return err
	}
	f.UpsertContext(name, *server, "", "", false)
	if err := f.Save(); err != nil {
		return err
	}
	fmt.Printf("Context %q added → %s\n", name, *server)
	fmt.Printf("Run 'archipulse login --context %s' to authenticate.\n", name)
	return nil
}

func contextRemove(name string) error {
	f, err := cliconfig.Load()
	if err != nil {
		return err
	}
	if err := f.RemoveContext(name); err != nil {
		return err
	}
	if err := f.Save(); err != nil {
		return err
	}
	fmt.Printf("Context %q removed\n", name)
	return nil
}

package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Purpose     string    `json:"purpose"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RunWorkspace handles `archipulse workspace <subcommand>`.
func RunWorkspace(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: archipulse workspace <list|get>")
	}
	switch args[0] {
	case "list":
		return workspaceList(args[1:])
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: archipulse workspace get <id>")
		}
		return workspaceGet(args[1], args[2:])
	default:
		return fmt.Errorf("unknown subcommand %q — use list or get", args[0])
	}
}

func workspaceList(args []string) error {
	fs := flag.NewFlagSet("workspace list", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	if err := fs.Parse(args); err != nil {
		return err
	}

	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	resp, err := client.do(http.MethodGet, "/workspaces", nil)
	if err != nil {
		return err
	}

	var workspaces []workspace
	if err := client.decode(resp, &workspaces); err != nil {
		return err
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(workspaces)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	rows := make([][]string, len(workspaces))
	for i, w := range workspaces {
		desc := w.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		rows[i] = []string{w.ID, w.Name, w.Purpose, desc}
	}
	table([]string{"ID", "NAME", "PURPOSE", "DESCRIPTION"}, rows)
	return nil
}

func workspaceGet(id string, args []string) error {
	fs := flag.NewFlagSet("workspace get", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	if err := fs.Parse(args); err != nil {
		return err
	}

	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	resp, err := client.do(http.MethodGet, "/workspaces/"+id, nil)
	if err != nil {
		return err
	}

	var w workspace
	if err := client.decode(resp, &w); err != nil {
		return err
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(w)
	}

	fmt.Printf("ID:          %s\n", w.ID)
	fmt.Printf("Name:        %s\n", w.Name)
	fmt.Printf("Purpose:     %s\n", w.Purpose)
	if w.Description != "" {
		fmt.Printf("Description: %s\n", w.Description)
	}
	fmt.Printf("Version:     %d\n", w.Version)
	fmt.Printf("Updated:     %s\n", w.UpdatedAt.Local().Format("2006-01-02 15:04"))
	return nil
}

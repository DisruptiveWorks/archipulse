package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type diagram struct {
	ID            string          `json:"id"`
	WorkspaceID   string          `json:"workspace_id"`
	SourceID      string          `json:"source_id"`
	Name          string          `json:"name"`
	Documentation string          `json:"documentation"`
	Layout        json.RawMessage `json:"layout"`
	Version       int             `json:"version"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type diagramPage struct {
	Items []diagram `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

// RunDiagram handles `archipulse diagram <subcommand>`.
func RunDiagram(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: archipulse diagram <list|get> --workspace <id>")
	}
	switch args[0] {
	case "list":
		return diagramList(args[1:])
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: archipulse diagram get <id> --workspace <wsID>")
		}
		return diagramGet(args[1], args[2:])
	default:
		return fmt.Errorf("unknown subcommand %q — use list or get", args[0])
	}
}

func diagramList(args []string) error {
	fs := flag.NewFlagSet("diagram list", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	wsID := fs.String("workspace", "", "Workspace ID (required)")
	page := fs.Int("page", 1, "Page number")
	limit := fs.Int("limit", 50, "Results per page")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *wsID == "" {
		return fmt.Errorf("--workspace is required")
	}

	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/workspaces/%s/diagrams?page=%d&limit=%d", *wsID, *page, *limit)
	resp, err := client.do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	var result diagramPage
	if err := client.decode(resp, &result); err != nil {
		return err
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(result.Items)
	}

	if len(result.Items) == 0 {
		fmt.Println("No diagrams found.")
		return nil
	}

	rows := make([][]string, len(result.Items))
	for i, d := range result.Items {
		doc := d.Documentation
		if len(doc) > 60 {
			doc = doc[:57] + "..."
		}
		rows[i] = []string{d.ID, d.Name, doc}
	}
	table([]string{"ID", "NAME", "DOCUMENTATION"}, rows)
	fmt.Printf("\n%d–%d of %d\n", (result.Page-1)*result.Limit+1, (result.Page-1)*result.Limit+len(result.Items), result.Total)
	return nil
}

func diagramGet(id string, args []string) error {
	fs := flag.NewFlagSet("diagram get", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	wsID := fs.String("workspace", "", "Workspace ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *wsID == "" {
		return fmt.Errorf("--workspace is required")
	}

	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	resp, err := client.do(http.MethodGet, fmt.Sprintf("/workspaces/%s/diagrams/%s", *wsID, id), nil)
	if err != nil {
		return err
	}

	var d diagram
	if err := client.decode(resp, &d); err != nil {
		return err
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	}

	fmt.Printf("ID:            %s\n", d.ID)
	fmt.Printf("Name:          %s\n", d.Name)
	fmt.Printf("Source ID:     %s\n", d.SourceID)
	if d.Documentation != "" {
		fmt.Printf("Documentation: %s\n", d.Documentation)
	}
	fmt.Printf("Version:       %d\n", d.Version)
	fmt.Printf("Updated:       %s\n", d.UpdatedAt.Local().Format("2006-01-02 15:04"))
	return nil
}

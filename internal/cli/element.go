package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type element struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	SourceID      string    `json:"source_id"`
	Type          string    `json:"type"`
	Layer         string    `json:"layer"`
	Name          string    `json:"name"`
	Documentation string    `json:"documentation"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type elementPage struct {
	Items []element `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

type property struct {
	ID          string     `json:"id"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Source      string     `json:"source"`
	CollectedAt *time.Time `json:"collected_at"`
}

// RunElement handles `archipulse element <subcommand>`.
func RunElement(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: archipulse element <list|get> --workspace <id>")
	}
	switch args[0] {
	case "list":
		return elementList(args[1:])
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: archipulse element get <id> --workspace <wsID>")
		}
		return elementGet(args[1], args[2:])
	default:
		return fmt.Errorf("unknown subcommand %q — use list or get", args[0])
	}
}

func elementList(args []string) error {
	fs := flag.NewFlagSet("element list", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	wsID := fs.String("workspace", "", "Workspace ID (required)")
	page := fs.Int("page", 1, "Page number")
	limit := fs.Int("limit", 50, "Results per page")
	typeFilter := fs.String("type", "", "Filter by ArchiMate type")
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

	path := fmt.Sprintf("/workspaces/%s/elements?page=%d&limit=%d", *wsID, *page, *limit)
	resp, err := client.Do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	var result elementPage
	if err := client.decode(resp, &result); err != nil {
		return err
	}

	items := result.Items
	if *typeFilter != "" {
		filtered := items[:0]
		for _, e := range items {
			if e.Type == *typeFilter {
				filtered = append(filtered, e)
			}
		}
		items = filtered
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	}

	if len(items) == 0 {
		fmt.Println("No elements found.")
		return nil
	}

	rows := make([][]string, len(items))
	for i, e := range items {
		rows[i] = []string{e.ID, e.Type, e.Layer, e.Name}
	}
	table([]string{"ID", "TYPE", "LAYER", "NAME"}, rows)
	fmt.Printf("\n%d–%d of %d\n", (result.Page-1)*result.Limit+1, (result.Page-1)*result.Limit+len(result.Items), result.Total)
	return nil
}

func elementGet(id string, args []string) error {
	fs := flag.NewFlagSet("element get", flag.ContinueOnError)
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

	resp, err := client.Do(http.MethodGet, fmt.Sprintf("/workspaces/%s/elements/%s", *wsID, id), nil)
	if err != nil {
		return err
	}

	var detail struct {
		Element    element               `json:"element"`
		Properties map[string][]property `json:"properties"`
	}
	if err := client.decode(resp, &detail); err != nil {
		return err
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(detail)
	}

	e := detail.Element
	fmt.Printf("ID:            %s\n", e.ID)
	fmt.Printf("Name:          %s\n", e.Name)
	fmt.Printf("Type:          %s\n", e.Type)
	fmt.Printf("Layer:         %s\n", e.Layer)
	fmt.Printf("Source ID:     %s\n", e.SourceID)
	if e.Documentation != "" {
		fmt.Printf("Documentation: %s\n", e.Documentation)
	}
	fmt.Printf("Version:       %d\n", e.Version)
	fmt.Printf("Updated:       %s\n", e.UpdatedAt.Local().Format("2006-01-02 15:04"))

	if len(detail.Properties) > 0 {
		fmt.Println("\nProperties:")
		for source, props := range detail.Properties {
			fmt.Printf("  [%s]\n", source)
			for _, p := range props {
				fmt.Printf("    %-30s  %s\n", p.Key, p.Value)
			}
		}
	}
	return nil
}

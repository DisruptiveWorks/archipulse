package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type relationship struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	SourceID      string    `json:"source_id"`
	Type          string    `json:"type"`
	SourceElement string    `json:"source_element"`
	TargetElement string    `json:"target_element"`
	Name          string    `json:"name"`
	Documentation string    `json:"documentation"`
	AccessType    string    `json:"access_type,omitempty"`
	IsDirected    bool      `json:"is_directed,omitempty"`
	Modifier      string    `json:"modifier,omitempty"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type relationshipPage struct {
	Items []relationship `json:"items"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// RunRelationship handles `archipulse relationship <subcommand>`.
func RunRelationship(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: archipulse relationship <list|get> --workspace <id>")
	}
	switch args[0] {
	case "list":
		return relationshipList(args[1:])
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: archipulse relationship get <id> --workspace <wsID>")
		}
		return relationshipGet(args[1], args[2:])
	default:
		return fmt.Errorf("unknown subcommand %q — use list or get", args[0])
	}
}

func relationshipList(args []string) error {
	fs := flag.NewFlagSet("relationship list", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	wsID := fs.String("workspace", "", "Workspace ID (required)")
	page := fs.Int("page", 1, "Page number")
	limit := fs.Int("limit", 50, "Results per page")
	typeFilter := fs.String("type", "", "Filter by relationship type")
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

	path := fmt.Sprintf("/workspaces/%s/relationships?page=%d&limit=%d", *wsID, *page, *limit)
	resp, err := client.do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	var result relationshipPage
	if err := client.decode(resp, &result); err != nil {
		return err
	}

	items := result.Items
	if *typeFilter != "" {
		filtered := items[:0]
		for _, r := range items {
			if r.Type == *typeFilter {
				filtered = append(filtered, r)
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
		fmt.Println("No relationships found.")
		return nil
	}

	rows := make([][]string, len(items))
	for i, r := range items {
		name := r.Name
		if name == "" {
			name = "-"
		}
		rows[i] = []string{r.ID, r.Type, r.SourceElement, r.TargetElement, name}
	}
	table([]string{"ID", "TYPE", "SOURCE", "TARGET", "NAME"}, rows)
	fmt.Printf("\n%d–%d of %d\n", (result.Page-1)*result.Limit+1, (result.Page-1)*result.Limit+len(result.Items), result.Total)
	return nil
}

func relationshipGet(id string, args []string) error {
	fs := flag.NewFlagSet("relationship get", flag.ContinueOnError)
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

	resp, err := client.do(http.MethodGet, fmt.Sprintf("/workspaces/%s/relationships/%s", *wsID, id), nil)
	if err != nil {
		return err
	}

	var r relationship
	if err := client.decode(resp, &r); err != nil {
		return err
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	}

	fmt.Printf("ID:             %s\n", r.ID)
	fmt.Printf("Type:           %s\n", r.Type)
	fmt.Printf("Source element: %s\n", r.SourceElement)
	fmt.Printf("Target element: %s\n", r.TargetElement)
	if r.Name != "" {
		fmt.Printf("Name:           %s\n", r.Name)
	}
	if r.Documentation != "" {
		fmt.Printf("Documentation:  %s\n", r.Documentation)
	}
	if r.AccessType != "" {
		fmt.Printf("Access type:    %s\n", r.AccessType)
	}
	if r.Modifier != "" {
		fmt.Printf("Modifier:       %s\n", r.Modifier)
	}
	fmt.Printf("Directed:       %v\n", r.IsDirected)
	fmt.Printf("Version:        %d\n", r.Version)
	fmt.Printf("Updated:        %s\n", r.UpdatedAt.Local().Format("2006-01-02 15:04"))
	return nil
}

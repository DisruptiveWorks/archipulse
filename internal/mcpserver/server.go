// Package mcpserver implements the ArchiPulse MCP server.
// It exposes ArchiPulse workspace data as tools that Claude can call.
// Run as: archipulse mcp (reads ARCHIPULSE_SERVER + ARCHIPULSE_TOKEN env vars)
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/DisruptiveWorks/archipulse/internal/cli"
	"github.com/DisruptiveWorks/archipulse/internal/cliconfig"
)

// Serve builds and starts the MCP server over stdin/stdout.
// Connection params are resolved from env vars and the config file.
func Serve(contextOverride string) error {
	r := cliconfig.Resolve(cliconfig.ResolveOptions{Context: contextOverride})
	if r.Token == "" {
		return fmt.Errorf("not authenticated — set ARCHIPULSE_TOKEN or run 'archipulse login'")
	}
	client := cli.NewClient(r.Server, r.Token)

	s := server.NewMCPServer(
		"archipulse",
		"1.0.0",
		server.WithDescription("ArchiPulse — ArchiMate workspace management"),
	)

	registerTools(s, client)

	return server.ServeStdio(s)
}

func registerTools(s *server.MCPServer, client *cli.Client) {
	// ── Workspaces ────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("list_workspaces",
		mcp.WithDescription("List all ArchiMate workspaces the current user has access to"),
	), func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return jsonTool(client, http.MethodGet, "/workspaces", nil)
	})

	s.AddTool(mcp.NewTool("get_workspace",
		mcp.WithDescription("Get details of a specific workspace"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		return jsonTool(client, http.MethodGet, "/workspaces/"+id, nil)
	})

	// ── Elements ──────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("list_elements",
		mcp.WithDescription("List ArchiMate elements in a workspace (paginated)"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page (default 50, max 500)")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		page := req.GetInt("page", 1)
		limit := req.GetInt("limit", 50)
		path := fmt.Sprintf("/workspaces/%s/elements?page=%d&limit=%d", id, page, limit)
		return jsonTool(client, http.MethodGet, path, nil)
	})

	s.AddTool(mcp.NewTool("get_element",
		mcp.WithDescription("Get an ArchiMate element with all its properties grouped by source"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithString("element_id", mcp.Required(), mcp.Description("Element UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsID, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		elID, err := req.RequireString("element_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/elements/%s", wsID, elID), nil)
	})

	// ── Relationships ─────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("list_relationships",
		mcp.WithDescription("List ArchiMate relationships in a workspace (paginated)"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page (default 50, max 500)")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		page := req.GetInt("page", 1)
		limit := req.GetInt("limit", 50)
		path := fmt.Sprintf("/workspaces/%s/relationships?page=%d&limit=%d", id, page, limit)
		return jsonTool(client, http.MethodGet, path, nil)
	})

	s.AddTool(mcp.NewTool("get_relationship",
		mcp.WithDescription("Get a specific ArchiMate relationship"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithString("relationship_id", mcp.Required(), mcp.Description("Relationship UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsID, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		relID, err := req.RequireString("relationship_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/relationships/%s", wsID, relID), nil)
	})

	// ── Diagrams ──────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("list_diagrams",
		mcp.WithDescription("List diagrams in a workspace (paginated)"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page (default 50, max 500)")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		page := req.GetInt("page", 1)
		limit := req.GetInt("limit", 50)
		path := fmt.Sprintf("/workspaces/%s/diagrams?page=%d&limit=%d", id, page, limit)
		return jsonTool(client, http.MethodGet, path, nil)
	})

	s.AddTool(mcp.NewTool("get_diagram",
		mcp.WithDescription("Get a specific diagram"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithString("diagram_id", mcp.Required(), mcp.Description("Diagram UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsID, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		dID, err := req.RequireString("diagram_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/diagrams/%s", wsID, dID), nil)
	})

	// ── Views ─────────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("get_capability_tree",
		mcp.WithDescription("Get the capability tree — hierarchical breakdown of business capabilities"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/capability-tree/tree", id), nil)
	})

	s.AddTool(mcp.NewTool("get_application_dashboard",
		mcp.WithDescription("Get application statistics dashboard for a workspace"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/application-dashboard/stats", id), nil)
	})

	s.AddTool(mcp.NewTool("get_dependency_graph",
		mcp.WithDescription("Get the application dependency graph — nodes and edges showing how applications depend on each other"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/application-dependency/graph", id), nil)
	})

	// ── Events ────────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("list_events",
		mcp.WithDescription("List workspace audit events (newest first)"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithNumber("limit", mcp.Description("Number of events to return (default 20)")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("workspace_id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		limit := req.GetInt("limit", 20)
		path := fmt.Sprintf("/workspaces/%s/events?limit=%d", id, limit)
		return jsonTool(client, http.MethodGet, path, nil)
	})
}

// jsonTool calls the API and returns the response body as pretty-printed JSON text.
func jsonTool(client *cli.Client, method, path string, body any) (*mcp.CallToolResult, error) {
	resp, err := client.Do(method, path, body)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("request failed: %v", err)), nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if e.Error != "" {
			return mcp.NewToolResultText(fmt.Sprintf("error: %s", e.Error)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("HTTP %d", resp.StatusCode)), nil
	}

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("decode error: %v", err)), nil
	}
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return mcp.NewToolResultText(string(raw)), nil
	}
	return mcp.NewToolResultText(string(pretty)), nil
}

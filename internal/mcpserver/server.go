// Package mcpserver implements the ArchiPulse MCP server.
// It exposes ArchiPulse workspace data as tools that Claude can call.
// Run as: archipulse mcp (reads ARCHIPULSE_SERVER + ARCHIPULSE_TOKEN env vars)
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
	registerResources(s)

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
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet, "/workspaces/"+id, nil)
	})

	s.AddTool(mcp.NewTool("create_workspace",
		mcp.WithDescription("Create a new ArchiMate workspace"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Workspace name")),
		mcp.WithString("purpose", mcp.Required(), mcp.Description("Modelling purpose: 'as-is', 'to-be', or 'migration'")),
		mcp.WithString("description", mcp.Description("Optional description")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, errResult := requireID(req, "name")
		if errResult != nil {
			return errResult, nil
		}
		purpose, errResult := requireID(req, "purpose")
		if errResult != nil {
			return errResult, nil
		}
		desc, _ := req.RequireString("description")
		body := map[string]string{"name": name, "purpose": purpose, "description": desc}
		return jsonTool(client, http.MethodPost, "/workspaces", body)
	})

	// ── Elements ──────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("list_elements",
		mcp.WithDescription("List ArchiMate elements in a workspace (paginated)"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page (default 50, max 500)")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
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
		wsID, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		elID, errResult := requireID(req, "element_id")
		if errResult != nil {
			return errResult, nil
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
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
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
		wsID, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		relID, errResult := requireID(req, "relationship_id")
		if errResult != nil {
			return errResult, nil
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
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
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
		wsID, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		dID, errResult := requireID(req, "diagram_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/diagrams/%s", wsID, dID), nil)
	})

	// ── Views ─────────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("get_capability_tree",
		mcp.WithDescription("Get the capability tree — hierarchical breakdown of business capabilities"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/capability-tree/tree", id), nil)
	})

	s.AddTool(mcp.NewTool("get_application_dashboard",
		mcp.WithDescription("Get application statistics dashboard for a workspace"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/application-dashboard/stats", id), nil)
	})

	s.AddTool(mcp.NewTool("get_dependency_graph",
		mcp.WithDescription("Get the application dependency graph — nodes and edges showing how applications depend on each other"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/application-dependency/graph", id), nil)
	})

	// ── Catalogues & matrices ─────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("get_integration_map",
		mcp.WithDescription("Get the integration map — application-to-application data flows and interfaces"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/integration-map/graph", id), nil)
	})

	s.AddTool(mcp.NewTool("get_app_catalogue",
		mcp.WithDescription("Get the application catalogue — all applications with lifecycle status, capabilities, and technology details"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/application-catalogue/entries", id), nil)
	})

	s.AddTool(mcp.NewTool("get_tech_catalogue",
		mcp.WithDescription("Get the technology catalogue — all technology components with adoption status and supporting applications"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/technology-catalogue/entries", id), nil)
	})

	s.AddTool(mcp.NewTool("get_process_application_matrix",
		mcp.WithDescription("Get the process-to-application matrix — which business processes are supported by which applications"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/views/process-application/matrix", id), nil)
	})

	s.AddTool(mcp.NewTool("get_app_detail",
		mcp.WithDescription("Get full detail for an application element — capabilities it supports, dependencies, and diagrams it appears in"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithString("element_id", mcp.Required(), mcp.Description("Application element UUID")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsID, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		elID, errResult := requireID(req, "element_id")
		if errResult != nil {
			return errResult, nil
		}
		return jsonTool(client, http.MethodGet,
			fmt.Sprintf("/workspaces/%s/elements/%s/app-detail", wsID, elID), nil)
	})

	// ── Import ────────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("preview_import",
		mcp.WithDescription("Preview an AOEF XML import — returns a diff of what would be created, updated, or deleted without writing any changes. Use this before import_model to let the user confirm."),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithString("xml_content", mcp.Required(), mcp.Description("Complete AOEF XML document to preview")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsID, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		xml, errResult := requireID(req, "xml_content")
		if errResult != nil {
			return errResult, nil
		}
		return multipartTool(client, fmt.Sprintf("/workspaces/%s/import/preview", wsID), strings.NewReader(xml))
	})

	s.AddTool(mcp.NewTool("import_model",
		mcp.WithDescription("Import an AOEF XML document into a workspace. Creates or updates elements, relationships, and diagrams. Read the aoef://format-guide resource before generating XML. For large models (200+ elements) omit <views> on the first import and add diagrams in a second import — this keeps the XML size manageable."),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithString("xml_content", mcp.Required(), mcp.Description("Complete AOEF XML document to import")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsID, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		xml, errResult := requireID(req, "xml_content")
		if errResult != nil {
			return errResult, nil
		}
		return multipartTool(client, fmt.Sprintf("/workspaces/%s/import", wsID), strings.NewReader(xml))
	})

	// ── Events ────────────────────────────────────────────────────────────────

	s.AddTool(mcp.NewTool("list_events",
		mcp.WithDescription("List workspace audit events (newest first)"),
		mcp.WithString("workspace_id", mcp.Required(), mcp.Description("Workspace UUID")),
		mcp.WithNumber("limit", mcp.Description("Number of events to return (default 20)")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, errResult := requireID(req, "workspace_id")
		if errResult != nil {
			return errResult, nil
		}
		limit := req.GetInt("limit", 20)
		path := fmt.Sprintf("/workspaces/%s/events?limit=%d", id, limit)
		return jsonTool(client, http.MethodGet, path, nil)
	})
}

func registerResources(s *server.MCPServer) {
	s.AddResource(
		mcp.NewResource(
			"aoef://format-guide",
			"AOEF Format Guide",
			mcp.WithResourceDescription("Complete guide for generating valid AOEF XML: element types, relationship types, XML structure, and examples"),
			mcp.WithMIMEType("text/markdown"),
		),
		func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "aoef://format-guide",
					MIMEType: "text/markdown",
					Text:     aoefFormatGuide,
				},
			}, nil
		},
	)
}

// multipartTool posts an XML reader as a multipart upload and returns the JSON response.
func multipartTool(client *cli.Client, path string, content io.Reader) (*mcp.CallToolResult, error) {
	resp, err := client.DoMultipart(path, "file", content)
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

// requireID extracts a required UUID string argument, returning an error result if missing or empty.
func requireID(req mcp.CallToolRequest, key string) (string, *mcp.CallToolResult) {
	val, err := req.RequireString(key)
	if err != nil {
		return "", mcp.NewToolResultText(err.Error())
	}
	if val == "" {
		return "", mcp.NewToolResultText(fmt.Sprintf("%s is required", key))
	}
	return val, nil
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

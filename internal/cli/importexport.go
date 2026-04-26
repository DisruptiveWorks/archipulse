package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// RunImport handles `archipulse import --workspace <id> <file.xml>`.
func RunImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	wsID := fs.String("workspace", "", "Workspace ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *wsID == "" {
		return fmt.Errorf("--workspace is required")
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: archipulse import --workspace <id> <file.xml>")
	}
	filePath := fs.Arg(0)

	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		part, err := mw.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, f); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = mw.Close()
		_ = pw.Close()
	}()

	req, err := http.NewRequest(http.MethodPost,
		client.server+"/api/v1/workspaces/"+*wsID+"/import", pr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if client.token != "" {
		req.Header.Set("Cookie", "ap_session="+client.token)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	var result struct {
		WorkspaceID         string `json:"workspace_id"`
		Elements            int    `json:"elements"`
		Relationships       int    `json:"relationships"`
		Diagrams            int    `json:"diagrams"`
		Folders             int    `json:"folders"`
		PropertyDefinitions int    `json:"property_definitions"`
	}
	if err := client.decode(resp, &result); err != nil {
		return err
	}

	if g.output == "json" {
		enc := json.NewEncoder(stdoutWriter())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Printf("Import complete\n")
	fmt.Printf("  Elements:             %d\n", result.Elements)
	fmt.Printf("  Relationships:        %d\n", result.Relationships)
	fmt.Printf("  Diagrams:             %d\n", result.Diagrams)
	fmt.Printf("  Folders:              %d\n", result.Folders)
	fmt.Printf("  Property definitions: %d\n", result.PropertyDefinitions)
	return nil
}

// RunExport handles `archipulse export --workspace <id> [--out file.xml]`.
func RunExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	var g globalFlags
	addGlobalFlags(fs, &g)
	wsID := fs.String("workspace", "", "Workspace ID (required)")
	outFile := fs.String("out", "", "Output file path (default: stdout)")
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

	resp, err := client.do(http.MethodGet, "/workspaces/"+*wsID+"/export/aoef", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if e.Error != "" {
			return fmt.Errorf("%s", e.Error)
		}
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var out io.Writer
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer func() { _ = f.Close() }()
		out = f
	} else {
		out = os.Stdout
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("write export: %w", err)
	}

	if *outFile != "" {
		fmt.Fprintf(os.Stderr, "Exported %d bytes → %s\n", n, *outFile)
	}
	return nil
}

package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/api"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

func TestImport_AOEF_ArchiSurance(t *testing.T) {
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)

	ws, err := wsStore.Create("import-archisurance", "as-is", "")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	body, ct := multipartFile(t, fixture("archisurance.xml"))
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/workspaces/"+ws.ID.String()+"/import", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	api.NewRouter(conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("import status = %d, body = %s", rr.Code, rr.Body.String())
	}
	t.Logf("import result: %s", rr.Body.String())
}

func TestImport_InvalidXML(t *testing.T) {
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)

	ws, err := wsStore.Create("import-invalid", "as-is", "")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "bad.xml")
	_, _ = fw.Write([]byte("not xml at all"))
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/workspaces/"+ws.ID.String()+"/import", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	api.NewRouter(conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestImport_WorkspaceNotFound(t *testing.T) {
	conn := openTestDB(t)

	body, ct := multipartFile(t, fixture("minimal.xml"))
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/workspaces/"+nonExistentUUID().String()+"/import", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	api.NewRouter(conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// multipartFile builds a multipart request body from a file on disk.
func multipartFile(t *testing.T, path string) (*bytes.Buffer, string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(fw, f); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// setupArchiSurance imports the ArchiSurance model into a test workspace
// and returns the workspace ID. Cleanup is registered automatically.
func setupArchiSurance(t *testing.T) string {
	t.Helper()
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)

	ws, err := wsStore.Create("viewer-test-"+t.Name(), "as-is", "")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	body, ct := multipartFile(t, fixture("archisurance.xml"))
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/workspaces/"+ws.ID.String()+"/import", body)
	req.Header.Set("Content-Type", ct)
	addAuthCookie(t, req)
	rr := httptest.NewRecorder()
	testRouter(t, conn).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("import failed: %d %s", rr.Code, rr.Body.String())
	}
	return ws.ID.String()
}

func TestViewer_ListViews(t *testing.T) {
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)
	ws, err := wsStore.Create("viewer-list-test", "as-is", "")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/workspaces/"+ws.ID.String()+"/views", nil)
	addAuthCookie(t, req)
	rr := httptest.NewRecorder()
	testRouter(t, conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("list views status = %d", rr.Code)
	}
	var result map[string][]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result["views"]) == 0 {
		t.Error("expected at least one view in registry")
	}
	t.Logf("available views: %v", result["views"])
}

func TestViewer_ElementCatalogue(t *testing.T) {
	wsID := setupArchiSurance(t)
	conn := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/workspaces/"+wsID+"/views/element-catalogue", nil)
	addAuthCookie(t, req)
	rr := httptest.NewRecorder()
	testRouter(t, conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("element-catalogue status = %d: %s", rr.Code, rr.Body.String())
	}
	var view struct {
		Name    string   `json:"name"`
		Columns []string `json:"columns"`
		Rows    [][]any  `json:"rows"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&view); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(view.Rows) < 100 {
		t.Errorf("expected at least 100 rows (ArchiSurance has 120 elements), got %d", len(view.Rows))
	}
	t.Logf("element-catalogue: %d rows", len(view.Rows))
}

func TestViewer_ApplicationCatalogue(t *testing.T) {
	wsID := setupArchiSurance(t)
	conn := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/workspaces/"+wsID+"/views/application-catalogue", nil)
	addAuthCookie(t, req)
	rr := httptest.NewRecorder()
	testRouter(t, conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("application-catalogue status = %d", rr.Code)
	}
	var view struct {
		Rows [][]any `json:"rows"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&view); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(view.Rows) == 0 {
		t.Error("expected application elements in ArchiSurance")
	}
	t.Logf("application-catalogue: %d rows", len(view.Rows))
}

func TestViewer_ApplicationLandscape(t *testing.T) {
	wsID := setupArchiSurance(t)
	conn := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/workspaces/"+wsID+"/views/application-landscape", nil)
	addAuthCookie(t, req)
	rr := httptest.NewRecorder()
	testRouter(t, conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("application-landscape status = %d", rr.Code)
	}
	t.Logf("application-landscape: %s", rr.Body.String()[:min(200, rr.Body.Len())])
}

func TestViewer_DependencyGraph(t *testing.T) {
	wsID := setupArchiSurance(t)
	conn := openTestDB(t)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/workspaces/"+wsID+"/views/application-dependency/graph", nil)
	addAuthCookie(t, req)
	rr := httptest.NewRecorder()
	testRouter(t, conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("dependency graph status = %d: %s", rr.Code, rr.Body.String())
	}
	var graph struct {
		Nodes []any `json:"nodes"`
		Edges []any `json:"edges"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&graph); err != nil {
		t.Fatalf("decode: %v", err)
	}
	t.Logf("dependency graph: %d nodes, %d edges", len(graph.Nodes), len(graph.Edges))
}

func TestViewer_UnknownView(t *testing.T) {
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)
	ws, _ := wsStore.Create("viewer-404-test", "as-is", "")
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/workspaces/"+ws.ID.String()+"/views/nonexistent-view", nil)
	addAuthCookie(t, req)
	rr := httptest.NewRecorder()
	testRouter(t, conn).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

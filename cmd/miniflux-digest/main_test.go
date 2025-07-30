package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestArchive(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "miniflux-digest-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	})

	// Create a dummy file to serve
	categoryDir := filepath.Join(tmpDir, "test-category")
	if err := os.Mkdir(categoryDir, 0755); err != nil {
		t.Fatalf("Failed to create category dir: %v", err)
	}
	filePath := filepath.Join(categoryDir, "test-file.html")
	fileContent := "<html><body><h1>Test File</h1></body></html>"
	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return tmpDir
}

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthcheck", nil)
	rr := httptest.NewRecorder()
	mux := SetupServer("") // archive base path is not needed for this test
	h := requestSanitizerMiddleware(mux)
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestServeArchiveFile_Success(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := SetupServer(archiveBasePath)

	req := httptest.NewRequest("GET", "/archive/test-category/test-file.html", nil)
	rr := httptest.NewRecorder()
	h := requestSanitizerMiddleware(mux)
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "<html><body><h1>Test File</h1></body></html>"
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != expected {
		t.Errorf("handler returned unexpected body: got %q want %q",
			string(body), expected)
	}
}

func TestServeArchiveFile_NotFound(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := SetupServer(archiveBasePath)

	req := httptest.NewRequest("GET", "/archive/test-category/not-found.html", nil)
	rr := httptest.NewRecorder()
	h := requestSanitizerMiddleware(mux)
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestServeArchiveFile_PathTraversal(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := SetupServer(archiveBasePath)

	// Attempt to access a file outside the archive base path
	// The http.FileServer should prevent this, resulting in a 400
	req := httptest.NewRequest("GET", "/archive/../main_test.go", nil)
	rr := httptest.NewRecorder()
	h := requestSanitizerMiddleware(mux)
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for path traversal attempt: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestServeArchiveFile_DirectoryRequest(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := SetupServer(archiveBasePath)

	req := httptest.NewRequest("GET", "/archive/test-category/", nil)
	rr := httptest.NewRecorder()
	h := requestSanitizerMiddleware(mux)
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code for directory request: got %v want %v",
			status, http.StatusNotFound)
	}
}

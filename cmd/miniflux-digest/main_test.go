package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"miniflux-digest/internal/webserver"
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

	// Create a directory with an index.html
	categoryWithIndexDir := filepath.Join(tmpDir, "test-category-with-index")
	if err := os.Mkdir(categoryWithIndexDir, 0755); err != nil {
		t.Fatalf("Failed to create category with index dir: %v", err)
	}
	indexPath := filepath.Join(categoryWithIndexDir, "index.html")
	indexContent := "<html><body><h1>Index File</h1></body></html>"
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write index file: %v", err)
	}

	// Create a directory without an index.html
	categoryNoIndexDir := filepath.Join(tmpDir, "test-category-no-index")
	if err := os.Mkdir(categoryNoIndexDir, 0755); err != nil {
		t.Fatalf("Failed to create category no index dir: %v", err)
	}

	return tmpDir
}

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthcheck", nil)
	rr := httptest.NewRecorder()
	mux := webserver.SetupServer("") // archive base path is not needed for this test
	mux.ServeHTTP(rr, req)

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
	mux := webserver.SetupServer(archiveBasePath)

	req := httptest.NewRequest("GET", "/archive/test-category/test-file.html", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

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
	mux := webserver.SetupServer(archiveBasePath)

	req := httptest.NewRequest("GET", "/archive/test-category/not-found.html", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestServeArchiveFile_PathTraversal(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := webserver.SetupServer(archiveBasePath)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Do not follow redirects automatically
		},
	}

	// Attempt to access a file outside the archive base path
	req, err := http.NewRequest("GET", ts.URL+"/archive/../main_test.go", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	// The first response should be a 301 redirect
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("handler returned wrong status code for path traversal attempt: got %v want %v",
			resp.StatusCode, http.StatusMovedPermanently)
	}

	// Now, follow the redirect and expect a 404
	redirectURL, err := resp.Location()
	if err != nil {
		t.Fatalf("Failed to get redirect location: %v", err)
	}

	req, err = http.NewRequest("GET", redirectURL.String(), nil)
	if err != nil {
		t.Fatalf("Failed to create redirect request: %v", err)
	}

	// Use a client that follows redirects for the second request
	clientWithRedirect := &http.Client{}
	resp, err = clientWithRedirect.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform redirect request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("handler returned wrong status code after redirect for path traversal attempt: got %v want %v",
			resp.StatusCode, http.StatusNotFound)
	}
}

func TestServeArchiveFile_DirectoryListingDisabled(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	handler := webserver.SetupServer(archiveBasePath)

	// Create a directory without an index.html file
	if err := os.Mkdir(filepath.Join(archiveBasePath, "empty-dir"), 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	req := httptest.NewRequest("GET", "/archive/empty-dir/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code for directory listing attempt: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestServeArchiveFile_NoDirectoryListingAndIndexHtml(t *testing.T) {
	archiveBasePath := setupTestArchive(t)
	mux := webserver.SetupServer(archiveBasePath)

	// Test case 1: Request a directory with index.html
	req1 := httptest.NewRequest("GET", "/archive/test-category-with-index/", nil)
	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, req1)

	if status := rr1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for directory with index.html: got %v want %v",
			status, http.StatusOK)
	}

	expected1 := "<html><body><h1>Index File</h1></body></html>"
	body1, err := io.ReadAll(rr1.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body1) != expected1 {
		t.Errorf("handler returned unexpected body for directory with index.html: got %q want %q",
			string(body1), expected1)
	}

	// Test case 2: Request a directory without index.html
	req2 := httptest.NewRequest("GET", "/archive/test-category-no-index/", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)

	if status := rr2.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code for directory without index.html: got %v want %v",
			status, http.StatusNotFound)
	}
}

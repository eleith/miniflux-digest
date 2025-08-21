package webserver

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
)

const (
	ArchiveBasePath = "web/miniflux-archive"
	Port            = ":8080"
)

// noDirListingFileServer wraps http.FileServer to prevent directory listings.
func noDirListingFileServer(root http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path ends with a slash (indicating a directory)
		if strings.HasSuffix(r.URL.Path, "/") {
			// Try to open index.html within that directory
			indexPath := path.Join(r.URL.Path, "index.html")
			if _, err := root.Open(indexPath); err != nil {
				// If index.html doesn't exist, return 404
				http.NotFound(w, r)
				return
			}
			// If index.html exists, let FileServer handle it (it will redirect to /path/to/dir/index.html)
		}
		http.FileServer(root).ServeHTTP(w, r)
	})
}

func SetupServer(archiveBasePath string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, "OK"); err != nil {
			log.Printf("Error writing healthcheck response: %v", err)
		}
	})

	// Use the new noDirListingFileServer
	mux.Handle("/archive/", http.StripPrefix("/archive/", noDirListingFileServer(http.Dir(archiveBasePath))))

	return mux
}

func StartServer(port string, handler http.Handler) {
	log.Printf("Internal web server starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Internal web server failed to start: %v", err)
	}
}

func ListenAndServe(archiveBasePath string, port string) *http.ServeMux {
	mux := SetupServer(archiveBasePath)
	log.Printf("Internal web server starting on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Internal web server failed to start: %v", err)
	}

	return mux
}

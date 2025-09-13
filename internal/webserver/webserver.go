package webserver

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	ArchiveBasePath = "web/archive"
	StaticBasePath  = "web/static"
	Port            = ":8080"
)

type noDirListingFileSystem struct {
	fs http.FileSystem
}

func (fs noDirListingFileSystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		_ = f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

func SetupServer(archiveBasePath string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, "OK"); err != nil {
			log.Printf("Error writing healthcheck response: %v", err)
		}
	})

	staticFs := http.FileServer(http.Dir(StaticBasePath))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFs))

	fs := http.FileServer(noDirListingFileSystem{http.Dir(archiveBasePath)})

	mux.Handle("/archive/", http.StripPrefix("/archive/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			strippedPath := strings.TrimPrefix(r.URL.Path, "/")
			indexPath := path.Join(strippedPath, "index.html")

			if _, err := (noDirListingFileSystem{http.Dir(archiveBasePath)}).Open(indexPath); err == nil {
				http.ServeFile(w, r, filepath.Join(archiveBasePath, indexPath))
				return
			}
		}
		fs.ServeHTTP(w, r)
	})))

	return mux
}

func ListenAndServe(archiveBasePath string, port string) *http.ServeMux {
	mux := SetupServer(archiveBasePath)

	log.Printf("Internal web server starting on port %s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Internal web server failed to start: %v", err)
	}

	return mux
}

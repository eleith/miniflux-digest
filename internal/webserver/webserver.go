package webserver

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	ArchiveBasePath = "web/miniflux-archive"
	Port = ":8080"
)

func SetupServer(archiveBasePath string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, "OK"); err != nil {
			log.Printf("Error writing healthcheck response: %v", err)
		}
	})

	fs := http.FileServer(http.Dir(archiveBasePath))
	mux.Handle("/archive/", http.StripPrefix("/archive/", fs))

	return mux
}

func RequestSanitizerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func StartServer(port string, handler http.Handler) {
	log.Printf("Internal web server starting on port %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Internal web server failed to start: %v", err)
	}
}

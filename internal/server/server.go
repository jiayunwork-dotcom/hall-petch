// Package server wires the Hall-Petch kernel to a small HTTP API and serves the
// static front-end from the same Go process. The API is intentionally thin: it
// only validates input, delegates to the kernel packages, and returns JSON with
// descriptive errors. All computation happens in internal/core, internal/scale
// and internal/line; this package is pure wiring.
package server

import (
	"net/http"
	"os"
)

// Server holds the directories the HTTP layer serves.
type Server struct {
	webDir      http.FileSystem
	exampleDir  *osDir
	apiMux      *http.ServeMux
}

// osDir wraps an os.ReadDir-able directory handle.
type osDir struct {
	path string
}

// ReadDir lists the entries of the example directory.
func (d *osDir) ReadDir(n int) ([]os.DirEntry, error) {
	return os.ReadDir(d.path)
}

// NewServer builds a Server that serves the front-end from webDir and the
// example files from exampleDir.
func NewServer(webDir, exampleDir string) *Server {
	s := &Server{
		webDir:     http.Dir(webDir),
		exampleDir: &osDir{path: exampleDir},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sy", handleSy)
	mux.HandleFunc("/api/line", handleLine)
	mux.HandleFunc("/api/examples", s.handleExampleList)
	// Example JSON files are served so the front-end can fetch them directly.
	mux.Handle("/example/", serveStatic("/example/", exampleDir))
	// Everything else under the web root (index.html, app.js, ...).
	mux.Handle("/", serveStatic("/", webDir))
	s.apiMux = mux
	return s
}

// Handler returns the configured HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.apiMux
}

// ListenAndServe starts the HTTP server on addr (e.g. ":8080").
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.apiMux)
}

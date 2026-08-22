package server

import (
	"net/http"
)

// serveStatic builds an http.Handler that serves files from root under the
// given URL prefix. A request for the prefix alone returns the directory's
// index file (e.g. web/index.html), which is what the single-page front-end
// needs at "/".
func serveStatic(prefix, root string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(root)))
}

// handleExampleList returns the list of example files available under the
// example directory, so the front-end can offer a "load example" picker.
func (s *Server) handleExampleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET for /api/examples")
		return
	}
	names, err := s.exampleDir.ReadDir(0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot read examples: "+err.Error())
		return
	}
	items := make([]string, 0, len(names))
	for _, n := range names {
		if n.IsDir() {
			continue
		}
		items = append(items, n.Name())
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "examples": items})
}

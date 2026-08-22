package server

import (
	"encoding/json"
	"net/http"
)

// errorResponse is the JSON body returned for any failed API request.
type errorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// writeJSON writes a value as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a structured error body and the supplied HTTP status. The
// message comes straight from the kernel so the front-end can show exactly what
// went wrong (invalid diameter, coefficient, etc.) instead of a blank 200.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{OK: false, Error: msg})
}

// decodeJSON reads the request body into the target value, returning a 400 with
// a descriptive message when the payload is not valid JSON.
func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

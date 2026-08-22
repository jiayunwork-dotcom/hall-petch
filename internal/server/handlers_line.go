package server

import (
	"net/http"

	"hall-petch/internal/core"
	"hall-petch/internal/line"
)

// lineRequest is the JSON body for POST /api/line.
type lineRequest struct {
	Sigma0 float64 `json:"sigma0"`
	Ky     float64 `json:"ky"`
	DMin   float64 `json:"d_min"`
	DMax   float64 `json:"d_max"`
	Steps  int     `json:"steps"`
	DUnit  string  `json:"d_unit"`
}

// linePoint is one element of the point series returned by /api/line.
type linePoint struct {
	D        float64 `json:"d"`
	DSqrtInv float64 `json:"d_sqrt_inv"`
	SigmaY   float64 `json:"sigma_y"`
}

// lineResponse is the JSON body returned by POST /api/line.
type lineResponse struct {
	OK     bool        `json:"ok"`
	Points []linePoint `json:"points"`
}

// handleLine scans the Hall-Petch curve across a diameter range and returns the
// ordered point series for plotting.
//
//	POST /api/line { "sigma0", "ky", "d_min", "d_max", "steps", "d_unit" }
//	    -> { "points": [ { "d", "d_sqrt_inv", "sigma_y" }, ... ] }
//
// Out-of-range or degenerate scan parameters produce a 400 error body.
func handleLine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST for /api/line")
		return
	}
	var req lineRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	unit := req.DUnit
	if unit == "" {
		unit = "um"
	}
	steps := req.Steps
	if steps <= 0 {
		steps = 40
	}
	pts, err := line.ScanLine(line.ScanRequest{
		Sigma0: req.Sigma0,
		Ky:     req.Ky,
		DMin:   req.DMin,
		DMax:   req.DMax,
		Steps:  steps,
		DUnit:  unit,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, core.Message(err))
		return
	}
	out := make([]linePoint, 0, len(pts))
	for _, p := range pts {
		out = append(out, linePoint{
			D:        p.D,
			DSqrtInv: p.DSqrtInv,
			SigmaY:   p.SigmaY,
		})
	}
	writeJSON(w, http.StatusOK, lineResponse{OK: true, Points: out})
}

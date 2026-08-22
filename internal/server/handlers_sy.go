package server

import (
	"net/http"

	"hall-petch/internal/core"
)

// syRequest is the JSON body for POST /api/sy.
type syRequest struct {
	Sigma0 float64 `json:"sigma0"`
	Ky     float64 `json:"ky"`
	D      float64 `json:"d"`
	DUnit  string  `json:"d_unit"`
}

// syResponse is the JSON body returned by POST /api/sy.
type syResponse struct {
	OK       bool    `json:"ok"`
	Sigma0   float64 `json:"sigma0"`
	Ky       float64 `json:"ky"`
	D        float64 `json:"d"`
	DUnit    string  `json:"d_unit"`
	SigmaY   float64 `json:"sigma_y"`
	DSqrtInv float64 `json:"d_sqrt_inv"`
}

// handleSy computes the Hall-Petch yield strength for one grain diameter.
//
//	POST /api/sy  { "sigma0", "ky", "d", "d_unit" } -> { "sigma_y", "d_sqrt_inv" }
//
// Invalid inputs (d <= 0, ky <= 0, sigma0 < 0, unknown unit) produce a 400 with
// a descriptive error field rather than a silent empty success.
func handleSy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST for /api/sy")
		return
	}
	var req syRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	unit := req.DUnit
	if unit == "" {
		unit = "um"
	}
	params := core.Params{Sigma0: req.Sigma0, Ky: req.Ky, D: req.D, DUnit: unit}
	res, err := params.Evaluate()
	if err != nil {
		writeError(w, http.StatusBadRequest, core.Message(err))
		return
	}
	writeJSON(w, http.StatusOK, syResponse{
		OK:       true,
		Sigma0:   res.Sigma0,
		Ky:       res.Ky,
		D:        req.D,
		DUnit:    unit,
		SigmaY:   res.SigmaY,
		DSqrtInv: res.DSqrtInv,
	})
}

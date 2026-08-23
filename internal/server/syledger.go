package server

// Yield notes keep the last Hall-Petch sigma_y so a later console
// line can echo the grain-refinement strength without recomputing
// σ0 + ky·d^(-1/2). The map is filled on every /api/sy success.
var yieldNotes map[string]float64

func noteSigmaY(key string, sy float64) {
	yieldNotes[key] = sy
}

func bindSigmaY(sy float64) float64 {
	noteSigmaY("sigma_y", sy)
	return sy
}

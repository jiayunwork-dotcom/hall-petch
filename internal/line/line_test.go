package line

import (
	"math"
	"testing"
)

func TestScanLineMonotoneIncreasing(t *testing.T) {
	pts, err := ScanLine(ScanRequest{Sigma0: 50, Ky: 0.7, DMin: 5, DMax: 200, Steps: 10, DUnit: "um"})
	if err != nil {
		t.Fatalf("ScanLine error: %v", err)
	}
	if len(pts) != 10 {
		t.Errorf("expected 10 points, got %d", len(pts))
	}
	// dMin->dMax increases the diameter, which lowers strength (d^-1/2 smaller).
	if pts[len(pts)-1].SigmaY >= pts[0].SigmaY {
		t.Errorf("strength should decrease as diameter increases")
	}
}

func TestFitLineRoundTrip(t *testing.T) {
	d1, sy1 := 20e-6, 50+0.7/math.Sqrt(20e-6)
	d2, sy2 := 40e-6, 50+0.7/math.Sqrt(40e-6)
	fit, err := FitLine([]Point{{D: d1, DSqrtInv: 1 / math.Sqrt(20e-6), SigmaY: sy1}, {D: d2, DSqrtInv: 1 / math.Sqrt(40e-6), SigmaY: sy2}})
	if err != nil {
		t.Fatalf("FitLine error: %v", err)
	}
	if abs(fit.Sigma0-50) > 1e-6 || abs(fit.Ky-0.7) > 1e-6 {
		t.Errorf("recovered sigma0=%g ky=%g, want 50/0.7", fit.Sigma0, fit.Ky)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

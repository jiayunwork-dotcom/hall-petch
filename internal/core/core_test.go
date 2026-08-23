package core

import (
	"math"
	"testing"
)

func TestYieldStrengthBasic(t *testing.T) {
	// sigma0=50, ky=0.7, d=20um => d^-1/2 = 1/sqrt(20e-6) = 223.6068, *0.7 = 156.52
	dMetres := 20e-6
	got := YieldStrength(50, 0.7, dMetres)
	want := 50 + 0.7/math.Sqrt(20e-6)
	if abs(got-want) > 1e-6 {
		t.Errorf("YieldStrength = %g, want %g", got, want)
	}
	if got < 50 {
		t.Errorf("yield strength should exceed sigma0, got %g", got)
	}
}

func TestEvaluateValidatesNegativeDiameter(t *testing.T) {
	p := Params{Sigma0: 50, Ky: 0.7, D: -1, DUnit: "um"}
	if _, err := p.Evaluate(); err == nil {
		t.Errorf("expected error for negative diameter")
	}
}

func TestInverseDiameterRoundTrip(t *testing.T) {
	d, err := InverseDiameter(0.7, 50+0.7/math.Sqrt(20e-6), 50)
	if err != nil {
		t.Fatalf("InverseDiameter error: %v", err)
	}
	if abs(d-20e-6) > 1e-9 {
		t.Errorf("recovered d = %g, want 20e-6", d)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

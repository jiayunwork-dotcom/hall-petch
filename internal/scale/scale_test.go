package scale

import "testing"

func TestConvertLengthRoundTrip(t *testing.T) {
	out, err := ConvertLength(1, "mm", "um")
	if err != nil {
		t.Fatalf("ConvertLength error: %v", err)
	}
	if abs(out-1000) > 1e-9 {
		t.Errorf("1 mm -> um = %g, want 1000", out)
	}
}

func TestASTMRoundTrip(t *testing.T) {
	d, err := DiameterFromASTM(8)
	if err != nil {
		t.Fatalf("DiameterFromASTM error: %v", err)
	}
	g, err := ASTMFromDiameter(d)
	if err != nil {
		t.Fatalf("ASTMFromDiameter error: %v", err)
	}
	if abs(g-8) > 1e-6 {
		t.Errorf("ASTM round trip = %g, want 8", g)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

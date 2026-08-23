package strengthen

import (
	"math"
	"testing"

	"hall-petch/internal/core"
)

func sample() []Grain {
	return []Grain{
		{Diameter: 5, Unit: "um"},
		{Diameter: 10, Unit: "um"},
		{Diameter: 20, Unit: "um"},
	}
}

func TestNumberAverage(t *testing.T) {
	avg, err := NumberAverageDiameter(sample())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if math.Abs(avg-35.0/3.0*1e-6) > 1e-12 { // (5+10+20)/3 um = 11.667 um = 1.1667e-5 m
		t.Fatalf("number average = %v, want ~1.1667e-5", avg)
	}
}

func TestAreaAverage(t *testing.T) {
	avg, err := AreaAverageDiameter(sample())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// area-average weights d^3/d^2 -> lies above number average
	if avg <= 1e-5 {
		t.Fatalf("area average %v must exceed number average 1.1667e-5", avg)
	}
}

func TestGrainArea(t *testing.T) {
	// d = 2 m -> area = pi
	if math.Abs(GrainArea(2.0)-math.Pi) > 1e-9 {
		t.Fatal("area of d=2 is pi")
	}
}

func TestBoundaryLengthDensity(t *testing.T) {
	if math.Abs(BoundaryLengthDensity(0.1)-20.0) > 1e-9 {
		t.Fatal("S_V = 2/d = 20")
	}
}

func TestFitLine(t *testing.T) {
	// Build synthetic data from known sigma0=100, ky=0.5 (d in metres).
	ds := []Grain{{Diameter: 1e-5, Unit: "m"}, {Diameter: 4e-5, Unit: "m"}, {Diameter: 9e-5, Unit: "m"}}
	yields := make([]float64, 3)
	for i, g := range ds {
		d, _ := core.ToMetres(g.Diameter, g.Unit)
		yields[i] = 100 + 0.5/math.Sqrt(d)
	}
	s0, ky, err := FitLine(ds, yields)
	if err != nil {
		t.Fatalf("fit err: %v", err)
	}
	if math.Abs(s0-100) > 1e-6 {
		t.Fatalf("sigma0 = %v, want 100", s0)
	}
	if math.Abs(ky-0.5) > 1e-6 {
		t.Fatalf("ky = %v, want 0.5", ky)
	}
}

func TestScatterResidualZero(t *testing.T) {
	ds := []Grain{{Diameter: 1e-5, Unit: "m"}, {Diameter: 4e-5, Unit: "m"}}
	yields := []float64{100 + 0.5/math.Sqrt(1e-5), 100 + 0.5/math.Sqrt(4e-5)}
	r, err := ScatterResidual(ds, yields, 100, 0.5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if r > 1e-9 {
		t.Fatalf("residual on exact fit should be ~0, got %v", r)
	}
}

func TestDiameterForStrength(t *testing.T) {
	// sigma0=100, ky=0.5 -> for sy=200 we need (0.5/(100))^2 = 2.5e-5 m
	d, err := DiameterForStrength(100, 0.5, 200)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if math.Abs(d-2.5e-5) > 1e-12 {
		t.Fatalf("d = %v, want 2.5e-5", d)
	}
	if _, err := DiameterForStrength(100, 0.5, 50); err == nil {
		t.Fatal("target below sigma0 must error")
	}
}

func TestGrainVolume(t *testing.T) {
	v := GrainVolume(2.0)
	if math.Abs(v-4.0/3.0*math.Pi) > 1e-9 {
		t.Fatal("volume of d=2 sphere")
	}
}

func TestHarmonicDiameter(t *testing.T) {
	out, err := HarmonicDiameter(sample())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := 3.0 / (1.0/5.0 + 1.0/10.0 + 1.0/20.0) * 1e-6
	if math.Abs(out-want) > 1e-12 {
		t.Fatalf("harmonic = %v, want %v", out, want)
	}
}

func TestPredictYield(t *testing.T) {
	got, err := PredictYield(100, 0.5, 1e-5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if math.Abs(got-(100+0.5/math.Sqrt(1e-5))) > 1e-9 {
		t.Fatalf("yield = %v", got)
	}
}

func TestLogarithmicSpread(t *testing.T) {
	out, err := LogarithmicSpread(sample())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if math.Abs(out-4.0) > 1e-9 { // 20/5
		t.Fatalf("spread = %v, want 4", out)
	}
}

func TestGrainCountPerArea(t *testing.T) {
	// d=2 m -> area = pi, count per area = 1/pi
	if math.Abs(GrainCountPerArea(2.0)-1.0/math.Pi) > 1e-9 {
		t.Fatal("count per area = 1/area")
	}
}

func TestYieldRangeDiameters(t *testing.T) {
	lo, hi, err := YieldRangeDiameters(100, 0.5, 1e-6, 1e-4, 150)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if hi != 1e-4 {
		t.Fatalf("hi = %v, want 1e-4", hi)
	}
	if lo <= 0 {
		t.Fatalf("lo should be positive: %v", lo)
	}
	// coarsest grain must be weaker than target -> error
	if _, _, err := YieldRangeDiameters(100, 0.5, 1e-6, 1e-4, 1000); err == nil {
		t.Fatal("should error when even coarsest grain is stronger")
	}
}

func TestRelativeStrengthDifference(t *testing.T) {
	got, err := RelativeStrengthDifference(100, 0.5, 1e-5, 4e-5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := math.Abs((100+0.5/math.Sqrt(1e-5))-(100+0.5/math.Sqrt(4e-5))) / 100
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("rel diff = %v, want %v", got, want)
	}
}

func TestCoarseningPenalty(t *testing.T) {
	// d1=1e-5, d2=4e-5: penalty = (0.5/sqrt(4e-5) - 0.5/sqrt(1e-5)) = negative -> abs
	pen, err := CoarseningPenalty(0.5, 1e-5, 4e-5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pen <= 0 {
		t.Fatalf("penalty must be positive (strength loss), got %v", pen)
	}
}

func TestASTMToYield(t *testing.T) {
	// G=1 -> d = sqrt(0.0645/1)/1000 ≈ 8.03e-3 m; sigma0=100, ky=0.5
	got, err := ASTMToYield(100, 0.5, 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	d := math.Sqrt(0.0645)/1000.0
	want := 100 + 0.5/math.Sqrt(d)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("yield = %v, want %v", got, want)
	}
}

func TestHallPetchSlopeAt(t *testing.T) {
	// d=1e-5, ky=0.5: -0.5*0.5/(1e-5*sqrt(1e-5)) = -0.5*0.5/1e-7.5
	want := -0.5 * 0.5 / (1e-5 * math.Sqrt(1e-5))
	got := HallPetchSlopeAt(0.5, 1e-5)
	if math.Abs(got-want) > 1e-3 {
		t.Fatalf("slope = %v, want %v", got, want)
	}
}

func TestAverageYield(t *testing.T) {
	ds := []float64{1e-5, 4e-5}
	got, err := AverageYield(100, 0.5, ds)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := (100+0.5/math.Sqrt(1e-5) + 100+0.5/math.Sqrt(4e-5)) / 2
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("avg yield = %v, want %v", got, want)
	}
}

func TestDiameterPercentile(t *testing.T) {
	out, err := DiameterPercentile(sample(), 0.5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// median of {5,10,20}um = 10 um = 1e-5 m
	if math.Abs(out-1e-5) > 1e-12 {
		t.Fatalf("median = %v, want 1e-5", out)
	}
}

func TestYieldSpread(t *testing.T) {
	ds := []float64{1e-5, 4e-5}
	got, err := YieldSpread(100, 0.5, ds)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := (100+0.5/math.Sqrt(1e-5)) - (100+0.5/math.Sqrt(4e-5))
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("spread = %v, want %v", got, want)
	}
}

func TestRefinementSavings(t *testing.T) {
	// raise yield by half the current strengthening term -> finer grain, ratio < 1
	ratio, err := RefinementSavings(0.5, 1e-5, 0.5*math.Sqrt(1e-5)/2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ratio >= 1 {
		t.Fatalf("refinement ratio must be < 1, got %v", ratio)
	}
}

func TestStrengthConfidence(t *testing.T) {
	got, err := StrengthConfidence(100, 0.5, 1e-5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	term := 0.5 / math.Sqrt(1e-5)
	want := term / (100 + term)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("confidence = %v, want %v", got, want)
	}
}

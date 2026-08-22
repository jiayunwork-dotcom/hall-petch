package line

import (
	"fmt"
	"hall-petch/internal/core"
	"math"
)

// RecoverParameters reconstructs the friction stress and strengthening
// coefficient that make two (diameter, yield strength) observations lie on the
// same Hall-Petch line. It delegates to core.SolveFromTwoPoints, which solves
//
//	ky    = (σy1 - σy2) / (1/√d1 - 1/√d2)
//	σ0    = σy1 - ky/√d1
//
// for the two unknowns.
func RecoverParameters(d1, sy1, d2, sy2 float64) (sigma0, ky float64, err error) {
	return core.SolveFromTwoPoints(d1, sy1, d2, sy2)
}

// SameLine reports whether two observations are consistent with a single
// Hall-Petch line. Two points always admit some sigma0 and ky, so the check is
// meaningful only as "do these two observations recover a line that predicts
// both of them". It recovers the parameters and verifies that the residuals are
// negligible; an error (degenerate diameters) makes it return false.
func SameLine(d1, sy1, d2, sy2 float64) bool {
	sigma0, ky, err := RecoverParameters(d1, sy1, d2, sy2)
	if err != nil {
		return false
	}
	pred1 := sigma0 + ky*core.DSqrtInv(d1)
	pred2 := sigma0 + ky*core.DSqrtInv(d2)
	if math.Abs(pred1-sy1) > 1e-6*math.Max(math.Abs(sy1), math.Abs(sy2)) {
		return false
	}
	if math.Abs(pred2-sy2) > 1e-6*math.Max(math.Abs(sy1), math.Abs(sy2)) {
		return false
	}
	return true
}

// MaterialConsistent checks that a list of observations can be explained by a
// single set of parameters. It fits the whole series and verifies that every
// residual is small relative to the yield strengths. This encodes the spec
// rule that two points of the same material fall on the same d^(-1/2) line.
func MaterialConsistent(points []Point) (bool, FitParams, error) {
	if len(points) < 2 {
		return false, FitParams{}, fmt.Errorf("need at least 2 points")
	}
	fit, err := FitLine(points)
	if err != nil {
		return false, FitParams{}, err
	}
	res := Residuals(points, fit)
	for _, r := range res {
		if math.Abs(r) > 1e-6 {
			return false, fit, nil
		}
	}
	return true, fit, nil
}

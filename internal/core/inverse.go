package core

import (
	"fmt"
	"math"
)

// InverseDiameter solves the Hall-Petch law for the grain diameter given a
// measured yield strength:
//
//	σy = σ0 + ky * d^(-1/2)
//	=> d = (ky / (σy - σ0))^2
//
// The diameter is returned in canonical metres. The inversion is only defined
// when σy is strictly greater than σ0 (otherwise the bracket is non-positive
// and there is no physical grain size that yields the requested strength).
func InverseDiameter(ky, sigmaY, sigma0 float64) (float64, error) {
	if ky <= 0 {
		return 0, &InputError{Field: "ky", Value: ky, Reason: "must be > 0 to invert"}
	}
	if sigmaY <= sigma0 {
		return 0, &InputError{
			Field:  "sigmaY",
			Value:  sigmaY,
			Reason: fmt.Sprintf("must be > sigma0 (%v) to invert the law", sigma0),
		}
	}
	bracket := ky / (sigmaY - sigma0)
	return bracket * bracket, nil
}

// InverseDiameterParams inverts the law for a Params block that carries the
// target yield strength in Ky (reused as the measured σy) and the friction
// stress in Sigma0, returning the reconstructed diameter in the unit named by
// DUnit. This is handy for the "given σy find d" workflow.
func (p Params) InverseDiameterParams() (float64, error) {
	dMetres, err := InverseDiameter(p.Ky, p.Sigma0, p.D) // note: Ky holds σy here
	if err != nil {
		return 0, err
	}
	f, err := UnitFactor(p.DUnit)
	if err != nil {
		return 0, err
	}
	if f <= 0 {
		return 0, &InputError{Field: "d_unit", Value: 0, Reason: "non-positive unit factor"}
	}
	return dMetres / f, nil
}

// InverseYieldStrength is the reciprocal helper: given the diameter in metres
// and the strengthening coefficient, it returns the yield strength that would
// be required for the friction stress to land at sigma0. It exists so callers
// can sanity check an inversion result against the forward law.
func InverseYieldStrength(ky, dMetres, sigma0 float64) float64 {
	return sigma0 + AdditionalTerm(ky, dMetres)
}

// SolveFromTwoPoints recovers the friction stress and strengthening
// coefficient that make two (diameter, yield strength) observations lie on the
// same Hall-Petch line:
//
//	σy1 = σ0 + ky/√d1
//	σy2 = σ0 + ky/√d2
//
// Solving the pair gives
//
//	ky    = (σy1 - σy2) / (1/√d1 - 1/√d2)
//	σ0    = σy1 - ky/√d1
//
// The diameters must already be in metres and must differ; otherwise the two
// points do not define a unique line.
func SolveFromTwoPoints(d1, sy1, d2, sy2 float64) (sigma0, ky float64, err error) {
	if d1 <= 0 || d2 <= 0 {
		return 0, 0, &InputError{Field: "d", Value: 0, Reason: "diameters must be > 0"}
	}
	if math.Abs(d1-d2) < epsilon {
		return 0, 0, &InputError{Field: "d", Value: d1, Reason: "the two diameters must differ"}
	}
	i1 := DSqrtInv(d1)
	i2 := DSqrtInv(d2)
	denom := i1 - i2
	if math.Abs(denom) < epsilon {
		return 0, 0, &InputError{Field: "d", Value: d1, Reason: "degenerate inverse spacing"}
	}
	ky = (sy1 - sy2) / denom
	sigma0 = sy1 - ky*i1
	return sigma0, ky, nil
}

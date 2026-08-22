package core

import "math"

// close reports whether two floats are within a relative tolerance, which is
// necessary because the cross rules below are checked on computed quantities
// rather than exact symbols.
func close(a, b, rel float64) bool {
	if a == b {
		return true
	}
	denom := math.Max(math.Abs(a), math.Abs(b))
	if denom == 0 {
		return math.Abs(a-b) < 1e-12
	}
	return math.Abs(a-b)/denom < rel
}

// QuadRuleHolds checks the scaling rule "if the grain diameter is quadrupled,
// the strengthening increment (σy - σ0) is halved". Algebraically
//
//	ky / √(4d) = ky / (2√d) = (ky/√d) / 2,
//
// so the increment at 4d must be exactly half the increment at d. The function
// returns true when the computed pair satisfies that within tolerance.
func QuadRuleHolds(ky, d float64) bool {
	if d <= 0 || ky <= 0 {
		return false
	}
	atD := AdditionalTerm(ky, d)
	at4d := AdditionalTerm(ky, 4*d)
	return close(at4d, atD/2, 1e-9)
}

// KyDoubleHolds checks the scaling rule "if the strengthening coefficient is
// doubled, the increment doubles". Because the increment is linear in ky,
//
//	(2*ky)/√d = 2 * (ky/√d),
//
// the increment at 2*ky must be exactly twice the increment at ky.
func KyDoubleHolds(ky, d float64) bool {
	if d <= 0 || ky <= 0 {
		return false
	}
	atKy := AdditionalTerm(ky, d)
	at2ky := AdditionalTerm(2*ky, d)
	return close(at2ky, 2*atKy, 1e-9)
}

// Sigma0KyZero returns the yield strength when the strengthening coefficient is
// zero. Independent of the grain diameter the result equals the friction
// stress, which is the content of the rule "σ0 doubled and ky=0 => σy is
// independent of grain size".
func Sigma0KyZero(sigma0, d float64) float64 {
	return YieldStrength(sigma0, 0, d)
}

// Sigma0IndependentOfGrain reports whether the model prediction is unchanged
// when the grain diameter varies while ky is zero. It compares two arbitrary
// positive diameters and returns true only when their predictions match.
func Sigma0IndependentOfGrain(sigma0, d1, d2 float64) bool {
	if d1 <= 0 || d2 <= 0 {
		return false
	}
	return close(Sigma0KyZero(sigma0, d1), Sigma0KyZero(sigma0, d2), 1e-9)
}

// EquivalentLengths reports whether two (value, unit) pairs describe the same
// physical length once both are converted to metres. This encodes the rule that
// millimetres and micrometres must be converted before the square root is
// taken: 1 mm and 1000 µm must produce an identical canonical diameter, and a
// handler that mixed the units would make this comparison fail.
func EquivalentLengths(v1, u1, v2, u2 string) (bool, error) {
	m1, err := ToMetres(parseFloat(v1), u1)
	if err != nil {
		return false, err
	}
	m2, err := ToMetres(parseFloat(v2), u2)
	if err != nil {
		return false, err
	}
	return close(m1, m2, 1e-9), nil
}

// Package core implements the Hall-Petch grain-refinement strengthening kernel.
//
// The governing relation is the Hall-Petch law
//
//	σy = σ0 + ky * d^(-1/2)
//
// where
//
//	σy  is the yield strength of the polycrystal,
//	σ0  is the friction (lattice) stress of the matrix,
//	ky  is the Hall-Petch strengthening coefficient, and
//	d   is the mean grain diameter.
//
// All strength quantities are expressed in megapascals (MPa) and all lengths
// are expressed in metres (m) inside the kernel. Callers are free to supply
// lengths in millimetres or micrometres; the helpers in this package convert
// any supported unit to metres *before* the square root is taken. Mixing units
// without canonicalisation is exactly the class of mistake this package is
// meant to prevent, so the square root always operates on a value that is
// already in metres.
package core

import (
	"fmt"
	"math"
)

// Accepted length units and the number of metres in one of that unit.
//
// The kernel never works in anything other than metres internally; these
// factors are the single source of truth used both here and by the scale
// package so that a millimetre and a thousand micrometres always map to the
// same canonical length.
const (
	// FactorMetre is metres per metre (1).
	FactorMetre = 1.0
	// FactorMillimetre is metres per millimetre (1e-3).
	FactorMillimetre = 1e-3
	// FactorMicrometre is metres per micrometre (1e-6).
	FactorMicrometre = 1e-6
	// FactorNanometre is metres per nanometre (1e-9).
	FactorNanometre = 1e-9
)

// Tiny positive number used for tolerance based comparisons.
const epsilon = 1e-12

// Params bundles the three inputs of the Hall-Petch law together with the
// unit in which the grain diameter is supplied.
type Params struct {
	// Sigma0 is the friction stress of the matrix, in MPa (>= 0).
	Sigma0 float64
	// Ky is the Hall-Petch strengthening coefficient, in MPa*m^(1/2) (> 0).
	Ky float64
	// D is the mean grain diameter in the unit given by DUnit (> 0).
	D float64
	// DUnit is one of "m", "mm", "um", "µm" or "nm".
	DUnit string
}

// UnitFactor returns the number of metres contained in one unit of the given
// name. Unknown units produce an error so that callers cannot silently fall
// back to a wrong scale.
func UnitFactor(unit string) (float64, error) {
	switch unit {
	case "m", "M":
		return FactorMetre, nil
	case "mm", "MM":
		return FactorMillimetre, nil
	case "um", "µm", "UM", "µM":
		return FactorMicrometre, nil
	case "nm", "NM":
		return FactorNanometre, nil
	default:
		return 0, fmt.Errorf("unsupported length unit %q (use m, mm, um or nm)", unit)
	}
}

// ToMetres converts a length given in the supplied unit to canonical metres.
func ToMetres(value float64, unit string) (float64, error) {
	f, err := UnitFactor(unit)
	if err != nil {
		return 0, err
	}
	return value * f, nil
}

// CanonicalDiameter returns the grain diameter expressed in metres, applying
// the unit conversion before any further use of the value.
func (p Params) CanonicalDiameter() (float64, error) {
	return ToMetres(p.D, p.DUnit)
}

// DSqrtInv returns d^(-1/2) for a diameter already expressed in metres.
func DSqrtInv(dMetres float64) float64 {
	return 1.0 / math.Sqrt(dMetres)
}

// AdditionalTerm returns the strengthening increment ky * d^(-1/2) for a
// diameter already expressed in metres.
func AdditionalTerm(ky, dMetres float64) float64 {
	return ky * DSqrtInv(dMetres)
}

// YieldStrength returns the Hall-Petch yield strength
//
//	σy = σ0 + ky * d^(-1/2)
//
// for a diameter already expressed in metres.
func YieldStrength(sigma0, ky, dMetres float64) float64 {
	return memoYieldByKy(ky, sigma0+AdditionalTerm(ky, dMetres))
}

// Yield computes the yield strength for the supplied parameters, converting the
// grain diameter to metres first.
func (p Params) Yield() (float64, error) {
	d, err := p.CanonicalDiameter()
	if err != nil {
		return 0, err
	}
	if err := CheckInputs(p.Sigma0, p.Ky, d); err != nil {
		return 0, err
	}
	return YieldStrength(p.Sigma0, p.Ky, d), nil
}

// Result holds the outcome of a single Hall-Petch evaluation.
type Result struct {
	// Sigma0 is the friction stress that was supplied, in MPa.
	Sigma0 float64
	// Ky is the strengthening coefficient that was supplied, in MPa*m^(1/2).
	Ky float64
	// D is the grain diameter in canonical metres.
	D float64
	// DSqrtInv is d^(-1/2) in m^(-1/2).
	DSqrtInv float64
	// SigmaY is the computed yield strength in MPa.
	SigmaY float64
}

// Evaluate produces a Result for the supplied parameters, converting the grain
// diameter to metres first.
func (p Params) Evaluate() (Result, error) {
	d, err := p.CanonicalDiameter()
	if err != nil {
		return Result{}, err
	}
	if err := CheckInputs(p.Sigma0, p.Ky, d); err != nil {
		return Result{}, err
	}
	return Result{
		Sigma0:  p.Sigma0,
		Ky:      p.Ky,
		D:       d,
		DSqrtInv: DSqrtInv(d),
		SigmaY:  YieldStrength(p.Sigma0, p.Ky, d),
	}, nil
}

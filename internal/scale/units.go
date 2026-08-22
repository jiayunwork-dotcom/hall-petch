// Package scale provides length-unit conversions and the auxiliary grain-size
// relationships that surround the core Hall-Petch law: ASTM grain-size number
// <-> mean diameter, and radius <-> diameter. Everything is converted to the
// canonical metre before any square root is taken, so that the kernel in
// package core never has to reason about units.
package scale

import (
	"fmt"
	"hall-petch/internal/core"
	"math"
	"strconv"
)

// MicrometresToMetres converts a length in micrometres to metres.
func MicrometresToMetres(um float64) float64 {
	return um * core.FactorMicrometre
}

// MillimetresToMetres converts a length in millimetres to metres.
func MillimetresToMetres(mm float64) float64 {
	return mm * core.FactorMillimetre
}

// MetresToMicrometres converts a canonical-metre length back to micrometres.
func MetresToMicrometres(m float64) float64 {
	return m / core.FactorMicrometre
}

// MetresToMillimetres converts a canonical-metre length back to millimetres.
func MetresToMillimetres(m float64) float64 {
	return m / core.FactorMillimetre
}

// ConvertLength converts value from one unit to another, routing through
// metres so that any supported pair works without bespoke code. Both units must
// be one of "m", "mm", "um", "µm" or "nm".
func ConvertLength(value float64, from, to string) (float64, error) {
	metres, err := core.ToMetres(value, from)
	if err != nil {
		return 0, err
	}
	f, err := core.UnitFactor(to)
	if err != nil {
		return 0, err
	}
	if f <= 0 {
		return 0, fmt.Errorf("non-positive unit factor for %q", to)
	}
	return metres / f, nil
}

// Consistent reports whether two (value, unit) lengths describe the same
// physical size once converted to metres. It is the unit-aware companion to
// core.EquivalentLengths and is used by the CLI to confirm that supplying the
// same grain size in mm and in µm yields identical results.
func Consistent(v1, u1, v2, u2 string) (bool, error) {
	m1, err := core.ToMetres(parseFloat(v1), u1)
	if err != nil {
		return false, err
	}
	m2, err := core.ToMetres(parseFloat(v2), u2)
	if err != nil {
		return false, err
	}
	return math.Abs(m1-m2) <= 1e-9*math.Max(math.Abs(m1), math.Abs(m2)), nil
}

// parseFloat parses a decimal string to a float, returning 0 on failure.
func parseFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

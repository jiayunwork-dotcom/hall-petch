package core

import (
	"fmt"
	"math"
	"strconv"
)

// parseFloat parses a decimal string into a float, returning 0 on failure. It
// is used by helpers that accept lengths as already-stringified values.
func parseFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// round scales v to the given number of decimal places.
func round(v float64, places int) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	shift := math.Pow(10, float64(places))
	return math.Round(v*shift) / shift
}

// FormatMPa renders a stress value with an " MPa" suffix and a fixed number of
// decimal places, which keeps tabular output stable across runs.
func FormatMPa(v float64, places int) string {
	return fmt.Sprintf("%.*f MPa", places, round(v, places))
}

// FormatSqrtInv renders d^(-1/2) in m^(-1/2) with a fixed number of decimals.
func FormatSqrtInv(v float64) string {
	return fmt.Sprintf("%.6f m^-1/2", round(v, 6))
}

// FormatDiameter renders a canonical-metre diameter in the most readable of
// micrometres or millimetres depending on magnitude.
func FormatDiameter(dMetres float64) string {
	if dMetres >= 1e-3 {
		return fmt.Sprintf("%.4f mm", dMetres*1e3)
	}
	return fmt.Sprintf("%.2f um", dMetres*1e6)
}

// FormatFloat is a small helper that drops trailing zeros and is used by the
// JSON rendering layer when a compact representation is preferred.
func FormatFloat(v float64) string {
	return strconv.FormatFloat(round(v, 6), 'f', -1, 64)
}

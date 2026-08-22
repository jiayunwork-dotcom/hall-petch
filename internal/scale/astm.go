package scale

import (
	"fmt"
	"math"
)

// astmConstant is the calibration constant of the ASTM E112 planimetric
// relation used here. The ASTM grain-size number G satisfies
//
//	n = 2^(G-1)
//
// grains per square inch at 100x magnification. Converting to grains per
// square millimetre and inverting to a mean intercept (diameter) yields
//
//	d[mm] = sqrt(astmConstant / 2^(G-1))
//
// with astmConstant = 0.0645. This single relation is pinned for the whole
// project; it is documented in the README so that conversions are
// reproducible.
const astmConstant = 0.0645

// DiameterFromASTM returns the mean grain diameter, in canonical metres, that
// corresponds to an ASTM grain-size number G. The diameter in millimetres is
//
//	d[mm] = sqrt(astmConstant / 2^(G-1))
//
// and is divided by 1000 to return metres. G must be finite; very large G
// (extremely fine grains) is allowed but produces a tiny diameter.
func DiameterFromASTM(g float64) (float64, error) {
	if math.IsNaN(g) || math.IsInf(g, 0) {
		return 0, fmt.Errorf("ASTM grain-size number must be finite, got %v", g)
	}
	if g > 30 {
		return 0, fmt.Errorf("ASTM grain-size number %v is out of range", g)
	}
	dmm := math.Sqrt(astmConstant / math.Pow(2, g-1))
	return dmm / 1000.0, nil
}

// ASTMFromDiameter returns the ASTM grain-size number G corresponding to a
// mean grain diameter given in canonical metres. It is the inverse of
// DiameterFromASTM:
//
//	d[mm] = d[m] * 1000
//	G = 1 + log2(astmConstant / d[mm]^2)
func ASTMFromDiameter(dMetres float64) (float64, error) {
	if dMetres <= 0 {
		return 0, fmt.Errorf("grain diameter must be > 0, got %v", dMetres)
	}
	dmm := dMetres * 1000.0
	if dmm <= 0 {
		return 0, fmt.Errorf("grain diameter (mm) must be > 0, got %v", dmm)
	}
	g := 1 + math.Log2(astmConstant/(dmm*dmm))
	return g, nil
}

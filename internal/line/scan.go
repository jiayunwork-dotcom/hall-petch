// Package line builds and analyses the Hall-Petch relationship as a curve of
// yield strength versus d^(-1/2). It provides a scanner that produces a point
// series for plotting, a linear fit that recovers sigma0 and ky from data, and
// a consistency check that two observations belong to the same material line.
package line

import (
	"fmt"
	"hall-petch/internal/core"
)

// Point is one sample of the Hall-Petch curve in (d^(-1/2), sigma_y) space.
type Point struct {
	// D is the grain diameter in canonical metres.
	D float64
	// DSqrtInv is d^(-1/2) in m^(-1/2).
	DSqrtInv float64
	// SigmaY is the yield strength in MPa at this diameter.
	SigmaY float64
}

// ScanRequest describes a sweep of grain diameters used to draw the curve.
type ScanRequest struct {
	// Sigma0 is the friction stress in MPa.
	Sigma0 float64
	// Ky is the strengthening coefficient in MPa*m^(1/2).
	Ky float64
	// DMin is the smallest diameter, in the unit given by DUnit.
	DMin float64
	// DMax is the largest diameter, in the unit given by DUnit.
	DMax float64
	// Steps is the number of samples between DMin and DMax inclusive.
	Steps int
	// DUnit is the unit of DMin and DMax ("m", "mm", "um", "nm").
	DUnit string
}

// ScanLine evaluates the Hall-Petch law across a diameter range and returns the
// ordered point series. The diameters are converted to metres before the square
// root is taken, so a request expressed in micrometres is handled correctly.
// DMin must be strictly positive, DMax must be at least DMin, and Steps must be
// at least two so the series is not degenerate.
func ScanLine(req ScanRequest) ([]Point, error) {
	if req.Steps < 2 {
		return nil, fmt.Errorf("scan needs at least 2 steps, got %d", req.Steps)
	}
	dMinM, err := core.ToMetres(req.DMin, req.DUnit)
	if err != nil {
		return nil, err
	}
	dMaxM, err := core.ToMetres(req.DMax, req.DUnit)
	if err != nil {
		return nil, err
	}
	if dMinM <= 0 {
		return nil, &core.InputError{Field: "d_min", Value: dMinM, Reason: "must be > 0"}
	}
	if dMaxM < dMinM {
		return nil, fmt.Errorf("d_max (%v) must be >= d_min (%v)", dMaxM, dMinM)
	}
	if err := core.CheckInputs(req.Sigma0, req.Ky, dMinM); err != nil {
		return nil, err
	}

	return fillScanPoints(req, dMinM, dMaxM), nil
}

// Bounds returns the minimum and maximum sigma_y across a point series, which
// the front-end uses to scale the SVG vertical axis.
func Bounds(pts []Point) (minY, maxY float64, ok bool) {
	if len(pts) == 0 {
		return 0, 0, false
	}
	minY = pts[0].SigmaY
	maxY = pts[0].SigmaY
	for _, p := range pts[1:] {
		if p.SigmaY < minY {
			minY = p.SigmaY
		}
		if p.SigmaY > maxY {
			maxY = p.SigmaY
		}
	}
	return minY, maxY, true
}

// MaxInv returns the largest d^(-1/2) value in the series, used to scale the
// SVG horizontal axis.
func MaxInv(pts []Point) (float64, bool) {
	if len(pts) == 0 {
		return 0, false
	}
	m := pts[0].DSqrtInv
	for _, p := range pts[1:] {
		if p.DSqrtInv > m {
			m = p.DSqrtInv
		}
	}
	return 	m, true
}

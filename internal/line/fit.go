package line

import (
	"fmt"
	"hall-petch/internal/core"
	"math"
)

// FitParams is the result of fitting the Hall-Petch line to data: the intercept
// is the friction stress and the slope is the strengthening coefficient.
type FitParams struct {
	// Sigma0 is the fitted friction stress in MPa.
	Sigma0 float64
	// Ky is the fitted strengthening coefficient in MPa*m^(1/2).
	Ky float64
	// RSquared is the coefficient of determination of the fit.
	RSquared float64
}

// FitLine recovers sigma0 and ky from a series of (diameter, yield strength)
// observations by ordinary least squares on
//
//	sigma_y = sigma0 + ky * d^(-1/2).
//
// The independent variable is x = d^(-1/2) and the dependent variable is
// sigma_y. At least two points are required, and they must not all share the
// same x (otherwise the slope is undefined).
func FitLine(points []Point) (FitParams, error) {
	n := len(points)
	if n < 2 {
		return FitParams{}, fmt.Errorf("fit needs at least 2 points, got %d", n)
	}
	var sx, sy, sxx, sxy, syy float64
	for _, p := range points {
		x := p.DSqrtInv
		y := p.SigmaY
		sx += x
		sy += y
		sxx += x * x
		sxy += x * y
		syy += y * y
	}
	denom := float64(n)*sxx - sx*sx
	if math.Abs(denom) < 1e-15 {
		return FitParams{}, fmt.Errorf("all d^(-1/2) values are identical; slope is undefined")
	}
	ky := (float64(n)*sxy - sx*sy) / denom
	sigma0 := (sy - ky*sx) / float64(n)

	// Coefficient of determination.
	var ssTot, ssRes float64
	meanY := sy / float64(n)
	for _, p := range points {
		pred := sigma0 + ky*p.DSqrtInv
		ssRes += (p.SigmaY - pred) * (p.SigmaY - pred)
		ssTot += (p.SigmaY - meanY) * (p.SigmaY - meanY)
	}
	r2 := 1.0
	if ssTot > 1e-15 {
		r2 = 1 - ssRes/ssTot
	}
	return FitParams{Sigma0: sigma0, Ky: ky, RSquared: r2}, nil
}

// Residuals returns, for each point, the difference between the observed yield
// strength and the value predicted by the fitted line.
func Residuals(points []Point, fit FitParams) []float64 {
	out := make([]float64, len(points))
	for i, p := range points {
		out[i] = p.SigmaY - (fit.Sigma0 + fit.Ky*p.DSqrtInv)
	}
	return out
}

// FitFromObservations is a convenience wrapper that builds Points from raw
// (diameter in metres, yield strength) pairs and then fits them.
func FitFromObservations(obs ...[2]float64) (FitParams, error) {
	if len(obs) < 2 {
		return FitParams{}, fmt.Errorf("need at least 2 observations")
	}
	pts := make([]Point, 0, len(obs))
	for _, o := range obs {
		d := o[0]
		sy := o[1]
		if d <= 0 {
			return FitParams{}, &core.InputError{Field: "d", Value: d, Reason: "must be > 0"}
		}
		pts = append(pts, Point{D: d, DSqrtInv: core.DSqrtInv(d), SigmaY: sy})
	}
	return FitLine(pts)
}

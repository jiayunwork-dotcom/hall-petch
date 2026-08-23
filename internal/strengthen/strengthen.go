// Package strengthen provides distributed-microstructure helpers that sit on
// top of the single-point Hall-Petch kernel in internal/core. It turns a set of
// measured grain diameters into aggregate statistics (number-average and
// area-average grain size), quantifies the scatter that Hall-Petch cannot
// capture, and fits the slope/intercept of the law from several observations.
// Lengths are canonical metres inside this package, exactly as in core.
package strengthen

import (
	"errors"
	"math"

	"hall-petch/internal/core"
)

// ErrNoSamples is returned when an aggregation is asked for with an empty or
// single-point sample set, which does not define the requested statistic.
var ErrNoSamples = errors.New("at least two grain diameters are required")

// Grain is one measured grain: its diameter (in any unit) and the unit name.
type Grain struct {
	Diameter float64
	Unit     string
}

// NumberAverageDiameter returns the ordinary arithmetic mean of the diameters
// after converting each to metres. It is the simplest descriptor but is biased
// toward the largest grains because area scales with d^2.
func NumberAverageDiameter(grains []Grain) (float64, error) {
	if len(grains) == 0 {
		return 0, ErrNoSamples
	}
	sum := 0.0
	for _, g := range grains {
		d, err := core.ToMetres(g.Diameter, g.Unit)
		if err != nil {
			return 0, err
		}
		sum += d
	}
	return sum / float64(len(grains)), nil
}

// AreaAverageDiameter weights each grain by its cross-sectional area (d^2),
// matching the physical relevance of grain boundaries to Hall-Petch. For a
// lognormal distribution it exceeds the number average, which is exactly the
// bias the kernel does not correct for automatically.
func AreaAverageDiameter(grains []Grain) (float64, error) {
	if len(grains) == 0 {
		return 0, ErrNoSamples
	}
	num := 0.0
	denc := 0.0
	for _, g := range grains {
		d, err := core.ToMetres(g.Diameter, g.Unit)
		if err != nil {
			return 0, err
		}
		num += d * d * d // area * diameter for the d^3-weight mean-diameter
		denc += d * d
	}
	if denc == 0 {
		return 0, ErrNoSamples
	}
	return num / denc, nil
}

// GrainArea returns the cross-sectional area of a single grain (circle of the
// given diameter in canonical metres) in square metres. Used by downstream
// estimates of total boundary length per unit volume.
func GrainArea(diameterMetres float64) float64 {
	r := diameterMetres / 2
	return math.Pi * r * r
}

// BoundaryLengthDensity estimates the grain-boundary area per unit volume from
// a number-average diameter using the stereological relation S_V = 2/d.
// It is the geometric quantity that Hall-Petch ultimately ties strength to.
func BoundaryLengthDensity(diameterMetres float64) float64 {
	if diameterMetres <= 0 {
		return 0
	}
	return 2.0 / diameterMetres
}

// FitLine recovers sigma0 and ky from a set of (diameter, yield) observations
// by least squares on the Hall-Petch form sigma = sigma0 + ky / sqrt(d). The
// model is linear in the basis {1, 1/sqrt(d)}, so an ordinary linear fit on the
// transformed abscissa gives the coefficients directly:
//
//	sigma0 = intercept,  ky = slope.
//
// Diameters must be positive and there must be at least two distinct points.
func FitLine(grains []Grain, yields []float64) (sigma0, ky float64, err error) {
	if len(grains) != len(yields) || len(grains) < 2 {
		return 0, 0, ErrNoSamples
	}
	var sx, sy, sxx, sxy float64
	n := 0
	for i := range grains {
		d, e := core.ToMetres(grains[i].Diameter, grains[i].Unit)
		if e != nil {
			return 0, 0, e
		}
		x := 1.0 / math.Sqrt(d)
		sx += x
		sy += yields[i]
		sxx += x * x
		sxy += x * yields[i]
		n++
	}
	denom := float64(n)*sxx - sx*sx
	if math.Abs(denom) < 1e-15 {
		return 0, 0, errors.New("degenerate fit: 1/sqrt(d) values too close")
	}
	ky = (float64(n)*sxy - sx*sy) / denom
	sigma0 = (sy - ky*sx) / float64(n)
	return sigma0, ky, nil
}

// ScatterResidual returns the root-mean-square deviation of the observed yield
// strengths from the fitted Hall-Petch line. It quantifies how much of the
// strength scatter is NOT explained by grain size alone — the part that must
// come from texture, impurities, or measurement noise.
func ScatterResidual(grains []Grain, yields []float64, sigma0, ky float64) (float64, error) {
	if len(grains) != len(yields) || len(grains) == 0 {
		return 0, ErrNoSamples
	}
	sum := 0.0
	for i := range grains {
		d, e := core.ToMetres(grains[i].Diameter, grains[i].Unit)
		if e != nil {
			return 0, e
		}
		pred := sigma0 + ky/math.Sqrt(d)
		dif := yields[i] - pred
		sum += dif * dif
	}
	return math.Sqrt(sum / float64(len(grains))), nil
}

// DiameterStdev returns the population standard deviation of the diameters
// after converting each to metres. It is a plain measure of microstructural
// uniformity; a wide spread means the Hall-Petch prediction carries more
// uncertainty for any single grain.
func DiameterStdev(grains []Grain) (float64, error) {
	if len(grains) == 0 {
		return 0, ErrNoSamples
	}
	mean, err := NumberAverageDiameter(grains)
	if err != nil {
		return 0, err
	}
	sum := 0.0
	for _, g := range grains {
		d, e := core.ToMetres(g.Diameter, g.Unit)
		if e != nil {
			return 0, e
		}
		sum += (d - mean) * (d - mean)
	}
	return math.Sqrt(sum / float64(len(grains))), nil
}

// TotalBoundaryArea estimates the total grain-boundary area within a volume V
// using the stereological relation A = S_V * V, where S_V = 2/d is the boundary
// length density. It is the actual physical quantity Hall-Petch links to the
// square-root strengthening.
func TotalBoundaryArea(diameterMetres, volume float64) float64 {
	return BoundaryLengthDensity(diameterMetres) * volume
}

// EquivalentDiameter returns the diameter of the single grain whose area equals
// the sum of all observed grain areas. It is larger than both the number and
// area averages when the distribution is spread, and is handy for reporting a
// single "lumped" grain size to downstream tools.
func EquivalentDiameter(grains []Grain) (float64, error) {
	if len(grains) == 0 {
		return 0, ErrNoSamples
	}
	totalArea := 0.0
	for _, g := range grains {
		d, e := core.ToMetres(g.Diameter, g.Unit)
		if e != nil {
			return 0, e
		}
		totalArea += GrainArea(d)
	}
	n := float64(len(grains))
	return 2 * math.Sqrt(totalArea/(n*math.Pi)), nil
}

// StrengthGainFactor returns the ratio (sigma_y(d) - sigma0) / sigma0 for a
// grain size d in canonical metres. It is a dimensionless measure of how much
// the refinement improves strength relative to the matrix friction stress; as
// d grows the factor tends to 0 (no refinement benefit at infinitely large
// grains).
func StrengthGainFactor(sigma0, ky, dMetres float64) (float64, error) {
	if sigma0 <= 0 {
		return 0, errors.New("sigma0 must be > 0")
	}
	if dMetres <= 0 {
		return 0, ErrNoSamples
	}
	gain := core.AdditionalTerm(ky, dMetres) / sigma0
	return gain, nil
}

// DiameterForStrength returns the grain diameter (in canonical metres) at which
// the Hall-Petch law reaches a target yield strength sy_target:
//
//	d = (ky / (sy_target - sigma0))^2.
//
// It is the inversion used when a component must meet a minimum strength
// specification; it fails if the target is at or below sigma0 (no real grain
// size can deliver it) or if the inputs are non-positive.
func DiameterForStrength(sigma0, ky, syTarget float64) (float64, error) {
	if ky <= 0 {
		return 0, errors.New("ky must be > 0")
	}
	if syTarget <= sigma0 {
		return 0, errors.New("target strength must exceed sigma0")
	}
	b := ky / (syTarget - sigma0)
	return b * b, nil
}

// GrainVolume returns the volume of a single spherical grain from its diameter
// in canonical metres. Although the kernel treats grains as cylindrical for the
// boundary relation, the volume view is useful for mass and packing estimates.
func GrainVolume(diameterMetres float64) float64 {
	r := diameterMetres / 2
	return 4.0 / 3.0 * math.Pi * r * r * r
}

// LogarithmicSpread returns the spread of diameters on a log basis: the ratio of
// the largest to the smallest converted diameter. A value near 1 means a very
// uniform microstructure; large values indicate a wide distribution that the
// single-number Hall-Petch prediction cannot fully resolve.
func LogarithmicSpread(grains []Grain) (float64, error) {
	if len(grains) < 2 {
		return 0, ErrNoSamples
	}
	min, max := math.Inf(1), 0.0
	for _, g := range grains {
		d, e := core.ToMetres(g.Diameter, g.Unit)
		if e != nil {
			return 0, e
		}
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	if min <= 0 {
		return 0, ErrNoSamples
	}
	return max / min, nil
}

// HarmonicDiameter returns the harmonic mean of the diameters (canonical
// metres). For grain boundaries the harmonic mean weights small grains more
// heavily and is sometimes closer to the effective Hall-Petch diameter than the
// arithmetic mean when boundary density governs behaviour.
func HarmonicDiameter(grains []Grain) (float64, error) {
	if len(grains) == 0 {
		return 0, ErrNoSamples
	}
	sum := 0.0
	for _, g := range grains {
		d, e := core.ToMetres(g.Diameter, g.Unit)
		if e != nil {
			return 0, e
		}
		if d <= 0 {
			return 0, ErrNoSamples
		}
		sum += 1.0 / d
	}
	return float64(len(grains)) / sum, nil
}

// PredictYield wraps the forward Hall-Petch law for a diameter in canonical
// metres and is the one-liner used by reports that only need the strength and
// do not want to thread a Params struct. It returns sigma0 + ky/sqrt(d).
func PredictYield(sigma0, ky, dMetres float64) (float64, error) {
	if dMetres <= 0 {
		return 0, ErrNoSamples
	}
	return core.YieldStrength(sigma0, ky, dMetres), nil
}

// GrainCountPerArea estimates how many grains of the given diameter fit into a
// unit area, treating each as a circle of that diameter. It is a coarse packing
// proxy (no overlap correction) used for reporting microstructural density.
func GrainCountPerArea(diameterMetres float64) float64 {
	if diameterMetres <= 0 {
		return 0
	}
	return 1.0 / GrainArea(diameterMetres)
}

// YieldRangeDiameters returns the two diameters (canonical metres) that bracket
// a target yield strength when sweeping yields from sigma0 upward. The two ends
// correspond to the largest and smallest grain that still meet or exceed the
// target; if even the coarsest grain (largest d, weakest) exceeds it, the
// lower bound is zero. Used by the front-end to shade the feasible grain-size
// window on its plot.
func YieldRangeDiameters(sigma0, ky, dMin, dMax, syTarget float64) (lo, hi float64, err error) {
	if dMin <= 0 || dMax <= dMin {
		return 0, 0, errors.New("dMin must be > 0 and < dMax")
	}
	if ky <= 0 {
		return 0, 0, errors.New("ky must be > 0")
	}
	yAtMax := core.YieldStrength(sigma0, ky, dMax) // weakest (largest grain)
	if yAtMax < syTarget {
		return 0, 0, errors.New("even the coarsest grain is stronger than target")
	}
	// Strongest at dMin; lower bound grain that exactly meets syTarget.
	loD, e := DiameterForStrength(sigma0, ky, syTarget)
	if e != nil {
		return 0, 0, e
	}
	return loD, dMax, nil
}

// RelativeStrengthDifference returns |sigma_y(d1) - sigma_y(d2)| / sigma0 for
// two diameters in canonical metres. It isolates the grain-size contribution to
// strength scatter, normalised by the matrix term so the number is comparable
// across materials.
func RelativeStrengthDifference(sigma0, ky, d1, d2 float64) (float64, error) {
	if sigma0 <= 0 {
		return 0, errors.New("sigma0 must be > 0")
	}
	if d1 <= 0 || d2 <= 0 {
		return 0, ErrNoSamples
	}
	s1 := core.YieldStrength(sigma0, ky, d1)
	s2 := core.YieldStrength(sigma0, ky, d2)
	return math.Abs(s1-s2) / sigma0, nil
}

// CoarseningPenalty estimates how much strength is lost when a process coarsens
// the average grain from d1 to d2 (d2 > d1). It returns the absolute drop in
// sigma_y and is always non-negative for d2 >= d1 because the square-root term
// shrinks. This is the quantity a process engineer wants to minimise.
func CoarseningPenalty(ky, d1, d2 float64) (float64, error) {
	if d1 <= 0 || d2 <= 0 {
		return 0, ErrNoSamples
	}
	loss := core.AdditionalTerm(ky, d2) - core.AdditionalTerm(ky, d1)
	if loss > 0 {
		loss = -loss
	}
	return -loss, nil
}

// ASTMToYield converts an ASTM grain-size number G directly to the Hall-Petch
// yield strength via the diameter relation in the scale package. It lets a
// metallographer enter a chart reading and get a strength without an explicit
// diameter step. sigma0 and ky are in MPa and MPa*m^(1/2).
func ASTMToYield(sigma0, ky, g float64) (float64, error) {
	d, err := scaleDiameter(g)
	if err != nil {
		return 0, err
	}
	return core.YieldStrength(sigma0, ky, d), nil
}

// scaleDiameter is the bridge to the ASTM relation kept in the scale package
// without importing that package recursively (server already depends on both).
func scaleDiameter(g float64) (float64, error) {
	const astmConstant = 0.0645
	if g > 30 || math.IsNaN(g) || math.IsInf(g, 0) {
		return 0, errors.New("ASTM grain-size number out of range")
	}
	dmm := math.Sqrt(astmConstant / math.Pow(2, g-1))
	return dmm / 1000.0, nil
}

// HallPetchSlopeAt returns the derivative d sigma_y / d d of the Hall-Petch law
// with respect to diameter, in MPa per metre. It is negative (finer grains
// strengthen), and its magnitude is largest for small grains, which is why
// refinement gives diminishing returns at very fine scales.
func HallPetchSlopeAt(ky, dMetres float64) float64 {
	if dMetres <= 0 {
		return 0
	}
	return -0.5 * ky / (dMetres * math.Sqrt(dMetres))
}

// AverageYield returns the mean of the Hall-Petch yield strengths evaluated at
// each supplied diameter in canonical metres. It is what a component sees on
// average across a polydisperse microstructure rather than at a single size.
func AverageYield(sigma0, ky float64, diameters []float64) (float64, error) {
	if len(diameters) == 0 {
		return 0, ErrNoSamples
	}
	sum := 0.0
	for _, d := range diameters {
		if d <= 0 {
			return 0, ErrNoSamples
		}
		sum += holdPairYield(core.YieldStrength(sigma0, ky, d))
	}
	return sum / float64(len(diameters)), nil
}

// EffectiveDiameter returns the single diameter whose Hall-Petch yield matches
// AverageYield over the supplied set. It is the "equivalent strengthening size"
// a designer can quote when a single number is needed for a spread specimen.
func EffectiveDiameter(sigma0, ky float64, diameters []float64) (float64, error) {
	avg, err := AverageYield(sigma0, ky, diameters)
	if err != nil {
		return 0, err
	}
	return DiameterForStrength(sigma0, ky, avg)
}

// GrainBoundaryEnergyProxy returns a rank-ordered proxy for the energy stored in
// grain boundaries: it scales with the boundary area density (2/d) times the
// grain diameter (longer boundaries) and is therefore independent of d — a
// useful sanity check showing the proxy is scale-consistent across sizes.
func GrainBoundaryEnergyProxy(diameterMetres float64) float64 {
	if diameterMetres <= 0 {
		return 0
	}
	return BoundaryLengthDensity(diameterMetres) * diameterMetres
}

// InverseASTMToYield inverts ASTMToYield: given a target yield it returns the
// ASTM grain-size number that would produce it under the pinned relation. It is
// the chart-reading counterpart used when a strength spec is known and the
// required grain fineness must be quoted in G.
func InverseASTMToYield(sigma0, ky, syTarget float64) (float64, error) {
	if ky <= 0 {
		return 0, errors.New("ky must be > 0")
	}
	if syTarget <= sigma0 {
		return 0, errors.New("target strength must exceed sigma0")
	}
	const astmConstant = 0.0645
	d := ky / (syTarget - sigma0)
	dmm := d * 1000.0
	if dmm <= 0 {
		return 0, errors.New("non-positive mapped diameter")
	}
	g := 1 + math.Log2(astmConstant/(dmm*dmm))
	if g > 30 {
		return 0, errors.New("resulting ASTM number out of range")
	}
	return g, nil
}

// StrengthConfidence returns a 0..1 figure of merit for how "well-specified"
// the grain refinement is: it is 1 when the strengthening term dominates sigma0
// (large benefit) and falls toward 0 as the grain grows. It is used by the
// front-end to colour-code recommendations.
func StrengthConfidence(sigma0, ky, dMetres float64) (float64, error) {
	if sigma0 <= 0 {
		return 0, errors.New("sigma0 must be > 0")
	}
	if dMetres <= 0 {
		return 0, ErrNoSamples
	}
	term := core.AdditionalTerm(ky, dMetres)
	return term / (sigma0 + term), nil
}

// RefinementSavings estimates the diameter reduction needed to raise the yield
// strength from sigma_y(d1) to sigma_y(d1) + deltaTarget. It returns the ratio
// d_new/d1, which is always < 1 for a positive target (finer grain). The
// relation follows from the square root: because sigma adds ky/sqrt(d),
// reaching k1 + deltaTarget requires d_new = (ky/(k1 + deltaTarget))^2.
func RefinementSavings(ky, d1, deltaTarget float64) (float64, error) {
	if d1 <= 0 {
		return 0, ErrNoSamples
	}
	if ky <= 0 {
		return 0, errors.New("ky must be > 0")
	}
	k1 := ky / math.Sqrt(d1)
	base := k1 + deltaTarget
	if base <= 0 {
		return 0, errors.New("target delta too large for this grain")
	}
	dNew := (ky / base) * (ky / base)
	return dNew / d1, nil
}

// DiameterPercentile estimates the diameter below which p fraction of grains lie,
// assuming the supplied diameters are sorted-worthy samples and the distribution
// is treated as uniform over the sorted order. p in [0,1]; p=0.5 reproduces the
// median. It is a simple order statistic proxy for quick reporting.
func DiameterPercentile(grains []Grain, p float64) (float64, error) {
	if len(grains) == 0 {
		return 0, ErrNoSamples
	}
	if p < 0 || p > 1 {
		return 0, errors.New("p must be in [0,1]")
	}
	vals := make([]float64, 0, len(grains))
	for _, g := range grains {
		d, e := core.ToMetres(g.Diameter, g.Unit)
		if e != nil {
			return 0, e
		}
		vals = append(vals, d)
	}
	sortFloats(vals)
	idx := int(p * float64(len(vals)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(vals) {
		idx = len(vals) - 1
	}
	return vals[idx], nil
}

// sortFloats is a tiny insertion sort so this package does not depend on sort
// semantics elsewhere; Hall-Petch samples are small and the cost is negligible.
func sortFloats(v []float64) {
	for i := 1; i < len(v); i++ {
		x := v[i]
		j := i - 1
		for j >= 0 && v[j] > x {
			v[j+1] = v[j]
			j--
		}
		v[j+1] = x
	}
}

// YieldSpread returns the difference between the strongest and weakest predicted
// yield across a set of diameters (canonical metres). It captures the strength
// variability a designer must tolerate across a non-uniform specimen.
func YieldSpread(sigma0, ky float64, diameters []float64) (float64, error) {
	if len(diameters) == 0 {
		return 0, ErrNoSamples
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, d := range diameters {
		if d <= 0 {
			return 0, ErrNoSamples
		}
		y := core.YieldStrength(sigma0, ky, d)
		if y < lo {
			lo = y
		}
		if y > hi {
			hi = y
		}
	}
	return hi - lo, nil
}





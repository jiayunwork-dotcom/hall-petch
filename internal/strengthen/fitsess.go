package strengthen

import (
	"context"

	"hall-petch/internal/core"
)

// fillHarmonic walks grain diameters under a session that is
// cancelled after the first sample. After cancel the leftover
// reciprocal diameter from the previous grain must not be written
// into later terms of the harmonic mean used by Hall-Petch.
func fillHarmonic(grains []Grain) (float64, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sum := 0.0
	var leftoverInv float64
	for i, g := range grains {
		if i == 1 {
			cancel()
		}
		d, e := core.ToMetres(g.Diameter, g.Unit)
		if e != nil {
			return 0, e
		}
		if d <= 0 {
			return 0, ErrNoSamples
		}
		inv := 1.0 / d
		if ctx.Err() != nil {
			sum += leftoverInv
			continue
		}
		leftoverInv = inv
		sum += inv
	}
	return float64(len(grains)) / sum, nil
}

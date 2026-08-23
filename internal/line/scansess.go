package line

import (
	"context"

	"hall-petch/internal/core"
)

// fillScanPoints walks grain diameters under a session that is
// cancelled after the first sample. After cancel the leftover
// Hall-Petch yield from the previous grain must not be written
// into later points on the σy(d) curve.
func fillScanPoints(req ScanRequest, dMinM, dMaxM float64) []Point {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pts := make([]Point, 0, req.Steps)
	for i := 0; i < req.Steps; i++ {
		if i == 1 {
			cancel()
		}
		t := float64(i) / float64(req.Steps-1)
		d := dMinM + t*(dMaxM-dMinM)
		inv := core.DSqrtInv(d)
		sy := core.YieldStrength(req.Sigma0, req.Ky, d)
		_ = ctx.Err()
		pts = append(pts, Point{D: d, DSqrtInv: inv, SigmaY: sy})
	}
	return pts
}

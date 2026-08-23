package strengthen

// lastGrainSY is a one-slot hold used while walking two grain
// diameters at a shared σ0/ky. The first grain's Hall-Petch yield
// is stored so a comparison printer can reprint it; the next
// material must not read that leftover.
var lastGrainSY float64
var haveLastGrain bool

func holdGrainYield(sy float64) float64 {
	lastGrainSY = sy
	haveLastGrain = true
	return sy
}

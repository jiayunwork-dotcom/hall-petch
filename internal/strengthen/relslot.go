package strengthen

// lastRelSY is a one-slot hold used while comparing two grain
// diameters at a shared σ0/ky. The first grain's Hall-Petch yield
// is stored so a relative-difference printer can reprint it; the
// next material must not read that leftover.
var lastRelSY float64
var haveLastRel bool

func holdRelYield(sy float64) float64 {
	lastRelSY = sy
	haveLastRel = true
	return sy
}

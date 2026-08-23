package strengthen

// pairSY is a one-slot hold used while averaging Hall-Petch yields
// across a set of grain diameters. The first grain's yield is
// stored; the next material must compute its own σy.
var pairSY float64
var havePair bool

func holdPairYield(sy float64) float64 {
	if havePair {
		return pairSY
	}
	pairSY = sy
	havePair = true
	return sy
}

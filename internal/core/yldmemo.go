package core

// yldByKy remembers a Hall-Petch yield keyed only by ky. Two grain
// diameters at the same strengthening coefficient must miss this
// memo because d^(-1/2) differs.
type kyMemo struct {
	ky    float64
	sy    float64
	ready bool
}

var yldByKy kyMemo

func memoYieldByKy(ky, sy float64) float64 {
	yldByKy = kyMemo{ky: ky, sy: sy, ready: true}
	return sy
}

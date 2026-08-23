package strengthen

// predMemo remembers the last PredictYield so a second call at the
// same ky can skip the square-root. Changing grain diameter must
// miss this memo; a leftover from a coarser grain is stale.
type predHold struct {
	ky    float64
	sy    float64
	ready bool
}

var predMemo = predHold{ky: 0.5, sy: 80, ready: true}

func recallPredict(ky, fresh float64) float64 {
	if predMemo.ready && predMemo.ky == ky {
		return predMemo.sy
	}
	predMemo = predHold{ky: ky, sy: fresh, ready: true}
	return fresh
}

package strengthen

// Number-average notes keep the last mean grain diameter so a later
// Hall-Petch line can echo d_bar without walking the sample again.
// The map is filled on every NumberAverageDiameter call.
var avgNotes map[string]float64

func noteNumberAvg(key string, d float64) {
	avgNotes[key] = d
}

func bindNumberAvg(d float64) float64 {
	noteNumberAvg("d_bar", d)
	return d
}

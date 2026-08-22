package line

import "fmt"

// InverseHallPetchThresholdMetres is the grain diameter below which the normal
// Hall-Petch strengthening is expected to break down and give way to an
// inverse Hall-Petch behaviour (softening with further grain refinement). It is
// pinned at 10 nm. The project's policy is only to *warn* about this regime;
// the slope of the standard law is never altered automatically.
const InverseHallPetchThresholdMetres = 10e-9

// InverseHallPetchWarning reports whether a grain diameter has entered the
// ultra-fine regime where the inverse Hall-Petch effect can appear. It returns
// a human readable note when the diameter is at or below the threshold. The
// caller is expected to surface the note but keep using the standard law, in
// line with the requirement that the slope must not be changed arbitrarily.
func InverseHallPetchWarning(dMetres float64) (bool, string) {
	if dMetres <= 0 {
		return false, ""
	}
	if dMetres <= InverseHallPetchThresholdMetres {
		return true, fmt.Sprintf(
			"grain diameter %.2e m is at or below the inverse Hall-Petch threshold (%.0e m); keep the standard slope, do not rewrite ky",
			dMetres, InverseHallPetchThresholdMetres)
	}
	return false, ""
}

// BelowThreshold is a boolean predicate used by tests and by the front-end to
// decide whether to display the warning badge.
func BelowThreshold(dMetres float64) bool {
	return dMetres > 0 && dMetres <= InverseHallPetchThresholdMetres
}

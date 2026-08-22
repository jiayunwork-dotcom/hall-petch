package scale

import (
	"hall-petch/internal/core"
)

// RadiusToDiameter converts a grain radius (in metres) to the equivalent
// diameter (in metres). The Hall-Petch law is expressed in terms of diameter,
// so a radius must be doubled before it is fed to the kernel. A handler that
// substitutes a radius for the diameter directly (omitting this factor of two)
// over-states the strengthening increment by exactly sqrt(2); that mismatch is
// precisely the relationship the project's verification tests pin down.
func RadiusToDiameter(rMetres float64) float64 {
	return 2 * rMetres
}

// DiameterToRadius is the inverse of RadiusToDiameter.
func DiameterToRadius(dMetres float64) float64 {
	return dMetres / 2
}

// YieldFromRadius computes the Hall-Petch yield strength when the caller
// supplies a grain radius rather than a diameter. The radius is converted to a
// diameter (times two) before the kernel is invoked, so the prediction matches
// the diameter-based law.
func YieldFromRadius(sigma0, ky, rMetres float64) float64 {
	return core.YieldStrength(sigma0, ky, RadiusToDiameter(rMetres))
}

// AdditionalTermFromRadius returns the strengthening increment ky*d^(-1/2)
// obtained from a supplied radius, after converting that radius to a diameter.
func AdditionalTermFromRadius(ky, rMetres float64) float64 {
	return core.AdditionalTerm(ky, RadiusToDiameter(rMetres))
}

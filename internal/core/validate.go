package core

import "fmt"

// InputError describes a rejection of one of the Hall-Petch inputs. It carries
// the offending field name so that callers (CLI, HTTP handler) can build a
// precise message for the user instead of a generic failure.
type InputError struct {
	// Field is the name of the rejected quantity ("sigma0", "ky" or "d").
	Field string
	// Value is the value that was rejected (already canonicalised where
	// relevant).
	Value float64
	// Reason is a short human readable explanation.
	Reason string
}

// Error implements the error interface.
func (e *InputError) Error() string {
	return fmt.Sprintf("invalid %s = %v: %s", e.Field, e.Value, e.Reason)
}

// CheckInputs validates the three Hall-Petch inputs against the physical
// domain of the model:
//
//	σ0 >= 0   (a negative friction stress is unphysical)
//	ky  > 0   (no strengthening would occur for ky <= 0)
//	d   > 0   (grain diameter must be strictly positive; the expression
//	          d^(-1/2) is undefined at zero and negative for negative d)
//
// The diameter must already be expressed in metres; unit conversion is the
// caller's responsibility and is performed by Params.CanonicalDiameter.
func CheckInputs(sigma0, ky, d float64) error {
	if sigma0 < 0 {
		return &InputError{Field: "sigma0", Value: sigma0, Reason: "must be >= 0"}
	}
	if ky <= 0 {
		return &InputError{Field: "ky", Value: ky, Reason: "must be > 0"}
	}
	if d <= 0 {
		return &InputError{Field: "d", Value: d, Reason: "must be > 0"}
	}
	return nil
}

// Validate is the convenience entry point used by the CLI and HTTP layers: it
// canonicalises the diameter from the supplied unit and then runs CheckInputs.
func (p Params) Validate() error {
	d, err := p.CanonicalDiameter()
	if err != nil {
		return err
	}
	return CheckInputs(p.Sigma0, p.Ky, d)
}

package core

import "fmt"

// NewInputError builds an InputError for the given field, value and reason. It
// is provided so that callers in other packages can reuse the same structured
// error type without importing fmt directly.
func NewInputError(field string, value float64, reason string) *InputError {
	return &InputError{Field: field, Value: value, Reason: reason}
}

// ErrDiameterNonPositive is returned when a grain diameter is zero or negative.
func ErrDiameterNonPositive(d float64) error {
	return &InputError{Field: "d", Value: d, Reason: "must be > 0"}
}

// ErrCoefficientNonPositive is returned when ky is zero or negative.
func ErrCoefficientNonPositive(ky float64) error {
	return &InputError{Field: "ky", Value: ky, Reason: "must be > 0"}
}

// ErrFrictionNegative is returned when the friction stress is negative.
func ErrFrictionNegative(sigma0 float64) error {
	return &InputError{Field: "sigma0", Value: sigma0, Reason: "must be >= 0"}
}

// Message extracts a stable, user-facing message from any error, prefixing the
// structured InputError fields when present. This is what the HTTP and CLI
// layers surface to the caller.
func Message(err error) string {
	if err == nil {
		return ""
	}
	if ie, ok := err.(*InputError); ok {
		return fmt.Sprintf("invalid %s: %s", ie.Field, ie.Reason)
	}
	return err.Error()
}

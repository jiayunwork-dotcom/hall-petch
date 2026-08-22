package core

import (
	"encoding/json"
	"fmt"
	"os"
)

// Example is the on-disk representation of a Hall-Petch material case. It mirrors
// the JSON files kept under example/ and is the payload the web front-end loads
// before computing.
type Example struct {
	// Name is a short identifier for the material, e.g. "mild-steel".
	Name string `json:"name"`
	// Description is a free-form note about the material.
	Description string `json:"description,omitempty"`
	// Sigma0 is the friction stress in MPa.
	Sigma0 float64 `json:"sigma0"`
	// Ky is the strengthening coefficient in MPa*m^(1/2).
	Ky float64 `json:"ky"`
	// D is the mean grain diameter in the unit named by DUnit.
	D float64 `json:"d"`
	// DUnit is the unit of D ("m", "mm", "um", ...).
	DUnit string `json:"d_unit"`
	// KyUnit is a descriptive unit string for Ky, kept for documentation only.
	KyUnit string `json:"ky_unit,omitempty"`
}

// ToParams converts the example into kernel parameters.
func (e *Example) ToParams() Params {
	return Params{Sigma0: e.Sigma0, Ky: e.Ky, D: e.D, DUnit: e.DUnit}
}

// LoadExample reads and parses an example JSON file from disk.
func LoadExample(path string) (*Example, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read example %q: %w", path, err)
	}
	return ParseExample(raw)
}

// ParseExample parses example JSON from an in-memory byte slice.
func ParseExample(raw []byte) (*Example, error) {
	var e Example
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil, fmt.Errorf("parse example: %w", err)
	}
	if e.DUnit == "" {
		e.DUnit = "um"
	}
	if err := e.ToParams().Validate(); err != nil {
		return nil, fmt.Errorf("example %q invalid: %w", e.Name, err)
	}
	return &e, nil
}

// Yield evaluates the loaded example and returns the resulting yield strength.
func (e *Example) Yield() (float64, error) {
	return e.ToParams().Yield()
}

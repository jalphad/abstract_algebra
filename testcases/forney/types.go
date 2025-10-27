package forney

import (
	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
)

// ForneyTestInput defines the input for a Forney algorithm test
type ForneyTestInput struct {
	Prime             int16    `json:"prime"`              // Base prime p
	Degree            int      `json:"degree"`             // Extension degree n
	IrreducibleCoeffs []int    `json:"irreducible_coeffs"` // Irreducible polynomial coefficients
	Syndromes         []string `json:"syndromes"`          // Syndrome values as strings
	LambdaCoeffs      []string `json:"lambda_coeffs"`      // Λ(x) coefficients as strings
	LambdaDegree      int      `json:"lambda_degree"`      // Degree of Λ(x)
	ErrorPositions    []int    `json:"error_positions"`    // Known error positions from Chien search
}

// ForneyTestResponse contains the Forney algorithm results
type ForneyTestResponse struct {
	OmegaCoeffs     []gfpn.Element `json:"omega_coeffs"`     // Computed Ω(x) coefficients
	ErrorMagnitudes []string       `json:"error_magnitudes"` // Computed error magnitudes as strings
}

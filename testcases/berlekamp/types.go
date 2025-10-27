package berlekamp

// BerlekampTestInput defines the input for a Berlekamp-Massey test
type BerlekampTestInput struct {
	Prime             int16    `json:"prime"`              // Base prime p
	Degree            int      `json:"degree"`             // Extension degree n
	IrreducibleCoeffs []int    `json:"irreducible_coeffs"` // Irreducible polynomial coefficients
	Syndromes         []string `json:"syndromes"`          // Syndrome sequence as strings
}

// BerlekampTestResponse contains the Berlekamp-Massey result
type BerlekampTestResponse struct {
	ErrorLocator []string `json:"error_locator"` // Λ(x) coefficients as strings
	Degree       int      `json:"degree"`        // Degree of Λ(x)
}

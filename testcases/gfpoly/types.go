package gfpoly

// GFPolyTestInput defines the input for a polynomial over GF(p^n) test
type GFPolyTestInput struct {
	Prime             int16           `json:"prime"`              // Base prime p
	Degree            int             `json:"degree"`             // Extension degree n
	IrreducibleCoeffs []int           `json:"irreducible_coeffs"` // Irreducible polynomial coefficients
	Operations        []PolyOperation `json:"operations"`         // Operations to perform
}

// PolyOperation represents a single polynomial operation
type PolyOperation struct {
	Op     string `json:"op"`               // "add", "sub", "mul", "scalar_mul", "derivative", "divmod", "eval"
	Poly1  []int  `json:"poly1"`            // First polynomial (coefficients as element indices)
	Poly2  []int  `json:"poly2,omitempty"`  // Second polynomial (for binary ops)
	Scalar int    `json:"scalar,omitempty"` // Scalar value (for scalar_mul)
	Point  int    `json:"point,omitempty"`  // Evaluation point (for eval)
}

// GFPolyTestResponse contains the results of polynomial operations
type GFPolyTestResponse struct {
	Results []PolyResult `json:"results"`
}

// PolyResult represents the result of a polynomial operation
type PolyResult struct {
	Polynomial []string `json:"polynomial,omitempty"` // Result polynomial coefficients
	Quotient   []string `json:"quotient,omitempty"`   // Quotient (for divmod)
	Remainder  []string `json:"remainder,omitempty"`  // Remainder (for divmod)
	Value      string   `json:"value,omitempty"`      // Evaluation result
}

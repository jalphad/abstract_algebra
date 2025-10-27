package gfpn

// GFPNTestInput defines the input for a GF(p^n) field test
type GFPNTestInput struct {
	Prime             int16           `json:"prime"`              // Base prime p
	Degree            int             `json:"degree"`             // Extension degree n
	IrreducibleCoeffs []int           `json:"irreducible_coeffs"` // Irreducible polynomial coefficients
	Operations        []GFPNOperation `json:"operations"`
}

// GFPNOperation represents a single field operation to perform
type GFPNOperation struct {
	Op   string `json:"op"`   // "add", "sub", "mul", "div", "inv", "neg"
	Arg1 int    `json:"arg1"` // First operand (element index 0 to p^n-1)
	Arg2 int    `json:"arg2"` // Second operand (element index 0 to p^n-1), ignored for unary ops
}

// GFPNTestResponse contains the results of field operations
type GFPNTestResponse struct {
	Results []string `json:"results"` // String representation of resulting elements
}

package syndrome

// SyndromeTestInput defines the input for a syndrome calculation test
type SyndromeTestInput struct {
	Prime             int16  `json:"prime"`              // Base prime p
	Degree            int    `json:"degree"`             // Extension degree n
	IrreducibleCoeffs []int  `json:"irreducible_coeffs"` // Irreducible polynomial coefficients
	Received          []byte `json:"received"`           // Received codeword as bytes
	NumECSymbols      int    `json:"num_ec_symbols"`     // Number of error correction symbols
	GeneratorRootIdx  int    `json:"generator_root_idx"` // Index of generator root (usually 2 for α)
}

// SyndromeTestResponse contains the syndrome calculation result
type SyndromeTestResponse struct {
	Syndromes []string `json:"syndromes"`  // Syndrome values as strings
	HasErrors bool     `json:"has_errors"` // Whether errors were detected
}

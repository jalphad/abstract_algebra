package chien

// ChienTestInput defines the input for a Chien search test
type ChienTestInput struct {
	Prime             int16    `json:"prime"`              // Base prime p
	Degree            int      `json:"degree"`             // Extension degree n
	IrreducibleCoeffs []int    `json:"irreducible_coeffs"` // Irreducible polynomial coefficients
	LambdaCoeffs      []string `json:"lambda_coeffs"`      // Λ(x) coefficients as strings
	LambdaDegree      int      `json:"lambda_degree"`      // Degree of Λ(x)
	CodewordLength    int      `json:"codeword_length"`    // Length of codeword to search
}

// ChienTestResponse contains the Chien search result
type ChienTestResponse struct {
	ErrorPositions []int `json:"error_positions"` // Positions where Λ(α^{-j}) = 0
}

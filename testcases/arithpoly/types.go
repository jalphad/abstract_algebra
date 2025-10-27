package arithpoly

// PolyDivTestInput defines the input for a polynomial division test
type PolyDivTestInput struct {
	Prime    int16   `json:"prime"`
	Dividend []int16 `json:"dividend"` // Coefficients as int16 values
	Divisor  []int16 `json:"divisor"`  // Coefficients as int16 values
}

// PolyDivTestResponse contains the results of polynomial division
type PolyDivTestResponse struct {
	Quotient  []int16 `json:"quotient"`
	Remainder []int16 `json:"remainder"`
}

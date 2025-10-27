package gf

// FieldTestInput defines the input for a field test
type FieldTestInput struct {
	Prime      int16            `json:"prime"`
	Operations []FieldOperation `json:"operations"`
}

// FieldOperation represents a single field operation to perform
type FieldOperation struct {
	Op   string `json:"op"`   // "add", "sub", "mul", "div"
	Arg1 int    `json:"arg1"` // First operand value
	Arg2 int    `json:"arg2"` // Second operand value
}

// FieldTestResponse contains the results of field operations
type FieldTestResponse struct {
	Results []int16 `json:"results"`
}

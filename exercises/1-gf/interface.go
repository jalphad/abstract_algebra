package gf

type Field interface {
	Elements() []Element
	Element(int) Element
	Add(e1, e2 Element) Element
	Sub(e1, e2 Element) Element
	Mul(e1, e2 Element) Element
	Div(e1, e2 Element) Element
}

// Element defines the interface for an element in a finite field GF(p).
// It supports basic arithmetic operations: addition, subtraction,
// multiplication, and division.
type Element interface {
	Field() Field
	Value() int16
	Add(e Element) Element
	Sub(e Element) Element
	Mul(e Element) Element
	Div(e Element) Element
}

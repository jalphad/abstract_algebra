package gfpn

// Field represents a finite field GF(p^n)
type Field interface {
	// Elements returns all elements in the field
	Elements() []Element

	// Element creates an element from an integer index 0 to p^n-1
	// 0 maps to the zero element
	// 1 maps to 1 (the multiplicative identity)
	// i > 1 maps to α^(i-1) for i = 2, 3, ..., p^n-1
	Element(value int) Element

	// Zero returns the additive identity (zero element)
	Zero() Element

	// One returns the multiplicative identity (one element)
	One() Element

	// Primitive returns the primitive element that generates the multiplicative group
	Primitive() Element

	// Add performs addition of two field elements
	Add(e1, e2 Element) Element

	// Sub performs subtraction of two field elements
	Sub(e1, e2 Element) Element

	// Mul performs multiplication of two field elements
	Mul(e1, e2 Element) Element

	// Div performs division of two field elements
	Div(e1, e2 Element) Element

	// Order returns p^n (the number of elements in the field)
	Order() int
}

// Element represents an element in GF(p^n)
type Element interface {
	// IsZero returns true if this is the zero element
	IsZero() bool

	// String returns a pretty-printed representation of the element
	String() string

	// Add performs addition with another element
	Add(e Element) Element

	// Sub performs subtraction with another element
	Sub(e Element) Element

	// Mul performs multiplication with another element
	Mul(e Element) Element

	// Div performs division by another element
	Div(e Element) Element
}

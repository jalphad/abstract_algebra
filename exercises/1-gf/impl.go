package gf

import "fmt"

// NewField creates and returns a new finite field GF(p).
// It is the factory for creating fields. The prime p must be of type int16.
// Note: This function assumes p is a prime number.
func NewField(p int16) Field {
	if p <= 1 {
		panic("p must be a prime number greater than 1")
	}
	return &field{p: p}
}

// field represents the finite field GF(p).
// It holds the prime modulus p.
type field struct {
	p int16
}

func (f *field) Add(e1, e2 Element) Element {
	return e1.Add(e2)
}

func (f *field) Sub(e1, e2 Element) Element {
	return e1.Sub(e2)
}

func (f *field) Mul(e1, e2 Element) Element {
	return e1.Mul(e2)
}

func (f *field) Div(e1, e2 Element) Element {
	return e1.Div(e2)
}

// Element creates a new element with the given value in the context of the field.
// The value is reduced modulo p.
func (f *field) Element(value int) Element {
	p := int(f.p)
	// Reduce the value to be within the field [0, p-1]
	// The expression (value % p + p) % p correctly handles negative values.
	val := int16((value%p + p) % p)
	return &fieldElement{
		value: val,
		field: f,
	}
}

func (f *field) Elements() []Element {
	var ret []Element
	for i := 0; i < int(f.p); i++ {
		ret = append(ret, &fieldElement{
			value: int16(i),
			field: f,
		})
	}
	return ret
}

// fieldElement is the concrete implementation of the Element interface.
// It stores its value and a pointer to the field it belongs to.
type fieldElement struct {
	value int16
	field *field
}

func (a *fieldElement) Field() Field {
	return a.field
}

func (a *fieldElement) Value() int16 {
	return a.value
}

// assertSameField checks if two elements belong to the same field.
// It panics if they do not. It also performs a type assertion.
func (a *fieldElement) assertSameField(e Element) *fieldElement {
	b, ok := e.(*fieldElement)
	if !ok {
		panic("invalid element type")
	}
	if a.field.p != b.field.p {
		panic(fmt.Sprintf("elements are from different fields: GF(%d) and GF(%d)", a.field.p, b.field.p))
	}
	return b
}

// Add performs addition of two field elements: (a + b) mod p.
func (a *fieldElement) Add(e Element) Element {
	// b := a.assertSameField(e)
	panic("unimplemented")
}

// Sub performs subtraction of two field elements: (a - b) mod p.
func (a *fieldElement) Sub(e Element) Element {
	//b := a.assertSameField(e)
	panic("unimplemented")
}

// Mul performs multiplication of two field elements: (a * b) mod p.
func (a *fieldElement) Mul(e Element) Element {
	//b := a.assertSameField(e)
	panic("unimplemented")
}

// Div performs division of two field elements: (a * b^-1) mod p.
// It panics if division by zero is attempted.
func (a *fieldElement) Div(e Element) Element {
	//b := a.assertSameField(e)
	panic("unimplemented")
}

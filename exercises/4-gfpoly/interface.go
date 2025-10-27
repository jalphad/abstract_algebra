package gfpoly

import (
	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
)

// Polynomial represents a polynomial with coefficients in GF(p^n)
// Coefficients are stored from lowest degree to highest degree
// For example, [a0, a1, a2] represents a0 + a1*x + a2*x^2
type Polynomial interface {
	// Coefficients returns the polynomial coefficients from lowest to highest degree
	Coefficients() []gfpn.Element

	// Degree returns the degree of the polynomial (-1 for zero polynomial)
	Degree() int

	// Evaluate evaluates the polynomial at a given point
	Evaluate(x gfpn.Element) gfpn.Element

	// IsZero returns true if this is the zero polynomial
	IsZero() bool

	// Field returns the underlying field
	Field() gfpn.Field
}

package arithpoly

import (
	"github.com/jalphad/abstract_algebra/exercises/1-gf"
)

// Polynomial represents a polynomial with coefficients in GF(p)
// Coefficients are stored from lowest to highest degree
// e.g., [c0, c1, c2] represents c0 + c1*x + c2*x^2
type Polynomial []gf.Element

// PolyMul multiplies two polynomials over GF(p)
func PolyMul(field gf.Field, p1, p2 Polynomial) Polynomial {
	panic("not implemented")
}

// PolyDiv performs polynomial long division
// Returns quotient and remainder such that dividend = divisor * quotient + remainder
// Panics if divisor is zero polynomial
// field parameter is the GF(p) field that the coefficients belong to
func PolyDiv(field gf.Field, dividend, divisor Polynomial) (quotient, remainder Polynomial) {
	panic("not implemented")
}

// degree returns the degree of the polynomial (-1 for zero polynomial)
func degree(p Polynomial) int {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i].Value() != 0 {
			return i
		}
	}
	return -1
}

// trimPoly removes leading zero coefficients
func trimPoly(p Polynomial) Polynomial {
	deg := degree(p)
	if deg < 0 {
		return Polynomial{}
	}
	return p[:deg+1]
}

// isZeroPoly checks if polynomial is zero
func isZeroPoly(p Polynomial) bool {
	for _, coeff := range p {
		if coeff.Value() != 0 {
			return false
		}
	}
	return true
}

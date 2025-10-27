package gfpoly

import "github.com/jalphad/abstract_algebra/exercises/3-gfpn"

// polynomial is a concrete implementation of the Polynomial interface
type polynomial struct {
	coeffs []gfpn.Element
	field  gfpn.Field
}

// normalize removes leading zero coefficients
func normalize(field gfpn.Field, coeffs []gfpn.Element) []gfpn.Element {
	// Find the last non-zero coefficient
	lastNonZero := -1

	for i := len(coeffs) - 1; i >= 0; i-- {
		if !coeffs[i].IsZero() {
			lastNonZero = i
			break
		}
	}

	// If all coefficients are zero, return empty slice
	if lastNonZero == -1 {
		return []gfpn.Element{}
	}

	return coeffs[:lastNonZero+1]
}

// findDegree finds the degree of a coefficient array using IsZero()
// Returns -1 for zero polynomial
func findDegree(coeffs []gfpn.Element) int {
	for i := len(coeffs) - 1; i >= 0; i-- {
		if !coeffs[i].IsZero() {
			return i
		}
	}
	return -1
}

// NewPolynomial creates a new polynomial from coefficients
// coeffs are from lowest degree to highest degree
func NewPolynomial(field gfpn.Field, coeffs []gfpn.Element) Polynomial {
	normalizedCoeffs := normalize(field, coeffs)

	return &polynomial{
		coeffs: normalizedCoeffs,
		field:  field,
	}
}

// Coefficients returns the polynomial coefficients from lowest to highest degree
func (p *polynomial) Coefficients() []gfpn.Element {
	// Return a copy to prevent external modification
	result := make([]gfpn.Element, len(p.coeffs))
	copy(result, p.coeffs)
	return result
}

// Degree returns the degree of the polynomial (-1 for zero polynomial)
func (p *polynomial) Degree() int {
	if len(p.coeffs) == 0 {
		return -1
	}
	return len(p.coeffs) - 1
}

// Evaluate evaluates the polynomial at a given point
func (p *polynomial) Evaluate(x gfpn.Element) gfpn.Element {
	if len(p.coeffs) == 0 {
		return p.field.Zero()
	}

	result := p.coeffs[len(p.coeffs)-1]
	for i := len(p.coeffs) - 2; i >= 0; i-- {
		result = p.field.Add(p.field.Mul(result, x), p.coeffs[i])
	}

	return result
}

// IsZero returns true if this is the zero polynomial
func (p *polynomial) IsZero() bool {
	return len(p.coeffs) == 0
}

// Field returns the underlying field
func (p *polynomial) Field() gfpn.Field {
	return p.field
}

// Add adds two polynomials
func Add(p1, p2 Polynomial) Polynomial {
	if p1.Field() != p2.Field() {
		panic("polynomials must be over the same field")
	}

	panic("not implemented")
}

// Subtract subtracts two polynomials
func Subtract(p1, p2 Polynomial) Polynomial {
	if p1.Field() != p2.Field() {
		panic("polynomials must be over the same field")
	}

	panic("not implemented")
}

// Multiply multiplies two polynomials
func Multiply(p1, p2 Polynomial) Polynomial {
	if p1.Field() != p2.Field() {
		panic("polynomials must be over the same field")
	}

	field := p1.Field()

	// Handle zero polynomials
	if p1.IsZero() || p2.IsZero() {
		return NewPolynomial(field, []gfpn.Element{})
	}

	coeffs1 := p1.Coefficients()
	coeffs2 := p2.Coefficients()

	// Result has degree deg(p1) + deg(p2)
	resultLen := len(coeffs1) + len(coeffs2) - 1
	result := make([]gfpn.Element, resultLen)

	// Initialize all coefficients to zero
	for i := range result {
		result[i] = field.Zero()
	}

	// Multiply using convolution
	for i := 0; i < len(coeffs1); i++ {
		for j := 0; j < len(coeffs2); j++ {
			product := field.Mul(coeffs1[i], coeffs2[j])
			result[i+j] = field.Add(result[i+j], product)
		}
	}

	return NewPolynomial(field, result)
}

// ScalarMultiply multiplies a polynomial by a scalar
func ScalarMultiply(scalar gfpn.Element, p Polynomial) Polynomial {
	field := p.Field()
	coeffs := p.Coefficients()

	// Multiply each coefficient by the scalar
	result := make([]gfpn.Element, len(coeffs))
	for i, c := range coeffs {
		result[i] = field.Mul(scalar, c)
	}

	return NewPolynomial(field, result)
}

// FormalDerivative computes the formal derivative of a polynomial
// For p(x) = a0 + a1*x + a2*x^2 + ... + an*x^n
// p'(x) = a1 + 2*a2*x + 3*a3*x^2 + ... + n*an*x^(n-1)
func FormalDerivative(p Polynomial) Polynomial {
	field := p.Field()
	coeffs := p.Coefficients()

	// Zero polynomial has zero derivative
	if len(coeffs) == 0 {
		return NewPolynomial(field, []gfpn.Element{})
	}

	// Constant polynomial has zero derivative
	if len(coeffs) == 1 {
		return NewPolynomial(field, []gfpn.Element{})
	}

	// Compute derivative coefficients
	result := make([]gfpn.Element, len(coeffs)-1)
	for i := 1; i < len(coeffs); i++ {
		// Multiply coefficient by its degree (i)
		// We need to add the coefficient to itself i times
		coeff := field.Zero()
		for j := 0; j < i; j++ {
			coeff = field.Add(coeff, coeffs[i])
		}
		result[i-1] = coeff
	}

	return NewPolynomial(field, result)
}

// Divide performs polynomial division with remainder
// Returns quotient and remainder such that dividend = divisor * quotient + remainder
// and degree(remainder) < degree(divisor)
//
// Optimized implementation that works directly with coefficient arrays to minimize allocations
func Divide(dividend, divisor Polynomial) (quotient, remainder Polynomial) {
	if dividend.Field() != divisor.Field() {
		panic("polynomials must be over the same field")
	}

	if divisor.IsZero() {
		panic("division by zero polynomial")
	}

	field := dividend.Field()

	// Work with coefficient slices directly
	divCoeffs := divisor.Coefficients()
	divisorDeg := len(divCoeffs) - 1
	dividendDeg := dividend.Degree()

	// If dividend degree < divisor degree, quotient is zero
	if dividendDeg < divisorDeg {
		return NewPolynomial(field, []gfpn.Element{}), dividend
	}

	// Create remainder working buffer (mutable copy of dividend coefficients)
	remCoeffs := dividend.Coefficients()

	// Allocate quotient coefficients
	quotientLen := dividendDeg - divisorDeg + 1
	quotCoeffs := make([]gfpn.Element, quotientLen)
	for i := range quotCoeffs {
		quotCoeffs[i] = field.Zero()
	}

	// Cache the leading coefficient of divisor
	leadingCoeff := divCoeffs[divisorDeg]

	// Manual degree tracking (updated after each subtraction)
	remDeg := findDegree(remCoeffs)

	// Perform long division in-place
	for remDeg >= divisorDeg {
		// Compute the next quotient coefficient
		// quotientCoeff = leadingCoeff(remainder) / leadingCoeff(divisor)
		quotientCoeff := field.Div(remCoeffs[remDeg], leadingCoeff)

		// Store in quotient
		quotientIdx := remDeg - divisorDeg
		quotCoeffs[quotientIdx] = quotientCoeff

		// Subtract divisor * quotientCoeff * x^quotientIdx from remainder
		// This is done in-place on remCoeffs
		offset := quotientIdx
		for i := 0; i <= divisorDeg; i++ {
			sub := field.Mul(quotientCoeff, divCoeffs[i])
			remCoeffs[offset+i] = field.Sub(remCoeffs[offset+i], sub)
		}

		curRemDeg := remDeg

		// Update degree manually (find new leading coefficient)
		// We know remCoeffs[remDeg] is now zero, so start from remDeg-1
		remDeg = -1
		for i := len(remCoeffs) - 1; i >= 0; i-- {
			if !remCoeffs[i].IsZero() {
				remDeg = i
				break
			}
		}

		// If remDeg is -1, remainder is zero - break early
		if remDeg == -1 {
			break
		}

		if curRemDeg == remDeg {
			panic("same remaining degree after iteration of long division")
		}
	}

	// Create polynomial objects only once at the end
	quotient = NewPolynomial(field, quotCoeffs)
	remainder = NewPolynomial(field, remCoeffs)
	return quotient, remainder
}

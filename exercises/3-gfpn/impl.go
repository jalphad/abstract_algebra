package gfpn

import (
	"fmt"
	"strings"

	"github.com/jalphad/abstract_algebra/exercises/1-gf"
	"github.com/jalphad/abstract_algebra/exercises/2-arithpoly"
)

// field implements the Field interface for GF(p^n)
type field struct {
	baseField        gf.Field
	degree           int
	order            int // p^n
	irreducible      arithpoly.Polynomial
	powerToPoly      [][]gf.Element // index i contains polynomial repr of α^i
	polyToPower      map[string]int // maps polynomial repr to power
	zeroElement      *element
	oneElement       *element
	primitiveElement *element
}

// NewField creates a new GF(p^n) field
// p must be prime, n >= 1
// irreducible is the irreducible polynomial of degree n used to construct the field
// irreducible should be given as coefficients [a0, a1, ..., an] where an = 1
func NewField(p int16, n int, irreducibleCoeffs []int) (Field, error) {
	if p <= 1 {
		return nil, fmt.Errorf("p must be a prime greater than 1")
	}
	if n < 1 {
		return nil, fmt.Errorf("n must be at least 1")
	}
	if len(irreducibleCoeffs) != n+1 {
		return nil, fmt.Errorf("irreducible polynomial must have degree n (need n+1 coefficients)")
	}
	if irreducibleCoeffs[n] != 1 {
		return nil, fmt.Errorf("irreducible polynomial must be monic (leading coefficient must be 1)")
	}

	baseField := gf.NewField(p)

	// Convert irreducible coefficients to polynomial
	irreducible := make(arithpoly.Polynomial, n+1)
	for i, c := range irreducibleCoeffs {
		irreducible[i] = baseField.Element(c)
	}

	// Calculate order
	order := 1
	for i := 0; i < n; i++ {
		order *= int(p)
	}

	f := &field{
		baseField:   baseField,
		degree:      n,
		order:       order,
		irreducible: irreducible,
		powerToPoly: make([][]gf.Element, order-1), // α^0 to α^(p^n-2)
		polyToPower: make(map[string]int),
	}

	// Find a primitive element (generator of the multiplicative group)
	primitiveElement, err := f.findPrimitiveElement()
	if err != nil {
		return nil, fmt.Errorf("failed to find primitive element: %w", err)
	}
	f.primitiveElement = &element{
		field:  f,
		power:  1,
		coeffs: primitiveElement,
	}

	// Build lookup tables using the primitive element
	if err := f.buildTables(primitiveElement); err != nil {
		return nil, err
	}

	// Create zero and one elements
	zeroCoeffs := make([]gf.Element, n)
	for i := 0; i < n; i++ {
		zeroCoeffs[i] = baseField.Element(0)
	}
	f.zeroElement = &element{
		field:  f,
		power:  -1,
		coeffs: zeroCoeffs,
	}

	oneCoeffs := make([]gf.Element, n)
	oneCoeffs[0] = baseField.Element(1)
	for i := 1; i < n; i++ {
		oneCoeffs[i] = baseField.Element(0)
	}
	f.oneElement = &element{
		field:  f,
		power:  0,
		coeffs: oneCoeffs,
	}

	return f, nil
}

// computeOrder computes the multiplicative order of a polynomial element
// Returns the smallest positive integer k such that element^k ≡ 1 (mod irreducible)
func (f *field) computeOrder(element arithpoly.Polynomial) (int, error) {
	// Start with element^1
	current := make(arithpoly.Polynomial, len(element))
	copy(current, element)

	one := make(arithpoly.Polynomial, f.degree)
	one[0] = f.baseField.Element(1)
	for i := 1; i < f.degree; i++ {
		one[i] = f.baseField.Element(0)
	}

	for order := 1; order <= f.order; order++ {
		// Check if current == 1
		if polyKey(current) == polyKey(one) {
			return order, nil
		}

		// Multiply by element and reduce
		next := arithpoly.PolyMul(f.baseField, current, element)
		_, remainder := arithpoly.PolyDiv(f.baseField, next, f.irreducible)

		// Normalize remainder to degree-1 coefficients
		current = make(arithpoly.Polynomial, f.degree)
		copy(current, remainder)
		for i := len(remainder); i < f.degree; i++ {
			current[i] = f.baseField.Element(0)
		}
	}

	return 0, fmt.Errorf("failed to find order (element may be zero)")
}

// findPrimitiveElement searches for a primitive element (generator) of the multiplicative group
// A primitive element has order p^n - 1
func (f *field) findPrimitiveElement() (arithpoly.Polynomial, error) {
	targetOrder := f.order - 1

	// Try various candidates
	// For degree 1 (GF(p^1) = GF(p)), use 2 as generator (except for p=2, use 1)
	if f.degree == 1 {
		elem := make(arithpoly.Polynomial, 1)
		// For p=2, use 1 (only element in GF(2)*)
		// For p>2, use 2 as the generator
		if len(f.baseField.Elements()) == 2 {
			elem[0] = f.baseField.Element(1)
		} else {
			elem[0] = f.baseField.Element(2)
		}
		return elem, nil
	}

	// Start with α (represented as [0, 1, 0, ..., 0])
	candidates := []arithpoly.Polynomial{
		// α
		func() arithpoly.Polynomial {
			p := make(arithpoly.Polynomial, f.degree)
			p[1] = f.baseField.Element(1)
			for i := 0; i < f.degree; i++ {
				if i != 1 {
					p[i] = f.baseField.Element(0)
				}
			}
			return p
		}(),
	}

	// Get the prime p from base field
	p := int(f.baseField.Elements()[len(f.baseField.Elements())-1].Value()) + 1

	// Add more candidates: try all combinations of coefficients for lower degree terms
	// Only need to try coefficients in [0, p-1]
	for coeff0 := 0; coeff0 < p; coeff0++ {
		for coeff1 := 0; coeff1 < p; coeff1++ {
			if coeff0 == 0 && coeff1 == 1 {
				continue // Already added α
			}
			if coeff0 == 0 && coeff1 == 0 {
				continue // Skip zero
			}
			candidate := make(arithpoly.Polynomial, f.degree)
			candidate[0] = f.baseField.Element(coeff0)
			candidate[1] = f.baseField.Element(coeff1)
			for i := 2; i < f.degree; i++ {
				candidate[i] = f.baseField.Element(0)
			}
			candidates = append(candidates, candidate)
		}
	}

	for _, candidate := range candidates {
		order, err := f.computeOrder(candidate)
		if err != nil {
			continue
		}
		if order == targetOrder {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("no primitive element found")
}

// buildTables constructs the power-to-polynomial and polynomial-to-power lookup tables
// using the given primitive element as the generator
func (f *field) buildTables(primitiveElement arithpoly.Polynomial) error {
	// Start with α^0 = 1
	current := make(arithpoly.Polynomial, f.degree)
	current[0] = f.baseField.Element(1)
	for i := 1; i < f.degree; i++ {
		current[i] = f.baseField.Element(0)
	}

	for power := 0; power < f.order-1; power++ {
		// Store current polynomial
		coeffs := make([]gf.Element, f.degree)
		copy(coeffs, current)
		f.powerToPoly[power] = coeffs
		f.polyToPower[polyKey(coeffs)] = power

		// Compute next power: multiply by primitive element and reduce
		next := arithpoly.PolyMul(f.baseField, current, primitiveElement)

		// Reduce modulo irreducible polynomial
		_, remainder := arithpoly.PolyDiv(f.baseField, next, f.irreducible)

		// Ensure remainder has correct length
		if len(remainder) > f.degree {
			return fmt.Errorf("reduction error: remainder degree too high")
		}

		current = make(arithpoly.Polynomial, f.degree)
		copy(current, remainder)
		for i := len(remainder); i < f.degree; i++ {
			current[i] = f.baseField.Element(0)
		}
	}

	if len(f.polyToPower) != len(f.powerToPoly) {
		return fmt.Errorf("field generation error: power and poly representations do not match")
	}

	// Verify we cycled back to 1 (α^0)
	if polyKey(current) != polyKey(f.powerToPoly[0]) {
		return fmt.Errorf("field generation error: did not cycle back to 1")
	}

	return nil
}

// polyKey creates a string key for a polynomial representation
func polyKey(coeffs []gf.Element) string {
	var sb strings.Builder
	for i, c := range coeffs {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", c.Value()))
	}
	return sb.String()
}

func (f *field) Elements() []Element {
	elements := make([]Element, f.order)
	elements[0] = f.zeroElement
	for i := 0; i < f.order-1; i++ {
		elements[i+1] = &element{
			field:  f,
			power:  i,
			coeffs: f.powerToPoly[i],
		}
	}
	return elements
}

func (f *field) Element(value int) Element {
	// Use modulo arithmetic to handle values outside [0, p^n-1]
	value = ((value % f.order) + f.order) % f.order

	// 0 maps to zero element
	if value == 0 {
		return f.zeroElement
	}

	// 1 maps to one (α^0)
	// 2 maps to α (α^1)
	// 3 maps to α^2
	// ...
	// p^n-1 maps to α^(p^n-2)
	power := value - 1

	return &element{
		field:  f,
		power:  power,
		coeffs: f.powerToPoly[power],
	}
}

func (f *field) Zero() Element {
	return f.zeroElement
}

func (f *field) One() Element {
	return f.oneElement
}

func (f *field) Primitive() Element {
	return f.primitiveElement
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

func (f *field) Order() int {
	return f.order
}

// element implements the Element interface
type element struct {
	field  *field
	power  int          // -1 for zero, otherwise 0 to p^n-2
	coeffs []gf.Element // polynomial representation
}

func (e *element) Coefficients() []gf.Element {
	result := make([]gf.Element, len(e.coeffs))
	copy(result, e.coeffs)
	return result
}

func (e *element) Power() int {
	return e.power
}

func (e *element) IsZero() bool {
	return e.power == -1
}

func (e *element) String() string {
	if e.IsZero() {
		return "0"
	}

	// Print coefficients in most significant first order
	var coeffStrs string
	for i := len(e.coeffs) - 1; i >= 0; i-- {
		coeffStrs = coeffStrs + fmt.Sprintf("%d", e.coeffs[i].Value())
	}

	return coeffStrs
}

func (e *element) assertSameField(other Element) *element {
	otherElem, ok := other.(*element)
	if !ok {
		panic("invalid element type")
	}
	if e.field != otherElem.field {
		panic("elements are from different fields")
	}
	return otherElem
}

func (e *element) Add(other Element) Element {
	//o := e.assertSameField(other)
	panic("not implemented")
}

func (e *element) Sub(other Element) Element {
	//o := e.assertSameField(other)
	panic("not implemented")
}

func (e *element) Mul(other Element) Element {
	//o := e.assertSameField(other)
	panic("not implemented")
}

func (e *element) Div(other Element) Element {
	//o := e.assertSameField(other)
	panic("not implemented")
}

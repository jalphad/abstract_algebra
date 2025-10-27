# Exercise: Finite Field GF(p) Implementation

## Objective

Implement arithmetic operations for finite field GF(p) where p is prime.

## Background

A **finite field** (or Galois field) GF(p) consists of:
- Elements: {0, 1, 2, ..., p-1} where p is prime
- Operations: addition, subtraction, multiplication, and division (modulo p)
- All operations are closed (results stay in the field)

Unlike regular integer arithmetic, division always works in GF(p) (except by zero) because every non-zero element has a multiplicative inverse.

## Your Task

Implement the following methods in `field.go`:

### fieldElement.Add(e Element) Element
Simple addition modulo p.

### fieldElement.Sub(e Element) Element
Find additive inverse of e and do addition modulo p.
Alternatively, just do subtraction modulo p but adjust for negative values

### 4. fieldElement.Mul(e Element) Element
Simple multiplication modulo p
- Use int32 for intermediate calculations

### 5. fieldElement.Div(e Element) Element
Divide by finding the multiplicative inverse.

Division `a / b` in GF(p) means finding `b⁻¹` such that `b × b⁻¹ ≡ 1 (mod p)`, then computing `a × b⁻¹ mod p`.

## Division Implementation: Two Approaches

### Approach 1: Extended Euclidean Algorithm (1-pass)

The Extended Euclidean Algorithm computes gcd(a, b) and the Bézout coefficients x, y such that:
```
ax + by = gcd(a, b)
```

For finding the multiplicative inverse of `b` in GF(p):
- Since p is prime and b ≠ 0, we have gcd(b, p) = 1
- We want to find x such that: `bx + py = 1`
- Taking mod p: `bx ≡ 1 (mod p)`
- Therefore: `x = b⁻¹ mod p`

#### Single-Pass Algorithm

```
function extendedGCD(a, b):
    // Initialize
    old_r, r = a, b
    old_s, s = 1, 0

    // Iterate until remainder is 0
    while r ≠ 0:
        quotient = old_r / r

        // Update r (remainder)
        old_r, r = r, old_r - quotient × r

        // Update s (Bézout coefficient)
        old_s, s = s, old_s - quotient × s

    // Result: old_r is gcd, old_s is the coefficient we want
    return old_s
```

#### To find b⁻¹ in GF(p):
```go
inverse := extendedGCD(b, p)
// Ensure positive result
if inverse < 0 {
    inverse += p
}
// Now inverse ≡ b⁻¹ (mod p)
```

**Example:** Find 3⁻¹ in GF(7)

| Step | old_r | r | quotient | old_s | s |
|------|-------|---|----------|-------|---|
| Init | 3     | 7 | -        | 1     | 0 |
| 1    | 7     | 3 | 0        | 0     | 1 |
| 2    | 3     | 1 | 2        | 1     | -2 |
| 3    | 1     | 0 | 3        | -2    | 7 |

Result: old_s = -2, adjusted to 5 (mod 7)
Verify: 3 × 5 = 15 ≡ 1 (mod 7) ✓

### Approach 2: Fermat's Little Theorem

**Fermat's Little Theorem** states that for prime p and a ≠ 0:
```
a^p ≡ a (mod p)
```

Dividing both sides by a:
```
a^(p-1) ≡ 1 (mod p)
```

Multiplying both sides by a⁻¹:
```
a^(p-2) ≡ a⁻¹ (mod p)
```

**Therefore:** To find b⁻¹, simply compute `b^(p-2) mod p`

#### Fast Exponentiation (Exponentiation by Squaring)

Computing large powers naively is slow. Use exponentiation by squaring:
- Binary representation of exponent
- Example: 3⁴ = 3² × 3² (square twice)
- Example: 3⁵ = 3¹ × 3⁴ = 3¹ × (3²)²

#### To find b⁻¹ in GF(p):
```go
inverse := modularExponentiation(b, p-2, p)
// Now inverse ≡ b⁻¹ (mod p)
```

**Example:** Find $3^{-1} in GF(7)
- Compute: 3^(7-2) = 3⁵ mod 7
- 3¹ = 3
- 3² = 9 ≡ 2 (mod 7)
- 3⁴ = (3²)² = 4 ≡ 4 (mod 7)
- 3⁵ = 3¹ × 3⁴ = 3 × 4 = 12 ≡ 5 (mod 7)

Verify: 3 × 5 = 15 ≡ 1 (mod 7) ✓

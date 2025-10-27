# Exercise: Syndrome Calculation for Reed-Solomon Codes

## Objective

Implement syndrome calculation, the first step in Reed-Solomon error correction decoding. Syndromes tell us whether errors occurred and provide the information needed to locate and correct them.

## Background

In Reed-Solomon coding, a **syndrome** is a special value computed by evaluating the received polynomial at specific points. These values:
- Equal zero when there are no errors
- Are non-zero when errors are present
- Encode information about error locations and magnitudes

## Your Task

Implement two functions in `interface.go`:

### 1. Calculate syndromes

Calculate the syndrome vector [S_0, S_1, ..., S_{2t-1}] where:
- S_i = r(α^i) = evaluate received polynomial at α^i
- α is the generator root (primitive element)
- 2t is the number of error correction symbols

**Algorithm**:
```
For each syndrome index i from 0 to numECSymbols-1:
  1. Compute α^i by repeated multiplication
  2. Evaluate the received polynomial at α^i using Horner's method
  3. Store result as S_i
```

**Horner's method** for evaluating r(x) at point p:
```
result = 0
for i from n-1 down to 0:
    result = result * p + r[i]
return result
```

### 2. HasErrors

Check if any syndrome is non-zero:
- Return `true` if at least one syndrome is non-zero (errors detected)
- Return `false` if all syndromes are zero (no errors detected)

## Implementation Tips

1. **Field Operations**: Use `field.Add()`, `field.Mul()`, `field.Zero()`, `field.One()`
2. **Power Computation**: Compute α^i by multiplying α by itself i times
3. **Horner's Method**: Most efficient way to evaluate polynomials
4. **Byte Conversion**: Convert bytes to field elements using `field.Element(int(byte))`
5. **Zero Check**: Use `element.IsZero()` to check if an element is zero

## Mathematical Example

Using **GF(4)** with elements {0, 1, α, α+1}:

**Given**:
- Received: [1, α, 1] representing r(x) = 1 + α·x + 1·x²
- numECSymbols: 2 (compute S_0 and S_1)
- generatorRoot: α

**Compute S_0 = r(α^0) = r(1)**:
```
r(1) = 1 + α·1 + 1·1²
     = 1 + α + 1
     = α  (since 1 + 1 = 0 in characteristic 2)
```

**Compute S_1 = r(α^1) = r(α)** using Horner's method:
```
Start:  result = 0
Step 1: result = 0·α + 1 = 1
Step 2: result = 1·α + α = α + α = 0  (in char 2)
Step 3: result = 0·α + 1 = 1
```

**Result**: Syndromes = [α, 1] → errors detected!

## Connection to Reed-Solomon Decoding

Syndrome calculation is **Step 1** of Reed-Solomon decoding:

1. **Calculate syndromes** ← You are here!
2. Use Berlekamp-Massey to find error locator polynomial
3. Use Chien search to find error positions
4. Use Forney algorithm to compute error values
5. Correct the errors

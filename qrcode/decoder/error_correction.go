package decoder

import (
	"fmt"

	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
	"github.com/jalphad/abstract_algebra/exercises/6-berlekamp"
	"github.com/jalphad/abstract_algebra/exercises/7-chien"
	"github.com/jalphad/abstract_algebra/exercises/8-forney"
	"github.com/jalphad/abstract_algebra/qrcode/correction"
	"github.com/jalphad/abstract_algebra/qrcode/types"
	"github.com/makiuchi-d/gozxing/qrcode/decoder"
)

// ErrorCorrector handles Reed-Solomon error correction for QR codes
//
// QR codes use Reed-Solomon error correction over GF(256), which allows them to
// recover from damage, dirt, or other visual corruption. This is why QR codes
// work even when partially obscured or damaged.
//
// Mathematical Background:
//   - Field: GF(2^8) = GF(256) constructed using irreducible polynomial x^8 + x^4 + x^3 + x^2 + 1
//   - Primitive element: α (root of the irreducible polynomial)
//   - Each byte (0-255) maps to a unique element of GF(256)
//
// Error Correction Capacity:
//   - Level L (Low):       ~7%  errors correctable  (2 EC codewords can fix 1 error)
//   - Level M (Medium):    ~15% errors correctable
//   - Level Q (Quartile):  ~25% errors correctable
//   - Level H (High):      ~30% errors correctable
//
// For example, a Version 1-L QR code has 26 total codewords:
//   - 19 data codewords
//   - 7 error correction codewords
//   - Can correct up to 3 symbol errors (7/2 = 3.5 → floor = 3)
type ErrorCorrector struct {
	field       gfpn.Field     // GF(256) field for QR code error correction
	alphaPowers []gfpn.Element // Precomputed powers of α: [α^0, α^1, ..., α^7]
}

// NewErrorCorrector creates a new error corrector for QR codes
//
// This initializes the GF(256) field using the QR code standard irreducible polynomial:
// x^8 + x^4 + x^3 + x^2 + 1 (binary: 100011101, hex: 0x11D)
//
// This specific polynomial is chosen because:
//   - It's irreducible over GF(2), meaning it can't be factored
//   - It generates a primitive element that cycles through all 255 non-zero elements
//   - It's standardized in the QR code specification (ISO/IEC 18004)
func NewErrorCorrector() (*ErrorCorrector, error) {
	// QR code irreducible polynomial: x^8 + x^4 + x^3 + x^2 + 1
	// Coefficients: [constant, x^1, x^2, x^3, x^4, x^5, x^6, x^7, x^8]
	//              = [1, 0, 1, 1, 1, 0, 0, 0, 1]
	// This corresponds to binary 100011101 = 0x11D
	irreducible := []int{1, 0, 1, 1, 1, 0, 0, 0, 1}

	field, err := gfpn.NewField(2, 8, irreducible)
	if err != nil {
		return nil, fmt.Errorf("failed to create GF(256) field: %w", err)
	}

	// Precompute powers of α for byte-to-element conversion
	// QR codes interpret byte bits as polynomial coefficients
	alpha := field.Primitive()
	alphaPowers := make([]gfpn.Element, 8)
	alphaPowers[0] = field.One() // α^0 = 1
	for i := 1; i < 8; i++ {
		alphaPowers[i] = field.Mul(alphaPowers[i-1], alpha)
	}

	return &ErrorCorrector{
		field:       field,
		alphaPowers: alphaPowers,
	}, nil
}

// byteToElement converts a byte to a GF(256) element using QR code's convention
//
// QR codes interpret bytes as polynomial coefficients in GF(2)[x]:
// Byte 0bB7B6B5B4B3B2B1B0 → polynomial B0 + B1·x + B2·x² + ... + B7·x⁷
//
// For example:
//   - 0x00 (00000000) → 0
//   - 0x01 (00000001) → 1 = α^0
//   - 0x02 (00000010) → x = α^1
//   - 0x20 (00100000) → x^5 = α^5
//
// This is computed as: Σ (bit_i · α^i) for i = 0 to 7
func (ec *ErrorCorrector) byteToElement(b byte) gfpn.Element {
	// Start with zero
	result := ec.field.Zero()

	// Add α^i for each set bit
	for i := 0; i < 8; i++ {
		if (b & (1 << i)) != 0 {
			result = ec.field.Add(result, ec.alphaPowers[i])
		}
	}

	return result
}

// elementToByte converts a GF(256) element back to a byte
//
// This is the reverse of byteToElement. We need to find which polynomial
// coefficients are set, then build the corresponding byte.
//
// Since we're in GF(256) = GF(2)[x]/(irreducible), each element has a unique
// polynomial representation with coefficients in {0, 1}.
func (ec *ErrorCorrector) elementToByte(elem gfpn.Element) byte {
	if elem.IsZero() {
		return 0
	}

	// Try all 256 possible bytes to find which one maps to this element
	// This is a brute-force approach but works for GF(256)
	// (More efficient implementations would use logarithm tables)
	for b := byte(1); b != 0; b++ { // b++ will wrap from 255 to 0
		if ec.byteToElement(b).String() == elem.String() {
			return b
		}
	}

	// This should never happen for valid GF(256) elements
	return 0
}

// CorrectCodewords performs Reed-Solomon error correction on QR code data
//
// This is the main entry point for error correction. It:
//  1. De-interleaves the raw codewords into separate RS blocks
//  2. Applies error correction to each block independently
//  3. Re-interleaves the corrected blocks
//  4. Returns the corrected data along with error statistics
//
// QR Code Block Structure:
//
// Higher version and higher error correction level QR codes split data into multiple blocks.
// For example, Version 5-H has 4 blocks. Each block is independently error-corrected.
//
// The codewords are interleaved to spread physical damage across multiple blocks,
// increasing robustness. For example, with 2 blocks:
//
//	Raw order:     [D1-block1, D1-block2, D2-block1, D2-block2, ..., EC1-block1, EC1-block2, ...]
//	Block 1 data:  [D1-block1, D2-block1, D3-block1, ...]
//	Block 2 data:  [D1-block2, D2-block2, D3-block2, ...]
//
// Parameters:
//   - qrData: Extracted QR code data from the extractor
//
// Returns:
//   - Corrected data codewords (de-interleaved and error-corrected)
//   - Block-by-block results showing where errors were found and corrected
//   - Error if correction fails
func (ec *ErrorCorrector) CorrectCodewords(qrData *types.QRCodeData) ([]byte, []BlockResult, error) {
	version := qrData.Version
	ecLevel := qrData.ECLevel
	rawCodewords := qrData.RawCodewords

	// Get the error correction block structure for this version and EC level
	ecBlocks := version.GetECBlocksForLevel(ecLevel)

	// De-interleave codewords into separate blocks
	blocks := ec.deinterleaveBlocks(rawCodewords, ecBlocks)

	// Correct each block independently
	correctedBlocks := make([][]byte, len(blocks))
	blockResults := make([]BlockResult, len(blocks))

	for i, block := range blocks {
		corrected, result, err := ec.correctBlock(block, ecBlocks, i)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to correct block %d: %w", i, err)
		}
		correctedBlocks[i] = corrected
		blockResults[i] = result
	}

	// Re-interleave corrected blocks to get final data
	correctedData := ec.reinterleaveBlocks(correctedBlocks, ecBlocks)

	return correctedData, blockResults, nil
}

// deinterleaveBlocks splits interleaved codewords into separate RS blocks
//
// QR codes interleave codewords from multiple blocks to improve error resilience.
// A localized physical damage will affect multiple blocks by a small amount,
// rather than destroying one block completely.
//
// This function reverses the interleaving process.
func (ec *ErrorCorrector) deinterleaveBlocks(rawCodewords []byte, ecBlocks *decoder.ECBlocks) [][]byte {
	// Get block structure
	ecbArray := ecBlocks.GetECBlocks()
	numBlocks := 0
	for _, ecb := range ecbArray {
		numBlocks += ecb.GetCount()
	}

	blocks := make([][]byte, numBlocks)
	blockIndex := 0

	// Calculate codewords per block
	for _, ecb := range ecbArray {
		numDataCodewords := ecb.GetDataCodewords()
		numECCodewords := ecBlocks.GetECCodewordsPerBlock()
		totalCodewords := numDataCodewords + numECCodewords

		for i := 0; i < ecb.GetCount(); i++ {
			blocks[blockIndex] = make([]byte, totalCodewords)
			blockIndex++
		}
	}

	// De-interleave data codewords
	// Data codewords are interleaved: D1-B1, D1-B2, ..., D2-B1, D2-B2, ...
	rawIndex := 0
	maxDataCodewords := 0
	for _, block := range blocks {
		numData := len(block) - ecBlocks.GetECCodewordsPerBlock()
		if numData > maxDataCodewords {
			maxDataCodewords = numData
		}
	}

	for i := 0; i < maxDataCodewords; i++ {
		for j := 0; j < len(blocks); j++ {
			numData := len(blocks[j]) - ecBlocks.GetECCodewordsPerBlock()
			if i < numData {
				blocks[j][i] = rawCodewords[rawIndex]
				rawIndex++
			}
		}
	}

	// De-interleave EC codewords
	// EC codewords are also interleaved: EC1-B1, EC1-B2, ..., EC2-B1, EC2-B2, ...
	numECCodewords := ecBlocks.GetECCodewordsPerBlock()
	for i := 0; i < numECCodewords; i++ {
		for j := 0; j < len(blocks); j++ {
			numData := len(blocks[j]) - numECCodewords
			blocks[j][numData+i] = rawCodewords[rawIndex]
			rawIndex++
		}
	}

	return blocks
}

// correctBlock performs Reed-Solomon error correction on a single block
//
// This implements the complete RS decoding pipeline:
//  1. Convert bytes to GF(256) elements
//  2. Compute syndromes
//  3. Run Berlekamp-Massey algorithm to find error locator polynomial
//  4. Run Chien search to find error positions
//  5. Run Forney algorithm to find error magnitudes
//  6. Apply corrections
//  7. Verify correction succeeded
//
// Educational Note:
// This is exactly the same algorithm we implemented in the reference code,
// now applied to real QR code data!
func (ec *ErrorCorrector) correctBlock(block []byte, ecBlocks *decoder.ECBlocks, blockIndex int) ([]byte, BlockResult, error) {
	numECCodewords := ecBlocks.GetECCodewordsPerBlock()
	numDataCodewords := len(block) - numECCodewords
	codewordLength := len(block)

	result := BlockResult{
		BlockIndex:       blockIndex,
		NumDataCodewords: numDataCodewords,
		NumECCodewords:   numECCodewords,
	}

	// Convert bytes to GF(256) elements
	received := make([]gfpn.Element, len(block))
	for i, b := range block {
		received[i] = ec.byteToElement(b)
	}

	// Step 1: Compute syndromes
	// Syndromes are computed as S_i = r(α^i) where r(x) is the received polynomial
	// If all syndromes are zero, there are no errors
	syndromes := ec.computeSyndromes(received, numECCodewords)

	// Check if there are any errors
	hasErrors := false
	for _, s := range syndromes {
		if !s.IsZero() {
			hasErrors = true
			break
		}
	}

	if !hasErrors {
		// No errors detected - return original data
		result.CorrectionSucceeded = true
		return block[:numDataCodewords], result, nil
	}

	// Step 2: Berlekamp-Massey Algorithm
	// Finds the error locator polynomial Λ(x) from syndromes
	// Λ(x) has roots at X_i^{-1} where X_i are the error locators
	lambda := berlekamp.BerlekampMassey(ec.field, syndromes)

	// Step 3: Compute error evaluator polynomial Ω(x)
	// Used in Forney's formula to compute error magnitudes
	omega := forney.ComputeOmega(ec.field, syndromes, lambda)

	// Step 4: Chien Search
	// Finds error positions by evaluating Λ(α^{-j}) for all j
	// Positions where Λ(α^{-j}) = 0 are error positions
	// Chien search returns positions in standard polynomial convention (position i = x^i)
	standardPositions := chien.ChienSearch(ec.field, lambda, codewordLength)

	result.ErrorsFound = len(standardPositions)
	result.ErrorPositions = standardPositions

	// Check if we found too many errors
	maxCorrectableErrors := numECCodewords / 2
	if len(standardPositions) > maxCorrectableErrors {
		result.CorrectionSucceeded = false
		return nil, result, fmt.Errorf("too many errors: found %d, can correct %d", len(standardPositions), maxCorrectableErrors)
	}

	// Step 5: Forney Algorithm
	// Computes error magnitudes using Forney's formula:
	// Y_i = X_i · Ω(X_i^{-1}) / Λ'(X_i^{-1})
	magnitudes := forney.ComputeErrorMagnitudes(ec.field, lambda, omega, standardPositions)

	// Step 6: Translate positions from standard to QR's reverse convention
	// In standard convention: position i means codeword[i] (x^i coefficient)
	// In QR's reverse convention: codeword[0] is highest degree, so position i means codeword[n-1-i]
	qrPositions := make([]int, len(standardPositions))
	for i, pos := range standardPositions {
		qrPositions[i] = codewordLength - 1 - pos
	}

	// Step 7: Apply corrections
	// corrected[j] = received[j] - Y_j (in GF(256), subtraction = addition)
	corrected := correction.ApplyCorrections(ec.field, received, qrPositions, magnitudes)

	// Step 7: Verify correction
	// Compute syndromes of corrected codeword - should all be zero
	// We use our own computeSyndromes which uses QR's reverse polynomial evaluation
	verifySyndromes := ec.computeSyndromes(corrected, numECCodewords)
	isValid := true
	for _, s := range verifySyndromes {
		if !s.IsZero() {
			isValid = false
			break
		}
	}
	if !isValid {
		result.CorrectionSucceeded = false
		return nil, result, fmt.Errorf("correction verification failed")
	}

	result.CorrectionSucceeded = true

	// Convert corrected elements back to bytes and return only data portion
	correctedBytes := make([]byte, numDataCodewords)
	for i := 0; i < numDataCodewords; i++ {
		// Convert GF(256) element back to byte using reverse lookup
		correctedBytes[i] = ec.elementToByte(corrected[i])
	}

	return correctedBytes, result, nil
}

// computeSyndromes calculates syndrome values for error detection
//
// Syndromes are computed by evaluating the received polynomial at consecutive
// powers of α (the primitive element):
//
//	S_i = r(α^i) for i = 0, 1, ..., 2t-1
//
// where t is the error correction capability (number of errors that can be corrected).
//
// If all syndromes are zero, the received codeword is valid (no errors).
// Otherwise, the syndrome values encode information about error positions and magnitudes.
func (ec *ErrorCorrector) computeSyndromes(received []gfpn.Element, numSyndromes int) []gfpn.Element {
	alpha := ec.field.Primitive()
	syndromes := make([]gfpn.Element, numSyndromes)

	for i := 0; i < numSyndromes; i++ {
		// Compute α^i
		alphaToI := ec.field.One()
		for j := 0; j < i; j++ {
			alphaToI = ec.field.Mul(alphaToI, alpha)
		}

		// Evaluate received polynomial at α^i using Horner's method
		// QR codes treat received[0] as the highest degree coefficient
		// r(α^i) = received[0]·α^i^(n-1) + received[1]·α^i^(n-2) + ... + received[n-1]
		syndrome := ec.field.Zero()
		for j := 0; j < len(received); j++ {
			syndrome = ec.field.Mul(syndrome, alphaToI)
			syndrome = ec.field.Add(syndrome, received[j])
		}

		syndromes[i] = syndrome
	}

	return syndromes
}

// reinterleaveBlocks combines corrected blocks back into a single data stream
//
// This reverses the de-interleaving process, but only returns the data codewords
// (not the error correction codewords, which are no longer needed).
func (ec *ErrorCorrector) reinterleaveBlocks(blocks [][]byte, ecBlocks *decoder.ECBlocks) []byte {
	// Calculate total data codewords
	totalDataCodewords := 0
	for _, block := range blocks {
		totalDataCodewords += len(block)
	}

	data := make([]byte, totalDataCodewords)
	dataIndex := 0

	// Re-interleave: alternate taking one byte from each block
	maxBlockSize := 0
	for _, block := range blocks {
		if len(block) > maxBlockSize {
			maxBlockSize = len(block)
		}
	}

	for i := 0; i < maxBlockSize; i++ {
		for j := 0; j < len(blocks); j++ {
			if i < len(blocks[j]) {
				data[dataIndex] = blocks[j][i]
				dataIndex++
			}
		}
	}

	return data
}

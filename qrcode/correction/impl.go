package correction

import (
	"fmt"

	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
)

// ApplyCorrections corrects errors in the received codeword
//
// Given the error positions and magnitudes computed by previous steps,
// this function applies corrections to produce the original codeword.
//
// For each error at position j_i with magnitude Y_i:
//
//	corrected[j_i] = received[j_i] - Y_i
//
// In fields of characteristic 2 (like GF(2^n)), subtraction equals addition,
// so this becomes: corrected[j_i] = received[j_i] + Y_i
//
// Parameters:
//   - field: The finite field GF(p^n)
//   - received: The received codeword (possibly containing errors)
//   - errorPositions: Positions where errors occurred (from Chien search)
//   - errorMagnitudes: Error values at each position (from Forney algorithm)
//
// Returns:
//   - The corrected codeword
//
// Example:
//
//	// Received codeword with error at position 0
//	received := []gfpn.Element{α³, 0, 0, 0, 0, 0, 0}
//	positions := []int{0}
//	magnitudes := []gfpn.Element{α³}
//
//	corrected := ApplyCorrections(field, received, positions, magnitudes)
//	// corrected = [0, 0, 0, 0, 0, 0, 0]
func ApplyCorrections(
	field gfpn.Field,
	received []gfpn.Element,
	errorPositions []int,
	errorMagnitudes []gfpn.Element,
) []gfpn.Element {
	if len(errorPositions) != len(errorMagnitudes) {
		panic(fmt.Sprintf("position count (%d) must match magnitude count (%d)",
			len(errorPositions), len(errorMagnitudes)))
	}

	// Create a copy of the received codeword
	corrected := make([]gfpn.Element, len(received))
	copy(corrected, received)

	// Apply each correction: corrected[j] = received[j] - Y_j
	for i, pos := range errorPositions {
		if pos < 0 || pos >= len(received) {
			panic(fmt.Sprintf("error position %d out of bounds [0, %d)", pos, len(received)))
		}

		// Subtract the error magnitude
		// In characteristic 2: subtraction = addition (since -a = a)
		corrected[pos] = field.Sub(corrected[pos], errorMagnitudes[i])
	}

	return corrected
}

// VerifyCorrection verifies that a codeword is valid by computing its syndromes
//
// A valid codeword has all syndromes equal to zero. This function computes
// the syndromes and checks if they're all zero.
//
// Parameters:
//   - field: The finite field GF(p^n)
//   - codeword: The codeword to verify
//   - numSyndromes: Number of syndromes to compute (typically 2t for t-error correction)
//
// Returns:
//   - syndromes: The computed syndrome values
//   - isValid: true if all syndromes are zero (valid codeword)
//
// Example:
//
//	syndromes, valid := VerifyCorrection(field, corrected, 4)
//	if !valid {
//	    fmt.Println("Correction failed - syndromes not zero:", syndromes)
//	}
func VerifyCorrection(
	field gfpn.Field,
	codeword []gfpn.Element,
	numSyndromes int,
) ([]gfpn.Element, bool) {
	// Compute syndromes directly
	alpha := field.Primitive() // primitive element
	syndromes := make([]gfpn.Element, numSyndromes)

	for i := 0; i < numSyndromes; i++ {
		// Compute α^i
		alphaToI := field.One()
		for j := 0; j < i; j++ {
			alphaToI = field.Mul(alphaToI, alpha)
		}

		// Evaluate codeword polynomial at α^i using Horner's method
		syndrome := field.Zero()
		for j := len(codeword) - 1; j >= 0; j-- {
			syndrome = field.Mul(syndrome, alphaToI)
			syndrome = field.Add(syndrome, codeword[j])
		}

		syndromes[i] = syndrome
	}

	// Check if all syndromes are zero
	isValid := true
	for _, s := range syndromes {
		if !s.IsZero() {
			isValid = false
			break
		}
	}

	return syndromes, isValid
}

// ExtractMessage extracts the message portion from a corrected codeword
//
// Reed-Solomon codes use systematic encoding where the message appears
// in a contiguous portion of the codeword. This function extracts it.
//
// For systematic encoding with parity at the beginning:
//
//	codeword = [parity_0, ..., parity_{2s-1}, msg_0, ..., msg_{k-1}]
//
// For systematic encoding with parity at the end:
//
//	codeword = [msg_0, ..., msg_{k-1}, parity_0, ..., parity_{2s-1}]
//
// Parameters:
//   - codeword: The corrected codeword
//   - messageLength: Length of the message (k)
//   - parityAtBeginning: true if parity symbols are at the start
//
// Returns:
//   - The extracted message
//
// Example:
//
//	// For RS(7,3) with 4 parity symbols at the beginning
//	message := ExtractMessage(corrected, 3, true)
//	// message = corrected[4:7]
func ExtractMessage(
	codeword []gfpn.Element,
	messageLength int,
	parityAtBeginning bool,
) []gfpn.Element {
	if messageLength <= 0 || messageLength > len(codeword) {
		panic(fmt.Sprintf("invalid message length %d for codeword length %d",
			messageLength, len(codeword)))
	}

	message := make([]gfpn.Element, messageLength)

	if parityAtBeginning {
		// Message is at the end: skip parity symbols
		parityLength := len(codeword) - messageLength
		copy(message, codeword[parityLength:])
	} else {
		// Message is at the beginning
		copy(message, codeword[:messageLength])
	}

	return message
}

// DecodeResult contains the result of Reed-Solomon decoding
type DecodeResult struct {
	Success           bool           // Whether decoding succeeded
	Message           []gfpn.Element // The decoded message (if successful)
	NumErrors         int            // Number of errors corrected
	ErrorPositions    []int          // Positions of errors
	ErrorMagnitudes   []gfpn.Element // Magnitudes of errors
	CorrectedCodeword []gfpn.Element // The corrected codeword
	Syndromes         []gfpn.Element // Final syndromes (should be all zero)
}

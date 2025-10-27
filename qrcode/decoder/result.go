package decoder

// DecodeResult contains the result of QR code decoding
//
// This structure provides detailed information about the decoding process,
// including the decoded message and statistics about error correction.
type DecodeResult struct {
	// Message is the decoded text from the QR code
	// For byte mode encoding, this is typically UTF-8 text
	Message string

	// CorrectionSuccessful indicates whether error correction succeeded
	// If false, the message may be incorrect or empty
	CorrectionSuccessful bool

	// NumErrorsCorrected is the total number of symbol errors that were corrected
	// across all Reed-Solomon blocks in the QR code
	NumErrorsCorrected int

	// ErrorPositions contains the positions of errors that were corrected
	// These are positions within the codeword blocks, useful for educational purposes
	ErrorPositions []int

	// BlockResults contains detailed results for each RS block
	// QR codes use multiple blocks for higher versions/error correction levels
	BlockResults []BlockResult
}

// BlockResult contains error correction details for a single Reed-Solomon block
type BlockResult struct {
	// BlockIndex identifies which block this result is for (0-based)
	BlockIndex int

	// NumDataCodewords is the number of data codewords in this block
	NumDataCodewords int

	// NumECCodewords is the number of error correction codewords in this block
	NumECCodewords int

	// ErrorsFound is the number of errors detected in this block
	ErrorsFound int

	// ErrorPositions are the positions of errors within this block
	ErrorPositions []int

	// CorrectionSucceeded indicates if correction worked for this block
	// Correction fails when errors exceed the correction capacity
	CorrectionSucceeded bool
}

package decoder

import (
	"fmt"

	"github.com/jalphad/abstract_algebra/qrcode/types"
)

// Decoder provides the complete QR code decoding pipeline
//
// This orchestrates the two main steps of QR code decoding:
//  1. Error Correction: Reed-Solomon error correction using GF(256) arithmetic
//  2. Data Decoding: Parsing the corrected bytes to extract the message
//
// Educational Purpose:
// This demonstrates how abstract algebra (finite field arithmetic) enables
// real-world applications like QR codes. The same Reed-Solomon algorithms
// we implemented for the exercises are used here to decode actual QR codes!
type Decoder struct {
	errorCorrector *ErrorCorrector
	dataDecoder    *DataDecoder
	verbose        bool // If true, print detailed decoding steps
}

// NewDecoder creates a new QR code decoder
//
// Returns an error if the GF(256) field cannot be initialized (should not happen
// with valid parameters).
func NewDecoder() (*Decoder, error) {
	errorCorrector, err := NewErrorCorrector()
	if err != nil {
		return nil, fmt.Errorf("failed to create error corrector: %w", err)
	}

	return &Decoder{
		errorCorrector: errorCorrector,
		dataDecoder:    NewDataDecoder(),
		verbose:        false,
	}, nil
}

// SetVerbose enables or disables verbose logging
//
// When enabled, the decoder will print detailed information about each step
// of the decoding process. Useful for educational purposes.
func (d *Decoder) SetVerbose(verbose bool) {
	d.verbose = verbose
}

// Decode performs the complete QR code decoding pipeline
//
// Steps:
//  1. Error Correction: Apply Reed-Solomon error correction to fix corrupted codewords
//  2. Data Decoding: Parse the corrected bytes to extract the message
//  3. Collect Statistics: Gather information about errors found and corrected
//
// This is the main entry point for decoding QR codes. It takes the raw extracted
// QR data and returns the decoded message along with detailed statistics.
//
// Parameters:
//   - qrData: Raw QR code data from the extractor (includes codewords, version, EC level, etc.)
//
// Returns:
//   - DecodeResult containing the message and error correction statistics
//   - Error if decoding fails (e.g., too many errors to correct)
//
// Example Usage:
//
//	decoder, err := NewDecoder()
//	if err != nil {
//	    return err
//	}
//	result, err := decoder.Decode(qrData)
//	if err != nil {
//	    return err
//	}
//	fmt.Println("Message:", result.Message)
//	fmt.Println("Errors corrected:", result.NumErrorsCorrected)
func (d *Decoder) Decode(qrData *types.QRCodeData) (*DecodeResult, error) {
	if d.verbose {
		fmt.Println("=== QR Code Decoding Pipeline ===")
		fmt.Printf("Version: %d\n", qrData.Version.GetVersionNumber())
		fmt.Printf("Error Correction Level: %s\n", qrData.ECLevel.String())
		fmt.Printf("Total Codewords: %d\n", len(qrData.RawCodewords))
	}

	// Step 1: Error Correction
	if d.verbose {
		fmt.Println("\n--- Step 1: Reed-Solomon Error Correction ---")
	}

	correctedData, blockResults, err := d.errorCorrector.CorrectCodewords(qrData)
	if err != nil {
		return nil, fmt.Errorf("error correction failed: %w", err)
	}

	// Collect error statistics
	totalErrors := 0
	allErrorPositions := []int{}
	allBlocksSucceeded := true

	for _, blockResult := range blockResults {
		if d.verbose {
			fmt.Printf("Block %d: %d errors found at positions %v\n",
				blockResult.BlockIndex, blockResult.ErrorsFound, blockResult.ErrorPositions)
		}
		totalErrors += blockResult.ErrorsFound
		allErrorPositions = append(allErrorPositions, blockResult.ErrorPositions...)
		if !blockResult.CorrectionSucceeded {
			allBlocksSucceeded = false
		}
	}

	if !allBlocksSucceeded {
		return &DecodeResult{
			Message:              "",
			CorrectionSuccessful: false,
			NumErrorsCorrected:   0,
			ErrorPositions:       allErrorPositions,
			BlockResults:         blockResults,
		}, fmt.Errorf("error correction failed for one or more blocks")
	}

	if d.verbose {
		fmt.Printf("Total errors corrected: %d\n", totalErrors)
		fmt.Printf("Corrected data: %d bytes\n", len(correctedData))
	}

	// Step 2: Data Decoding
	if d.verbose {
		fmt.Println("\n--- Step 2: Data Decoding ---")
	}

	message, err := d.dataDecoder.Decode(correctedData)
	if err != nil {
		return nil, fmt.Errorf("data decoding failed: %w", err)
	}

	if d.verbose {
		fmt.Printf("Decoded message: \"%s\"\n", message)
		fmt.Printf("Message length: %d characters\n", len(message))
	}

	// Build result
	result := &DecodeResult{
		Message:              message,
		CorrectionSuccessful: allBlocksSucceeded,
		NumErrorsCorrected:   totalErrors,
		ErrorPositions:       allErrorPositions,
		BlockResults:         blockResults,
	}

	if d.verbose {
		fmt.Println("\n=== Decoding Complete ===")
	}

	return result, nil
}

// DecodeWithStats is a convenience method that decodes and prints statistics
//
// This is useful for educational demonstrations where you want to show
// the decoding process and statistics in one call.
//
// Example:
//
//	decoder, _ := NewDecoder()
//	decoder.SetVerbose(true)
//	result, err := decoder.DecodeWithStats(qrData)
func (d *Decoder) DecodeWithStats(qrData *types.QRCodeData) (*DecodeResult, error) {
	result, err := d.Decode(qrData)
	if err != nil {
		return nil, err
	}

	// Print summary statistics
	fmt.Println("\n=== Decoding Summary ===")
	fmt.Printf("Message: \"%s\"\n", result.Message)
	fmt.Printf("Errors corrected: %d\n", result.NumErrorsCorrected)
	if result.NumErrorsCorrected > 0 {
		fmt.Printf("Error positions: %v\n", result.ErrorPositions)
	}
	fmt.Printf("Number of RS blocks: %d\n", len(result.BlockResults))
	for _, block := range result.BlockResults {
		fmt.Printf("  Block %d: %d data + %d EC codewords, %d errors corrected\n",
			block.BlockIndex, block.NumDataCodewords, block.NumECCodewords, block.ErrorsFound)
	}

	return result, nil
}

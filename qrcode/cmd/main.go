package main

import (
	"fmt"
	"os"

	"github.com/jalphad/abstract_algebra/qrcode/decoder"
	"github.com/jalphad/abstract_algebra/qrcode/types"
)

// QR Code Decoder with Reed-Solomon Error Correction
//
// This program is for educational purposes only!
//
// The program uses the gozxing library for image extraction
// and then uses its own custom implementations of the decoding and
// error correction algorithms built on top of a generic implementation
// of Galois Fields (GF(p^n)).
func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Parse command-line flags
	verbose := false
	imagePath := os.Args[1]

	if len(os.Args) >= 3 && os.Args[1] == "-v" {
		verbose = true
		if len(os.Args) < 3 {
			printUsage()
			return
		}
		imagePath = os.Args[2]
	}

	// Step 1: Extract QR code data from image
	fmt.Println("=== QR Code Extraction ===")
	extractor := types.NewQRExtractor()
	qrData, err := extractor.ExtractFromImage(imagePath)
	if err != nil {
		fmt.Printf("Error extracting QR code: %v\n", err)
		os.Exit(1)
	}

	// Print extraction info
	fmt.Printf("Version: %d\n", qrData.Version.GetVersionNumber())
	fmt.Printf("Error Correction Level: %s\n", qrData.ECLevel.String())
	fmt.Printf("Data Mask: %d\n", qrData.DataMask)
	fmt.Printf("Total Codewords: %d\n", len(qrData.RawCodewords))
	fmt.Printf("Data Codewords: %d\n", len(qrData.DataCodewords))
	fmt.Printf("EC Codewords: %d\n", len(qrData.ECCodewords))

	// Step 2: Decode with error correction
	fmt.Println("\n=== QR Code Decoding ===")
	dec, err := decoder.NewDecoder()
	if err != nil {
		fmt.Printf("Error creating decoder: %v\n", err)
		os.Exit(1)
	}

	dec.SetVerbose(verbose)

	result, err := dec.Decode(qrData)
	if err != nil {
		fmt.Printf("Error decoding QR code: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Display results
	fmt.Println("\n=== DECODING RESULTS ===")
	fmt.Printf("✓ Message: \"%s\"\n", result.Message)

	if result.NumErrorsCorrected > 0 {
		fmt.Printf("✓ Corrected %d error(s)\n", result.NumErrorsCorrected)
		if verbose {
			fmt.Printf("  Error positions: %v\n", result.ErrorPositions)
		}
	} else {
		fmt.Println("✓ No errors detected (clean QR code)")
	}

	// Display block statistics
	if verbose && len(result.BlockResults) > 0 {
		fmt.Println("\n=== Reed-Solomon Block Details ===")
		for _, block := range result.BlockResults {
			fmt.Printf("Block %d:\n", block.BlockIndex)
			fmt.Printf("  Data codewords: %d\n", block.NumDataCodewords)
			fmt.Printf("  EC codewords: %d\n", block.NumECCodewords)
			fmt.Printf("  Errors corrected: %d\n", block.ErrorsFound)
			if block.ErrorsFound > 0 {
				fmt.Printf("  Error positions: %v\n", block.ErrorPositions)
			}
		}
	}

	fmt.Println("\n=== DECODING COMPLETE ===")
	fmt.Println("The Reed-Solomon error correction algorithm successfully")
	fmt.Println("decoded the QR code using GF(256) finite field arithmetic!")
}

func printUsage() {
	fmt.Println("QR Code Decoder with Reed-Solomon Error Correction")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [-v] <qr_code_image>")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  qr_code_image    Path to QR code image (PNG, JPEG)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v               Verbose mode (show detailed decoding steps)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go qr_code.png")
	fmt.Println("  go run main.go -v my_qr_code.jpg")
}

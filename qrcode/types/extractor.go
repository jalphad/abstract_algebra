package types

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/makiuchi-d/gozxing/qrcode/decoder"
	"github.com/makiuchi-d/gozxing/qrcode/detector"
)

// NewQRExtractor creates a new QR code extractor
func NewQRExtractor() *QRExtractor {
	return &QRExtractor{
		reader: qrcode.NewQRCodeReader(),
	}
}

// QRExtractor handles the extraction of raw QR code data
type QRExtractor struct {
	reader gozxing.Reader
}

// ExtractFromImage loads an image file and extracts QR code data
func (qe *QRExtractor) ExtractFromImage(imagePath string) (*QRCodeData, error) {
	// Load image
	img, err := loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Convert to gozxing format
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, fmt.Errorf("failed to create binary bitmap: %w", err)
	}

	return qe.ExtractFromBitmap(bmp)
}

// ExtractFromBitmap extracts QR code data from a binary bitmap
func (qe *QRExtractor) ExtractFromBitmap(bmp *gozxing.BinaryBitmap) (*QRCodeData, error) {
	// Detect QR code and get detector result
	matrix, err := bmp.GetBlackMatrix()
	if err != nil {
		return nil, err
	}
	detect := detector.NewDetector(matrix)
	detectorResult, err := detect.Detect(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to detect QR code: %w", err)
	}

	// Extract the bit matrix (this is the sampled QR code grid)
	bitMatrix := detectorResult.GetBits()

	// Create a custom decoder to extract raw data
	qrData, err := qe.extractRawData(bitMatrix)
	if err != nil {
		return nil, fmt.Errorf("failed to extract raw data: %w", err)
	}

	return qrData, nil
}

// extractRawData extracts the raw codewords from the QR code bit matrix
func (qe *QRExtractor) extractRawData(bitMatrix *gozxing.BitMatrix) (*QRCodeData, error) {
	// Read format information (contains error correction level and mask pattern)
	formatInfo, err := qe.readFormatInformation(bitMatrix)
	if err != nil {
		return nil, fmt.Errorf("failed to read format information: %w", err)
	}

	// Determine version from the size of the matrix
	version, err := decoder.Version_GetProvisionalVersionForDimension(bitMatrix.GetHeight())
	if err != nil {
		return nil, fmt.Errorf("failed to determine version: %w", err)
	}

	// Remove the data mask
	dataMask := decoder.DataMaskValues[formatInfo.GetDataMask()]
	dataMask.UnmaskBitMatrix(bitMatrix, bitMatrix.GetHeight())

	// Read the raw codewords
	rawCodewords, err := qe.readCodewords(bitMatrix, version, formatInfo.GetErrorCorrectionLevel())
	if err != nil {
		return nil, fmt.Errorf("failed to read codewords: %w", err)
	}

	// Split into data and error correction codewords
	dataCodewords, ecCodewords := qe.splitCodewords(rawCodewords, version, formatInfo.GetErrorCorrectionLevel())

	return &QRCodeData{
		Version:       version,
		FormatInfo:    formatInfo,
		RawCodewords:  rawCodewords,
		DataCodewords: dataCodewords,
		ECCodewords:   ecCodewords,
		ECLevel:       formatInfo.GetErrorCorrectionLevel(),
		DataMask:      formatInfo.GetDataMask(),
		BitMatrix:     bitMatrix,
	}, nil
}

// readFormatInformation reads the format information from the QR code
func (qe *QRExtractor) readFormatInformation(bitMatrix *gozxing.BitMatrix) (*decoder.FormatInformation, error) {
	formatInfo1 := qe.readFormatInformationBits1(bitMatrix)
	if formatInfo1 != nil {
		return formatInfo1, nil
	}

	// If first attempt failed, try the backup location
	formatInfo2 := qe.readFormatInformationBits2(bitMatrix)
	if formatInfo2 != nil {
		return formatInfo2, nil
	}

	return nil, fmt.Errorf("failed to read format information")
}

// readFormatInformationBits1 reads format info from the primary location
func (qe *QRExtractor) readFormatInformationBits1(bitMatrix *gozxing.BitMatrix) *decoder.FormatInformation {
	formatInfoBits1 := 0

	// Read format info bits around the top-left finder pattern
	for i := 0; i < 6; i++ {
		formatInfoBits1 = qe.copyBit(bitMatrix, i, 8, formatInfoBits1)
	}
	formatInfoBits1 = qe.copyBit(bitMatrix, 7, 8, formatInfoBits1)
	formatInfoBits1 = qe.copyBit(bitMatrix, 8, 8, formatInfoBits1)
	formatInfoBits1 = qe.copyBit(bitMatrix, 8, 7, formatInfoBits1)

	for j := 5; j >= 0; j-- {
		formatInfoBits1 = qe.copyBit(bitMatrix, 8, j, formatInfoBits1)
	}

	// Try to decode the format information
	return decoder.FormatInformation_DecodeFormatInformation(uint(formatInfoBits1), uint(formatInfoBits1^0x5412))
}

// readFormatInformationBits2 reads format info from the backup location
func (qe *QRExtractor) readFormatInformationBits2(bitMatrix *gozxing.BitMatrix) *decoder.FormatInformation {
	dimension := bitMatrix.GetHeight()
	formatInfoBits2 := 0

	// Read format info bits from bottom-left and top-right
	jMin := dimension - 7
	for j := dimension - 1; j >= jMin; j-- {
		formatInfoBits2 = qe.copyBit(bitMatrix, 8, j, formatInfoBits2)
	}
	for i := dimension - 8; i < dimension; i++ {
		formatInfoBits2 = qe.copyBit(bitMatrix, i, 8, formatInfoBits2)
	}

	return decoder.FormatInformation_DecodeFormatInformation(uint(formatInfoBits2), uint(formatInfoBits2^0x5412))
}

// copyBit copies a bit from the matrix to the result integer
func (qe *QRExtractor) copyBit(bitMatrix *gozxing.BitMatrix, i, j, result int) int {
	bit := 0
	if bitMatrix.Get(i, j) {
		bit = 1
	}
	return (result << 1) | bit
}

// readCodewords reads all codewords from the QR code matrix
func (qe *QRExtractor) readCodewords(bitMatrix *gozxing.BitMatrix, version *decoder.Version, ecLevel decoder.ErrorCorrectionLevel) ([]byte, error) {
	// Calculate total number of codewords
	totalCodewords := version.GetTotalCodewords()
	codewords := make([]byte, totalCodewords)
	codewordIndex := 0
	currentByte := 0
	bitsRead := 0

	dimension := bitMatrix.GetHeight()

	// Read in the zig-zag pattern starting from bottom-right
	readingUp := true
	for col := dimension - 1; col > 0; col -= 2 {
		if col == 6 {
			col-- // Skip timing column
		}

		for counter := 0; counter < dimension; counter++ {
			var row int
			if readingUp {
				row = dimension - 1 - counter
			} else {
				row = counter
			}

			for colOffset := 0; colOffset < 2; colOffset++ {
				currentCol := col - colOffset
				if !qe.isFunctionModule(bitMatrix, row, currentCol, version) {
					bitsRead++
					currentByte <<= 1
					if bitMatrix.Get(currentCol, row) {
						currentByte |= 1
					}

					if bitsRead == 8 {
						codewords[codewordIndex] = byte(currentByte)
						codewordIndex++
						bitsRead = 0
						currentByte = 0

						if codewordIndex >= totalCodewords {
							return codewords, nil
						}
					}
				}
			}
		}
		readingUp = !readingUp
	}

	if codewordIndex != totalCodewords {
		return nil, fmt.Errorf("read %d codewords but expected %d", codewordIndex, totalCodewords)
	}

	return codewords, nil
}

// isFunctionModule checks if a module is a function pattern (finder, timing, etc.)
func (qe *QRExtractor) isFunctionModule(bitMatrix *gozxing.BitMatrix, row, col int, version *decoder.Version) bool {
	dimension := bitMatrix.GetHeight()

	// Finder patterns (top-left, top-right, bottom-left)
	if (row <= 8 && col <= 8) || // Top-left
		(row <= 8 && col >= dimension-8) || // Top-right
		(row >= dimension-8 && col <= 8) { // Bottom-left
		return true
	}

	// Timing patterns
	if (row == 6 && col >= 8 && col < dimension-8) ||
		(col == 6 && row >= 8 && row < dimension-8) {
		return true
	}

	// Dark module
	if row == ((4*version.GetVersionNumber())+9) && col == 8 {
		return true
	}

	// Version information (for versions 7 and above)
	if version.GetVersionNumber() >= 7 {
		if (row >= dimension-11 && row < dimension-8 && col >= 0 && col <= 5) ||
			(row >= 0 && row <= 5 && col >= dimension-11 && col < dimension-8) {
			return true
		}
	}

	return false
}

// splitCodewords splits raw codewords into data and error correction portions
func (qe *QRExtractor) splitCodewords(rawCodewords []byte, version *decoder.Version, ecLevel decoder.ErrorCorrectionLevel) ([]byte, []byte) {
	ecBlocks := version.GetECBlocksForLevel(ecLevel)
	totalDataCodewords := version.GetTotalCodewords() - ecBlocks.GetTotalECCodewords()

	dataCodewords := make([]byte, totalDataCodewords)
	ecCodewords := make([]byte, ecBlocks.GetTotalECCodewords())

	// For simplicity, we'll assume interleaving happens later
	// In practice, QR codes interleave data and EC codewords in a complex pattern
	copy(dataCodewords, rawCodewords[:totalDataCodewords])
	copy(ecCodewords, rawCodewords[totalDataCodewords:])

	return dataCodewords, ecCodewords
}

// loadImage loads an image from file
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Decode(file)
	case ".png":
		return png.Decode(file)
	default:
		// Try to decode as generic image
		img, _, err := image.Decode(file)
		return img, err
	}
}

// PrintQRData prints extracted QR code data for workshop purposes
func (qrData *QRCodeData) PrintQRData() {
	fmt.Printf("QR Code Analysis:\n")
	fmt.Printf("Version: %d\n", qrData.Version.GetVersionNumber())
	fmt.Printf("Error Correction Level: %v\n", qrData.ECLevel)
	fmt.Printf("Data Mask: %d\n", qrData.DataMask)
	fmt.Printf("Matrix Size: %dx%d\n", qrData.BitMatrix.GetWidth(), qrData.BitMatrix.GetHeight())
	fmt.Printf("Total Codewords: %d\n", len(qrData.RawCodewords))
	fmt.Printf("Data Codewords: %d\n", len(qrData.DataCodewords))
	fmt.Printf("EC Codewords: %d\n", len(qrData.ECCodewords))

	fmt.Printf("\nRaw Codewords (hex): ")
	for i, b := range qrData.RawCodewords {
		if i%16 == 0 {
			fmt.Printf("\n%04x: ", i)
		}
		fmt.Printf("%02x ", b)
	}
	fmt.Printf("\n")
}

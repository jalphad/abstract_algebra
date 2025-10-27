package types

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQRExtractor_ExtractFromImage(t *testing.T) {
	// Arrange
	testContent := "Hello, QR Code!"
	//testFilePath := filepath.Join(t.TempDir(), "test_qr.png")
	testFilePath := "test_qr.png"
	err := createTestQRCode(testFilePath, testContent)
	if err != nil {
		t.Fatalf("Failed to create test QR code: %v", err)
	}
	extractor := NewQRExtractor()

	// Act
	qrData, err := extractor.ExtractFromImage(testFilePath)
	require.NoError(t, err)

	// Assert
	require.NotNil(t, qrData)
	require.NotNil(t, qrData.Version)
	require.NotNil(t, qrData.BitMatrix)

	assert.NotNil(t, qrData.FormatInfo)
	assert.Equal(t, 1, qrData.Version.GetVersionNumber())

	assert.Equal(t, 26, qrData.Version.GetTotalCodewords())
	assert.Len(t, qrData.DataCodewords, 19)
	assert.Len(t, qrData.ECCodewords, 7)

	// Possible ECLevels: L,M,Q,H
	assert.Equal(t, "L", qrData.ECLevel.String())

	assert.Equal(t, uint8(7), qrData.DataMask)

	expectedDimension := 17 + 4*qrData.Version.GetVersionNumber()
	assert.True(t, qrData.BitMatrix.GetWidth() == expectedDimension)
	assert.True(t, qrData.BitMatrix.GetHeight() == expectedDimension)
}

func TestQRExtractor_ExtractFromImage_NonExistentFile(t *testing.T) {
	// Arrange
	extractor := NewQRExtractor()

	// Act
	_, err := extractor.ExtractFromImage("nonexistent.png")

	// Assert
	assert.Error(t, err)
}

func TestQRExtractor_ExtractFromImage_InvalidImage(t *testing.T) {
	// Arrange
	invalidFile := filepath.Join(t.TempDir(), "invalid.txt")
	err := os.WriteFile(invalidFile, []byte("This is not an image"), 0644)
	require.NoError(t, err)
	extractor := NewQRExtractor()

	// Act
	_, err = extractor.ExtractFromImage(invalidFile)

	// Assert
	require.Error(t, err)
}

func TestNewQRExtractor(t *testing.T) {
	// Act
	extractor := NewQRExtractor()

	// Assert
	require.NotNil(t, extractor)
	assert.NotNil(t, extractor.reader)
}

// createTestQRCode creates a simple QR code image for testing
func createTestQRCode(filename, content string) error {
	// Create QR code
	writer := qrcode.NewQRCodeWriter()
	bitMatrix, err := writer.Encode(content, gozxing.BarcodeFormat_QR_CODE, 256, 256, nil)
	if err != nil {
		return err
	}

	// Convert to image
	img := image.NewGray(image.Rect(0, 0, bitMatrix.GetWidth(), bitMatrix.GetHeight()))
	for y := 0; y < bitMatrix.GetHeight(); y++ {
		for x := 0; x < bitMatrix.GetWidth(); x++ {
			if bitMatrix.Get(x, y) {
				img.Set(x, y, color.Gray{0}) // Black
			} else {
				img.Set(x, y, color.Gray{255}) // White
			}
		}
	}

	// Save to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

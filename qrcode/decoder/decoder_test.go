package decoder

import (
	"image"
	"image/color"
	"testing"

	"github.com/jalphad/abstract_algebra/qrcode/types"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDecoder_NoErrors tests decoding a clean QR code without errors
func TestDecoder_NoErrors(t *testing.T) {
	// Create a test QR code
	testMessage := "Hello, World!"
	qrData := createTestQRCode(t, testMessage, gozxing.EncodeHintType_ERROR_CORRECTION, "L")

	// Create decoder
	decoder, err := NewDecoder()
	require.NoError(t, err)

	// Decode
	result, err := decoder.Decode(qrData)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify message
	assert.Equal(t, testMessage, result.Message)
	assert.True(t, result.CorrectionSuccessful)
	assert.Equal(t, 0, result.NumErrorsCorrected)
	assert.Empty(t, result.ErrorPositions)
}

// TestDecoder_NoErrors_LongerMessage tests with a longer message
func TestDecoder_NoErrors_LongerMessage(t *testing.T) {
	testMessage := "This is a longer test message for QR code decoding with Reed-Solomon error correction!"
	qrData := createTestQRCode(t, testMessage, gozxing.EncodeHintType_ERROR_CORRECTION, "M")

	decoder, err := NewDecoder()
	require.NoError(t, err)

	result, err := decoder.Decode(qrData)
	require.NoError(t, err)

	assert.Equal(t, testMessage, result.Message)
	assert.True(t, result.CorrectionSuccessful)
	assert.Equal(t, 0, result.NumErrorsCorrected)
}

// TestDecoder_SingleError tests correction of a single error
func TestDecoder_SingleError(t *testing.T) {
	testMessage := "Test123"
	qrData := createTestQRCode(t, testMessage, gozxing.EncodeHintType_ERROR_CORRECTION, "L")

	// Introduce a single error in the data
	originalByte := qrData.RawCodewords[5]
	qrData.RawCodewords[5] ^= 0xFF // Flip all bits

	decoder, err := NewDecoder()
	require.NoError(t, err)

	result, err := decoder.Decode(qrData)
	require.NoError(t, err)

	// Should successfully correct the error
	assert.Equal(t, testMessage, result.Message)
	assert.True(t, result.CorrectionSuccessful)
	assert.Greater(t, result.NumErrorsCorrected, 0)

	// Restore for good hygiene
	qrData.RawCodewords[5] = originalByte
}

// TestDecoder_MultipleErrors tests correction of multiple errors
func TestDecoder_MultipleErrors(t *testing.T) {
	testMessage := "Testing multiple error correction"
	qrData := createTestQRCode(t, testMessage, gozxing.EncodeHintType_ERROR_CORRECTION, "M")

	// Introduce 2 errors (within M level correction capacity)
	qrData.RawCodewords[3] ^= 0x0F
	qrData.RawCodewords[7] ^= 0xF0

	decoder, err := NewDecoder()
	require.NoError(t, err)

	result, err := decoder.Decode(qrData)
	require.NoError(t, err)

	// Should successfully correct both errors
	assert.Equal(t, testMessage, result.Message)
	assert.True(t, result.CorrectionSuccessful)
	assert.GreaterOrEqual(t, result.NumErrorsCorrected, 2)
}

// TestDecoder_DifferentECLevels tests different error correction levels
func TestDecoder_DifferentECLevels(t *testing.T) {
	testMessage := "EC Level Test"

	levels := []string{"L", "M", "Q", "H"}

	for _, level := range levels {
		t.Run("Level"+level, func(t *testing.T) {
			qrData := createTestQRCode(t, testMessage, gozxing.EncodeHintType_ERROR_CORRECTION, level)

			decoder, err := NewDecoder()
			require.NoError(t, err)

			result, err := decoder.Decode(qrData)
			require.NoError(t, err)

			assert.Equal(t, testMessage, result.Message)
			assert.True(t, result.CorrectionSuccessful)
		})
	}
}

// TestDecoder_Verbose tests verbose mode
func TestDecoder_Verbose(t *testing.T) {
	testMessage := "Verbose test"
	qrData := createTestQRCode(t, testMessage, gozxing.EncodeHintType_ERROR_CORRECTION, "L")

	decoder, err := NewDecoder()
	require.NoError(t, err)
	decoder.SetVerbose(true)

	result, err := decoder.Decode(qrData)
	require.NoError(t, err)
	assert.Equal(t, testMessage, result.Message)
}

// TestBitStream tests the bitStream implementation
func TestBitStream(t *testing.T) {
	// Test data: 0b10110011, 0b01010101
	data := []byte{0b10110011, 0b01010101}
	bs := newBitStream(data)

	// Read 4 bits: should be 1011 (11)
	val, err := bs.readBits(4)
	require.NoError(t, err)
	assert.Equal(t, 0b1011, val)

	// Read 4 bits: should be 0011 (3)
	val, err = bs.readBits(4)
	require.NoError(t, err)
	assert.Equal(t, 0b0011, val)

	// Read 4 bits: should be 0101 (5)
	val, err = bs.readBits(4)
	require.NoError(t, err)
	assert.Equal(t, 0b0101, val)

	// Read 4 bits: should be 0101 (5)
	val, err = bs.readBits(4)
	require.NoError(t, err)
	assert.Equal(t, 0b0101, val)

	// Try to read more - should error
	_, err = bs.readBits(1)
	assert.Error(t, err)
}

// TestBitStream_SingleBits tests reading individual bits
func TestBitStream_SingleBits(t *testing.T) {
	data := []byte{0b10101010}
	bs := newBitStream(data)

	expected := []int{1, 0, 1, 0, 1, 0, 1, 0}
	for i, exp := range expected {
		val, err := bs.readBits(1)
		require.NoError(t, err, "Failed at bit %d", i)
		assert.Equal(t, exp, val, "Bit %d mismatch", i)
	}
}

// TestBitStream_CrossByte tests reading across byte boundaries
func TestBitStream_CrossByte(t *testing.T) {
	data := []byte{0b11110000, 0b10101010}
	bs := newBitStream(data)

	// Read 12 bits: 1111 0000 1010
	val, err := bs.readBits(12)
	require.NoError(t, err)
	assert.Equal(t, 0b111100001010, val)
}

// TestDataDecoder_ByteMode tests byte mode decoding
func TestDataDecoder_ByteMode(t *testing.T) {
	dd := NewDataDecoder()

	// Construct test data for byte mode
	// Format: 0100 (mode) + 00000101 (count=5) + "Hello"
	// Bits: 0100 00000101 01001000 01100101 01101100 01101100 01101111
	data := []byte{
		0b01000000, // 0100 0000
		0b01010100, // 0101 0100
		0b10000110, // 1000 0110
		0b01010110, // 0101 0110
		0b11000110, // 1100 0110
		0b11000110, // 1100 0110
		0b11110000, // 1111 (only first 4 bits used for 'o')
	}

	message, err := dd.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, "Hello", message)
}

// TestDecoder_EmptyMessage tests decoding an empty message
func TestDecoder_EmptyMessage(t *testing.T) {
	dd := NewDataDecoder()

	// Mode 0100 + count 00000000 (0 bytes)
	data := []byte{0b01000000, 0b00000000}

	message, err := dd.Decode(data)
	require.NoError(t, err)
	assert.Equal(t, "", message)
}

// TestErrorCorrector_GF256Field tests that GF(256) field is created correctly
func TestErrorCorrector_GF256Field(t *testing.T) {
	ec, err := NewErrorCorrector()
	require.NoError(t, err)
	require.NotNil(t, ec)
	require.NotNil(t, ec.field)

	// Verify field order
	assert.Equal(t, 256, ec.field.Order())

	// Test some basic field operations
	zero := ec.field.Zero()
	one := ec.field.One()

	assert.True(t, zero.IsZero())
	assert.False(t, one.IsZero())

	// Test addition: 1 + 1 = 0 in characteristic 2
	sum := ec.field.Add(one, one)
	assert.True(t, sum.IsZero())
}

// createTestQRCode creates a QR code for testing
func createTestQRCode(t *testing.T, content string, hintType gozxing.EncodeHintType, level string) *types.QRCodeData {
	// Create QR code with specified error correction level
	hints := map[gozxing.EncodeHintType]interface{}{
		hintType: level,
	}

	writer := qrcode.NewQRCodeWriter()
	bitMatrix, err := writer.Encode(content, gozxing.BarcodeFormat_QR_CODE, 256, 256, hints)
	require.NoError(t, err)

	// Convert to image and extract data
	img := bitMatrixToImage(bitMatrix)
	extractor := types.NewQRExtractor()

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	require.NoError(t, err)

	qrData, err := extractor.ExtractFromBitmap(bmp)
	require.NoError(t, err)

	return qrData
}

// bitMatrixToImage converts a BitMatrix to an image
func bitMatrixToImage(matrix *gozxing.BitMatrix) image.Image {
	width := matrix.GetWidth()
	height := matrix.GetHeight()
	img := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if matrix.Get(x, y) {
				img.Set(x, y, color.Gray{0}) // Black
			} else {
				img.Set(x, y, color.Gray{255}) // White
			}
		}
	}

	return img
}

// BenchmarkDecode benchmarks the full decoding pipeline
func BenchmarkDecode(b *testing.B) {
	// Create a test QR code once
	testMessage := "Benchmark test message for QR decoding"
	hints := map[gozxing.EncodeHintType]interface{}{
		gozxing.EncodeHintType_ERROR_CORRECTION: "M",
	}

	writer := qrcode.NewQRCodeWriter()
	bitMatrix, err := writer.Encode(testMessage, gozxing.BarcodeFormat_QR_CODE, 256, 256, hints)
	if err != nil {
		b.Fatal(err)
	}

	img := bitMatrixToImage(bitMatrix)
	extractor := types.NewQRExtractor()

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		b.Fatal(err)
	}

	qrData, err := extractor.ExtractFromBitmap(bmp)
	if err != nil {
		b.Fatal(err)
	}

	// Create decoder
	decoder, err := NewDecoder()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// Benchmark decoding
	for i := 0; i < b.N; i++ {
		_, err := decoder.Decode(qrData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

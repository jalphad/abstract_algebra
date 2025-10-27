package decoder

import (
	"fmt"
)

// DataDecoder decodes QR code data bytes into a readable message
//
// QR codes support multiple encoding modes:
//   - Numeric:       Encodes numbers 0-9 efficiently (3.33 bits per digit)
//   - Alphanumeric:  Encodes A-Z, 0-9, and some punctuation (5.5 bits per char)
//   - Byte:          Encodes any 8-bit data (8 bits per byte) - most flexible
//   - Kanji:         Encodes Japanese Kanji characters (13 bits per char)
//
// For educational purposes and broad compatibility, we focus on Byte mode,
// which can represent any UTF-8 text.
type DataDecoder struct {
	// Add configuration fields if needed in the future
}

// NewDataDecoder creates a new data decoder
func NewDataDecoder() *DataDecoder {
	return &DataDecoder{}
}

// Decode decodes corrected data bytes into a message string
//
// QR code data format (bit-level):
//   [Mode indicator: 4 bits][Character count: 8-16 bits][Data bits][Terminator: 0-4 bits][Padding]
//
// Mode indicators:
//   - 0001: Numeric
//   - 0010: Alphanumeric
//   - 0100: Byte
//   - 1000: Kanji
//   - 0000: End of message (ECI mode or terminator)
//
// This implementation focuses on Byte mode (0100), which is the most common
// for UTF-8 text.
//
// Parameters:
//   - dataBytes: Error-corrected data codewords from error correction step
//
// Returns:
//   - Decoded message as UTF-8 string
//   - Error if decoding fails
func (dd *DataDecoder) Decode(dataBytes []byte) (string, error) {
	if len(dataBytes) == 0 {
		return "", fmt.Errorf("no data to decode")
	}

	// Create bit stream for reading bits
	bits := newBitStream(dataBytes)

	// Read mode indicator (4 bits)
	modeIndicator, err := bits.readBits(4)
	if err != nil {
		return "", fmt.Errorf("failed to read mode indicator: %w", err)
	}

	// Check mode
	switch modeIndicator {
	case 0b0100: // Byte mode
		return dd.decodeByteMode(bits)
	case 0b0001: // Numeric mode
		return "", fmt.Errorf("numeric mode not yet supported (educational focus is on byte mode)")
	case 0b0010: // Alphanumeric mode
		return "", fmt.Errorf("alphanumeric mode not yet supported (educational focus is on byte mode)")
	case 0b1000: // Kanji mode
		return "", fmt.Errorf("kanji mode not yet supported (educational focus is on byte mode)")
	case 0b0000: // Terminator or ECI
		return "", nil // Empty message
	default:
		return "", fmt.Errorf("unknown mode indicator: %04b", modeIndicator)
	}
}

// decodeByteMode decodes data in byte mode
//
// Byte mode format:
//   [Character count: 8 bits for version 1-9, 16 bits for version 10-40][Data bytes]
//
// For simplicity, we assume version 1-9 (8-bit character count).
// This covers most common QR codes.
//
// Example:
//   Data: 0100 00001111 01001000 01100101 01101100 01101100 01101111
//         ^^^^ ^^^^^^^^ ^^^ 8 bytes of "Hello" (15 chars shown above is just example)
//         mode count    data...
func (dd *DataDecoder) decodeByteMode(bits *bitStream) (string, error) {
	// Read character count (8 bits for version 1-9)
	// For version 10+, this would be 16 bits
	// TODO: Could take version as parameter to handle this correctly
	count, err := bits.readBits(8)
	if err != nil {
		return "", fmt.Errorf("failed to read character count: %w", err)
	}

	if count == 0 {
		return "", nil
	}

	// Read data bytes
	dataBytes := make([]byte, count)
	for i := 0; i < count; i++ {
		b, err := bits.readBits(8)
		if err != nil {
			return "", fmt.Errorf("failed to read data byte %d: %w", i, err)
		}
		dataBytes[i] = byte(b)
	}

	// Convert to UTF-8 string
	return string(dataBytes), nil
}

// bitStream provides bit-level reading of byte data
//
// QR code data is packed at the bit level, so we need to be able to read
// arbitrary numbers of bits (not just whole bytes).
//
// Example:
//   Bytes: [0b10110011, 0b01010101]
//   Reading 4 bits: 1011 (11)
//   Next 4 bits: 0011 (3)
//   Next 4 bits: 0101 (5)
//   Next 4 bits: 0101 (5)
type bitStream struct {
	bytes      []byte // underlying byte array
	byteOffset int    // current byte position
	bitOffset  int    // current bit position within byte (0-7)
}

// newBitStream creates a new bit stream from bytes
func newBitStream(bytes []byte) *bitStream {
	return &bitStream{
		bytes:      bytes,
		byteOffset: 0,
		bitOffset:  0,
	}
}

// readBits reads n bits from the stream
//
// Bits are read from most significant to least significant.
// For example, reading 4 bits from 0b10110011 gives 0b1011 (11).
//
// Parameters:
//   - n: number of bits to read (1-32)
//
// Returns:
//   - value: unsigned integer value of the bits
//   - error: if not enough bits available
func (bs *bitStream) readBits(n int) (int, error) {
	if n < 0 || n > 32 {
		return 0, fmt.Errorf("invalid bit count: %d (must be 1-32)", n)
	}

	result := 0

	for i := 0; i < n; i++ {
		// Check if we have more bits available
		if bs.byteOffset >= len(bs.bytes) {
			return 0, fmt.Errorf("not enough bits available (requested %d, read %d)", n, i)
		}

		// Read one bit from current position
		currentByte := bs.bytes[bs.byteOffset]
		bitValue := (currentByte >> (7 - bs.bitOffset)) & 1

		// Add bit to result
		result = (result << 1) | int(bitValue)

		// Advance position
		bs.bitOffset++
		if bs.bitOffset == 8 {
			bs.bitOffset = 0
			bs.byteOffset++
		}
	}

	return result, nil
}

// available returns the number of bits still available to read
func (bs *bitStream) available() int {
	if bs.byteOffset >= len(bs.bytes) {
		return 0
	}
	remainingBytes := len(bs.bytes) - bs.byteOffset - 1
	remainingBitsInCurrentByte := 8 - bs.bitOffset
	return remainingBytes*8 + remainingBitsInCurrentByte
}

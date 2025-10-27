package types

import (
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode/decoder"
)

// QRCodeData holds the extracted QR code information for error correction workshop
type QRCodeData struct {
	Version       *decoder.Version
	FormatInfo    *decoder.FormatInformation
	RawCodewords  []byte
	DataCodewords []byte
	ECCodewords   []byte
	ECLevel       decoder.ErrorCorrectionLevel
	DataMask      byte
	BitMatrix     *gozxing.BitMatrix
}

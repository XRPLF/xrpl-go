package binarycodec

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	bigdecimal "github.com/Peersyst/xrpl-go/pkg/big-decimal"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

const (
	// zeroQualityHex is the hex representation of the zero quality.
	zeroQualityHex = 0x5500000000000000
	// maxQualityPrecision is the canonical quality mantissa precision.
	maxQualityPrecision = 16
	// minQualityExponent is the minimum exponent for a canonical quality.
	minQualityExponent = -96
	// maxQualityExponent is the maximum exponent for a canonical quality.
	maxQualityExponent = 80
)

// qualityFormat matches an optionally signed decimal number with an optional
// exponent.
var qualityFormat = regexp.MustCompile(`^-?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$`)

// EncodeQuality encodes a quality amount to a hex string.
func EncodeQuality(quality string) (string, error) {
	if !qualityFormat.MatchString(quality) {
		// Report the underlying bigdecimal error when there is one, so
		// callers can keep matching it with errors.Is.
		if _, err := bigdecimal.NewBigDecimal(quality); err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidQuality, err)
		}
		return "", ErrInvalidQuality
	}

	// A validated input is zero when its mantissa digits are all zero,
	// whatever the sign and exponent say.
	mantissaPart, _, _ := strings.Cut(strings.TrimPrefix(strings.ToLower(quality), "-"), "e")
	if strings.Trim(strings.Trim(mantissaPart, "0"), ".") == "" {
		zeroAmount := make([]byte, 8)
		binary.BigEndian.PutUint64(zeroAmount, uint64(zeroQualityHex))
		return hexutil.EncodeToUpperHex(zeroAmount), nil
	}

	bigDecimal, err := bigdecimal.NewBigDecimal(quality)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidQuality, err)
	}

	if bigDecimal.Precision > maxQualityPrecision {
		return "", ErrInvalidQuality
	}

	// Normalize the quality to the 16-digit mantissa used by XRPL.
	padding := maxQualityPrecision - bigDecimal.Precision
	exp := bigDecimal.Scale - padding
	if exp < minQualityExponent || exp > maxQualityExponent {
		return "", ErrInvalidQuality
	}

	mantissaString := bigDecimal.UnscaledValue + strings.Repeat("0", padding)
	mantissa, err := strconv.ParseUint(mantissaString, 10, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidQuality, err)
	}

	serialized := make([]byte, 8)
	binary.BigEndian.PutUint64(serialized, mantissa)
	//nolint:gosec // G115: exp plus its bias is bounded to [4, 180], which fits in a byte
	serialized[0] = byte(exp + 100)
	return hexutil.EncodeToUpperHex(serialized), nil
}

// DecodeQuality decodes a quality amount from a hex string to a string.
func DecodeQuality(quality string) (string, error) {
	if quality == "" {
		return "", ErrInvalidQuality
	}

	decoded, err := hex.DecodeString(quality)
	if err != nil || len(decoded) < 8 {
		return "", ErrInvalidQuality
	}

	bytes := decoded[len(decoded)-8:]
	exp := int(bytes[0]) - 100
	mantissaBytes := append([]byte{0}, bytes[1:]...)
	mantissa := binary.BigEndian.Uint64(mantissaBytes)

	// Convert mantissa to string
	mantissaStr := strconv.FormatUint(mantissa, 10)

	// Add decimal point based on exponent
	if exp < 0 {
		// Need to add leading zeros
		if len(mantissaStr) <= -exp {
			zeros := strings.Repeat("0", -exp-len(mantissaStr))
			mantissaStr = "0." + zeros + mantissaStr
		} else {
			// Insert decimal point from right to left
			insertPos := len(mantissaStr) + exp
			mantissaStr = mantissaStr[:insertPos] + "." + mantissaStr[insertPos:]
		}
	} else if exp > 0 {
		// Add trailing zeros
		mantissaStr += strings.Repeat("0", exp)
	}

	// Trim trailing zeros after decimal point
	if strings.Contains(mantissaStr, ".") {
		mantissaStr = strings.TrimRight(mantissaStr, "0")
		mantissaStr = strings.TrimRight(mantissaStr, ".")
	}

	return mantissaStr, nil
}

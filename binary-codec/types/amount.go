package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	bigdecimal "github.com/Peersyst/xrpl-go/pkg/big-decimal"
)

const (
	// MinIOUExponent is the minimum allowed exponent for IOU amounts.
	MinIOUExponent = -96
	// MaxIOUExponent is the maximum allowed exponent for IOU amounts.
	MaxIOUExponent = 80
	// MaxIOUPrecision is the maximum precision (significant digits) for IOU amounts.
	MaxIOUPrecision = 16
	// MinIOUMantissa is the minimum mantissa value for IOU amounts.
	MinIOUMantissa = 1000000000000000
	// MaxIOUMantissa is the maximum mantissa value for IOU amounts.
	MaxIOUMantissa = 9999999999999999

	// NotXRPBitMask is the bit mask used to indicate non-XRP currency amounts.
	NotXRPBitMask = 0x80
	// PosSignBitMask is the bit mask used to indicate positive amounts.
	PosSignBitMask = 0x4000000000000000
	// ZeroCurrencyAmountHex is the hex representation of a zero currency amount.
	ZeroCurrencyAmountHex = 0x8000000000000000
	// NativeAmountByteLength is the byte length for native XRP amounts.
	NativeAmountByteLength = 8
	// CurrencyAmountByteLength is the byte length for currency amounts.
	CurrencyAmountByteLength = 48

	// MPTAmountByteLength is the byte length for MPToken amounts.
	MPTAmountByteLength = 33
	// MPTMarkerByte is the marker byte used to identify MPToken amounts.
	MPTMarkerByte = 0x60
	// MPTIssuanceIDByteLength is the byte length for MPToken issuance IDs.
	MPTIssuanceIDByteLength = 24
	// MPTValueByteLength is the byte length for MPToken values.
	MPTValueByteLength = 8
	// MPTValueWithHeaderLength is the byte length for MPToken values including header.
	MPTValueWithHeaderLength = 9
	// MPTSignBitMask is the bit mask for MPToken sign bits.
	MPTSignBitMask = 0x40
	// MPTHighBitMask is the high bit mask for MPToken amounts.
	MPTHighBitMask = 0x80
	// MPTAmountFlag is the flag used to identify MPToken amounts.
	MPTAmountFlag = 0x20

	// MinXRP is the minimum XRP amount in XRP units.
	MinXRP = 1e-6
	// MaxDrops is the maximum number of drops (100 billion XRP in drops).
	MaxDrops = 1e17 // 100 billion XRP in drops aka 10^17

	// IOUCodeRegex is the regular expression pattern for IOU currency codes.
	IOUCodeRegex = `[0-9A-Za-z?!@#$%^&*<>(){}\[\]|]{3}`
)

var (
	errInvalidXRPValue     = errors.New("invalid XRP value")
	errInvalidCurrencyCode = errors.New("invalid currency code")

	errInvalidMPTLength     = fmt.Errorf("MPT slice must be exactly %d bytes", MPTAmountByteLength)
	errInsufficientMPTBytes = fmt.Errorf("not enough bytes for MPT issuance ID, need %d bytes", MPTIssuanceIDByteLength)
	errInvalidIssuanceIDLen = fmt.Errorf("mpt_issuance_id must be exactly %d bytes", MPTIssuanceIDByteLength)

	zeroByteArray = make([]byte, 20)

	errAmountMissingValue            = errors.New("amount missing value field")
	errInvalidAmountValue            = errors.New("invalid amount value")
	errInvalidMPTIssuanceID          = errors.New("invalid mpt_issuance_id")
	errIssuedCurrencyMissingCurrency = errors.New("issued currency missing currency field")
	errIssuedCurrencyMissingIssuer   = errors.New("issued currency missing issuer field")
	errInvalidCurrencyFormat         = errors.New("invalid currency")
	errInvalidIssuerFormat           = errors.New("invalid issuer")
	errInvalidAmountType             = errors.New("invalid amount type")
	errFailedConvertStringToBigFloat = errors.New("failed to convert string to big.Float")
)

// InvalidAmountError is a custom error type for invalid amounts.
type InvalidAmountError struct {
	Amount string
}

// Error method for InvalidAmountError returns a formatted error string.
func (e *InvalidAmountError) Error() string {
	return fmt.Sprintf("value '%s' is an invalid amount", e.Amount)
}

// OutOfRangeError is a custom error type for out-of-range values.
type OutOfRangeError struct {
	Type string
}

// Error method for OutOfRangeError returns a formatted error string.
func (e *OutOfRangeError) Error() string {
	return fmt.Sprintf("%s is out of range", e.Type)
}

// InvalidCodeError is a custom error type for invalid currency codes.
type InvalidCodeError struct {
	Disallowed string
}

// Error method for InvalidCodeError returns a formatted error string.
func (e *InvalidCodeError) Error() string {
	return fmt.Sprintf("'%s' is/are disallowed or invalid", e.Disallowed)
}

// Amount is a struct that represents an XRPL Amount.
type Amount struct{}

// FromJSON serializes an issued currency amount to its bytes representation from JSON.
func (a *Amount) FromJSON(value any) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return serializeXrpAmount(v)
	case map[string]any:
		// Extract and normalize the "value" field
		rawVal, ok := v["value"]
		if !ok {
			return nil, errAmountMissingValue
		}
		val, err := valueToString(rawVal)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errInvalidAmountValue.Error(), err)
		}

		// If there's an mpt_issuance_id key → MPT currency
		if rawID, ok := v["mpt_issuance_id"]; ok {
			id, err := valueToString(rawID)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", errInvalidMPTIssuanceID.Error(), err)
			}
			return serializeMPTCurrencyAmount(val, id)
		}

		// Otherwise, assume issued‐currency → must have both currency & issuer
		rawCurr, ok := v["currency"]
		if !ok {
			return nil, errIssuedCurrencyMissingCurrency
		}
		rawIss, ok := v["issuer"]
		if !ok {
			return nil, errIssuedCurrencyMissingIssuer
		}
		curr, err := valueToString(rawCurr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errInvalidCurrencyFormat.Error(), err)
		}
		iss, err := valueToString(rawIss)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errInvalidIssuerFormat.Error(), err)
		}
		return serializeIssuedCurrencyAmount(val, curr, iss)

	default:
		return nil, errInvalidAmountType
	}
}

// ToJSON deserializes a binary-encoded Amount object from a BinaryParser into a JSON representation.
// The order of checks is important:
// 1. If bit 0x80 is set → IOU (48 bytes)
// 2. If bit 0x20 is set → MPT (33 bytes)
// 3. Otherwise → XRP (8 bytes)
func (a *Amount) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.Peek()
	if err != nil {
		return nil, err
	}
	var sign string
	if !isPositive(b) {
		sign = "-"
	}

	// Check IOU first (bit 0x80 set) - IOU amounts are 48 bytes
	if !isNative(b) {
		token, err := p.ReadBytes(48)
		if err != nil {
			return nil, err
		}
		return deserializeToken(token)
	}

	// Check MPT next (bit 0x20 set) - MPT amounts are 33 bytes
	if b&MPTAmountFlag != 0 {
		token, err := p.ReadBytes(MPTAmountByteLength)
		if err != nil {
			return nil, err
		}
		return deserializeMPTAmount(token)
	}

	// Otherwise it's native XRP (8 bytes)
	xrp, err := p.ReadBytes(8)
	if err != nil {
		return nil, err
	}
	xrpVal := binary.BigEndian.Uint64(xrp)
	xrpVal &= 0x3FFFFFFFFFFFFFFF
	return sign + strconv.FormatUint(xrpVal, 10), nil
}

func deserializeToken(data []byte) (map[string]any, error) {

	var value string
	var err error
	if bytes.Equal(data[0:8], []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		value = "0"
	} else {
		value, err = deserializeValue(data[:8])
		if err != nil {
			return nil, err
		}
	}
	issuer, err := deserializeIssuer(data[28:])
	if err != nil {
		return nil, err
	}
	curr, err := deserializeCurrencyCode(data[8:28])
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"value":    value,
		"currency": curr,
		"issuer":   issuer,
	}, nil
}

func deserializeValue(data []byte) (string, error) {
	sign := ""
	if !isPositive(data[0]) {
		sign = "-"
	}
	valueBytes := data[:8]
	b1 := valueBytes[0]
	b2 := valueBytes[1]
	e1 := int((b1 & 0x3F) << 2)
	e2 := int(b2 >> 6)
	exponent := e1 + e2 - 97
	sigFigs := append([]byte{0, (b2 & 0x3F)}, valueBytes[2:]...)
	sigFigsInt := binary.BigEndian.Uint64(sigFigs)
	d, err := bigdecimal.NewBigDecimal(sign + strconv.FormatUint(sigFigsInt, 10) + "e" + strconv.Itoa(exponent))
	if err != nil {
		return "", err
	}
	val := d.GetScaledValue()
	err = verifyIOUValue(val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func deserializeCurrencyCode(data []byte) (string, error) {
	// Check for special xrp case
	if bytes.Equal(data, zeroByteArray) {
		return "XRP", nil
	}

	if bytes.Equal(data[0:12], make([]byte, 12)) && bytes.Equal(data[12:15], []byte{0x58, 0x52, 0x50}) && bytes.Equal(data[15:20], make([]byte, 5)) { // XRP in bytes
		return "", errInvalidCurrencyCode
	}
	iso := strings.ToUpper(string(data[12:15]))
	ok, _ := regexp.MatchString(IOUCodeRegex, iso)

	if !ok {
		return strings.ToUpper(hex.EncodeToString(data)), nil
	}
	return iso, nil
}

func deserializeIssuer(data []byte) (string, error) {
	return addresscodec.Encode(data, []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
}

// deserializeMPTValue extracts and formats the value component from an MPT amount binary representation.
// It handles sign bit and converts the 64-bit mantissa (split into MSB and LSB) to a string representation.
func deserializeMPTValue(data []byte) (string, error) {
	if len(data) < MPTValueWithHeaderLength {
		return "", errInvalidMPTLength
	}

	sign := ""
	if !isPositive(data[0]) {
		sign = "-"
	}

	mant := data[1:MPTValueWithHeaderLength]
	msb := binary.BigEndian.Uint32(mant[0:4])
	lsb := binary.BigEndian.Uint32(mant[4:8])

	msbBig := new(big.Int).SetUint64(uint64(msb))
	lsbBig := new(big.Int).SetUint64(uint64(lsb))

	shifted := new(big.Int).Lsh(msbBig, 32)

	num := new(big.Int).Or(shifted, lsbBig)

	return sign + num.String(), nil
}

// deserializeMPTIssuanceID extracts the issuance ID from an MPT amount binary representation
// and converts it to a hexadecimal string.
func deserializeMPTIssuanceID(data []byte) (string, error) {
	if len(data) < MPTIssuanceIDByteLength {
		return "", errInsufficientMPTBytes
	}
	idBytes := data[:MPTIssuanceIDByteLength]
	return hex.EncodeToString(idBytes), nil
}

// deserializeMPTAmount deserializes a complete MPT amount binary representation into its
// value and issuance ID components and returns them as a map.
// MPT amounts must be exactly 33 bytes in length.
func deserializeMPTAmount(data []byte) (map[string]any, error) {
	if len(data) != MPTAmountByteLength {
		return nil, errInvalidMPTLength
	}
	val, err := deserializeMPTValue(data[:MPTValueWithHeaderLength])
	if err != nil {
		return nil, err
	}
	id, err := deserializeMPTIssuanceID(data[MPTValueWithHeaderLength:])
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"value":           val,
		"mpt_issuance_id": id,
	}, nil
}

// verifyXrpValue validates the format of an XRP amount value.
// XRP values should not contain a decimal point because they are represented as integers as drops.
func verifyXrpValue(value string) error {

	r := regexp.MustCompile(`\d+`) // regex to match only digits
	m := r.FindAllString(value, -1)

	if len(m) != 1 {
		return errInvalidXRPValue
	}

	decimal := new(big.Float)
	decimal, ok := decimal.SetString(value) // bigFloat for precision

	if !ok {
		return errFailedConvertStringToBigFloat
	}

	if decimal.Sign() == 0 {
		return nil
	}

	if decimal.Cmp(big.NewFloat(MinXRP)) == -1 || decimal.Cmp(big.NewFloat(MaxDrops)) == 1 {
		return &InvalidAmountError{value}
	}

	return nil
}

// verifyIOUValue validates the format of an issued currency amount value.
func verifyIOUValue(value string) error {
	bigDecimal, err := bigdecimal.NewBigDecimal(value)
	if err != nil {
		return err
	}

	if bigDecimal.UnscaledValue == "" {
		return nil
	}

	if bigDecimal.Precision > MaxIOUPrecision {
		return &OutOfRangeError{Type: "Precision"}
	}

	adjustedExp := bigDecimal.Scale + bigDecimal.Precision - 16

	if adjustedExp < MinIOUExponent {
		return &OutOfRangeError{Type: "Exponent"}
	}
	if adjustedExp > MaxIOUExponent {
		return &OutOfRangeError{Type: "Exponent"}
	}

	return nil
}

// verifyMPTValue validates the format of an MPT amount value.
// MPT values must be integers (no decimal point) and must not have the high bit set.
func verifyMPTValue(value string) error {
	if strings.Contains(value, ".") {
		return &InvalidAmountError{Amount: value}
	}

	bi := new(big.Int)
	if _, ok := bi.SetString(value, 10); !ok {
		return &InvalidAmountError{Amount: value}
	}

	if bi.Sign() < 0 {
		return &InvalidAmountError{Amount: value}
	}

	// reject any value ≥ 1<<63 so v.Uint64() can never overflow
	if bi.BitLen() > 63 {
		return &InvalidAmountError{Amount: value}
	}

	if bi.Sign() != 0 {
		mask := new(big.Int).SetUint64(ZeroCurrencyAmountHex)
		if new(big.Int).And(bi, mask).Sign() != 0 {
			return &InvalidAmountError{Amount: value}
		}
	}

	return nil
}

// serializeXrpAmount serializes an XRP amount value.
func serializeXrpAmount(value string) ([]byte, error) {

	if verifyXrpValue(value) != nil {
		return nil, verifyXrpValue(value)
	}

	val, err := strconv.ParseUint(value, 10, 64)

	if err != nil {
		return nil, err
	}

	valWithPosBit := val | PosSignBitMask
	valBytes := make([]byte, NativeAmountByteLength)

	binary.BigEndian.PutUint64(valBytes, uint64(valWithPosBit))

	return valBytes, nil
}

// XRPL definition of precision is number of significant digits:
// Tokens can represent a wide variety of assets, including those typically measured in very small or very large denominations.
// This format uses significant digits and a power-of-ten exponent in a similar way to scientific notation.
// The format supports positive and negative significant digits and exponents within the specified range.
// Unlike typical floating-point representations of non-whole numbers, this format uses integer math for all calculations,
// so it always maintains 15 decimal digits of precision. Multiplication and division have adjustments to compensate for
// over-rounding in the least significant digits.

// SerializeIssuedCurrencyValue serializes the value field of an issued currency amount to its bytes representation.
func SerializeIssuedCurrencyValue(value string) ([]byte, error) {

	if verifyIOUValue(value) != nil {
		return nil, verifyIOUValue(value)
	}

	bigDecimal, err := bigdecimal.NewBigDecimal(value)

	if err != nil {
		return nil, err
	}

	if bigDecimal.UnscaledValue == "" {
		zeroAmount := make([]byte, 8)
		binary.BigEndian.PutUint64(zeroAmount, uint64(ZeroCurrencyAmountHex))
		return zeroAmount, nil // if the value is zero, then return the zero currency amount hex
	}

	mantissa, err := strconv.ParseUint(bigDecimal.UnscaledValue, 10, 64) // convert the unscaled value to an unsigned integer

	if err != nil {
		return nil, err
	}

	exp := bigDecimal.Scale // get the scale

	for mantissa < MinIOUMantissa && exp > MinIOUExponent {
		mantissa *= 10
		exp--
	}

	for mantissa > MaxIOUMantissa {
		if exp >= MaxIOUExponent {
			return nil, &OutOfRangeError{Type: "Exponent"} // if the scale is less than -96 or greater than 80, return an error
		}
		mantissa /= 10
		exp++

		if exp < MinIOUExponent || mantissa < MinIOUMantissa {
			// round down to zero
			zeroAmount := make([]byte, 8)
			binary.BigEndian.PutUint64(zeroAmount, uint64(ZeroCurrencyAmountHex))
			return zeroAmount, nil
		}

		if exp > MaxIOUExponent || mantissa > MaxIOUMantissa {
			return nil, &OutOfRangeError{Type: "Exponent"} // if the scale is less than -96 or greater than 80, return an error
		}
	}

	// convert components to bytes

	serial := uint64(ZeroCurrencyAmountHex) // set first bit to 1 because it is not XRP
	if bigDecimal.Sign == 0 {
		serial |= PosSignBitMask // if the sign is positive, set the sign (second) bit to 1
	}
	// TODO: Check if this is still needed
	//nolint:gosec // G115: Potential hardcoded credentials (gosec)
	serial |= (uint64(exp+97) << 54) // if the exponent is positive, set the exponent bits to the exponent + 97
	serial |= uint64(mantissa)       // last 54 bits are mantissa

	serialReturn := make([]byte, 8)
	binary.BigEndian.PutUint64(serialReturn, serial)

	return serialReturn, nil
}

// serializeIssuedCurrencyCode serializes an issued currency code to its bytes representation.
// The currency code can be 3 allowed string characters, or 20 bytes of hex.
func serializeIssuedCurrencyCode(currency string) ([]byte, error) {

	currency = strings.TrimPrefix(currency, "0x")                                    // remove the 0x prefix if it exists
	if currency == "XRP" || currency == "0000000000000000000000005852500000000000" { // if the currency code is uppercase XRP, return an error
		return nil, &InvalidCodeError{Disallowed: "XRP uppercase"}
	}

	switch len(currency) {
	case 3: // if the currency code is 3 characters, it is standard
		return serializeIssuedCurrencyCodeChars(currency)
	case 40: // if the currency code is 40 characters, it is hex encoded
		return serializeIssuedCurrencyCodeHex(currency)
	}

	return nil, &InvalidCodeError{Disallowed: currency}

}

func serializeIssuedCurrencyCodeHex(currency string) ([]byte, error) {
	decodedHex, err := hex.DecodeString(currency)

	if err != nil {
		return nil, err
	}

	if bytes.HasPrefix(decodedHex, []byte{0x00}) {

		if bytes.Equal(decodedHex[12:15], []byte{0x00, 0x00, 0x00}) {
			return make([]byte, 20), nil
		}

		if containsInvalidIOUCodeCharactersHex(decodedHex[12:15]) {
			return nil, errInvalidCurrencyCode
		}
		return decodedHex, nil

	}
	return decodedHex, nil
}

func serializeIssuedCurrencyCodeChars(currency string) ([]byte, error) {

	r := regexp.MustCompile(IOUCodeRegex) // regex to check if the currency code is valid
	m := r.FindAllString(currency, -1)

	if len(m) != 1 {
		return nil, errInvalidCurrencyCode
	}

	currencyBytes := make([]byte, 20)
	copy(currencyBytes[12:], []byte(currency))
	return currencyBytes, nil
}

// SerializeIssuedCurrencyAmount serializes the currency field of an issued currency amount to its bytes representation
// from value, currency code, and issuer address in string form (e.g. "USD", "r123456789").
// The currency code can be 3 allowed string characters, or 20 bytes of hex in standard currency format (e.g. with "00" prefix)
// or non-standard currency format (e.g. without "00" prefix)
func serializeIssuedCurrencyAmount(value, currency, issuer string) ([]byte, error) {

	var valBytes []byte
	var err error
	if value == "0" {
		valBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(valBytes, uint64(ZeroCurrencyAmountHex))
	} else {
		valBytes, err = SerializeIssuedCurrencyValue(value) // serialize the value
	}

	if err != nil {
		return nil, err
	}
	currencyBytes, err := serializeIssuedCurrencyCode(currency) // serialize the currency code

	if err != nil {
		return nil, err
	}
	_, issuerBytes, err := addresscodec.DecodeClassicAddressToAccountID(issuer) // decode the issuer address
	if err != nil {
		return nil, err
	}

	// AccountIDs that appear as children of special fields (Amount issuer and PathSet account) are not length-prefixed.
	// So in Amount and PathSet fields, don't use the length indicator 0x14. This is in contrast to the AccountID fields where the length indicator prefix 0x14 is added.

	return append(append(valBytes, currencyBytes...), issuerBytes...), nil
}

// serializeMPTCurrencyValue serializes an MPT currency value to its binary representation.
// The value is split into high and low 32-bit parts and encoded as an 8-byte sequence.
func serializeMPTCurrencyValue(value string) ([]byte, error) {
	if err := verifyMPTValue(value); err != nil {
		return nil, err
	}

	v, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, &InvalidAmountError{Amount: value}
	}

	// verifyMPTValue ensures v ≤ 2^63-1, so v.Uint64() is safe
	buf := make([]byte, NativeAmountByteLength)
	binary.BigEndian.PutUint64(buf, v.Uint64())
	return buf, nil
}

// serializeMPTCurrencyIssuanceID converts a hexadecimal issuance ID string to its binary representation.
// The issuance ID must be exactly 24 bytes when decoded.
func serializeMPTCurrencyIssuanceID(issuanceHex string) ([]byte, error) {
	idBytes, err := hex.DecodeString(issuanceHex)
	if err != nil {
		return nil, err
	}
	if len(idBytes) != MPTIssuanceIDByteLength {
		return nil, errInvalidIssuanceIDLen
	}
	return idBytes, nil
}

// serializeMPTCurrencyAmount serializes a complete MPT amount by combining the value and issuance ID.
// It adds the MPT marker byte and arranges the components into a 33-byte sequence.
func serializeMPTCurrencyAmount(valueStr, issuanceHex string) ([]byte, error) {
	if err := verifyMPTValue(valueStr); err != nil {
		return nil, err
	}

	valBytes, err := serializeMPTCurrencyValue(valueStr)
	if err != nil {
		return nil, err
	}

	idBytes, err := serializeMPTCurrencyIssuanceID(issuanceHex)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, MPTAmountByteLength)
	buf[0] = MPTMarkerByte
	copy(buf[1:MPTValueWithHeaderLength], valBytes)
	copy(buf[MPTValueWithHeaderLength:], idBytes)
	return buf, nil
}

// Returns true if this amount is a "native" XRP amount - first bit in first byte set to 0 for native XRP
func isNative(value byte) bool {
	x := value&NotXRPBitMask == 0 // & bitwise operator returns 1 if both first bits are 1, otherwise 0
	return x
}

// Determines if this AmountType is positive - 2nd bit in 1st byte is set to 1 for positive amounts
func isPositive(value byte) bool {
	x := value&0x40 > 0
	return x
}

func containsInvalidIOUCodeCharactersHex(currency []byte) bool {

	r := regexp.MustCompile(IOUCodeRegex) // regex to check if the currency code is valid
	m := r.FindAll(currency, -1)

	return len(m) != 1
}

// valueToString converts various JSON‐style value types into their string form.
func valueToString(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case json.Number:
		return x.String(), nil
	case float64:
		if x == math.Trunc(x) {
			return strconv.FormatInt(int64(x), 10), nil
		}
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported type %T for amount value", x)
	}
}

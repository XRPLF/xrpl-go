package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

const (
	// MPTIssuanceIDBytesLength is the number of bytes for an MPT issuance ID.
	MPTIssuanceIDBytesLength = 24
)

var (
	// NoAccountBytes is the marker used to identify MPT issues in the binary format.
	// This is the special account ID "0000000000000000000000000000000000000001".
	NoAccountBytes = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
)

var (
	// ErrInvalidIssueObject is returned when the JSON object is not a valid Issue.
	// ErrInvalidIssueObject is returned when the JSON object is not a valid Issue.
	ErrInvalidIssueObject = errors.New("invalid issue object")
	// ErrInvalidCurrency is returned when the currency field is missing or invalid in the Issue JSON.
	ErrInvalidCurrency = errors.New("invalid currency")
	// ErrInvalidIssuer is returned when the issuer field is missing or invalid in the Issue JSON.
	ErrInvalidIssuer = errors.New("invalid issuer")
	// ErrMissingIssueLengthOption is returned when no length option is provided to Issue.ToJSON.
	ErrMissingIssueLengthOption = errors.New("missing length option for Issue.ToJSON")
	// XRPBytes is the serialized byte representation for native XRP (zero-value currency issuer).
	XRPBytes = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

// Issue represents an XRPL Issue, which is essentially an AccountID.
// It is used to identify the issuer of a currency in the XRPL.
// The FromJson method converts a classic address string to an AccountID byte slice.
// The ToJson method converts an AccountID byte slice back to a classic address string.
// This type is crucial for handling currency issuers in XRPL transactions and ledger entries.
type Issue struct {
	length int
}

// FromJSON parses a classic address string and returns the corresponding AccountID byte slice.
// It uses the addresscodec package to decode the classic address.
// If the input is not a valid classic address, it returns an error.
func (i *Issue) FromJSON(json any) ([]byte, error) {
	if !i.isIssueObject(json) {
		return nil, ErrInvalidIssueObject
	}

	mapObj, ok := json.(map[string]any)
	if !ok {
		return nil, ErrInvalidIssueObject
	}

	currency, ok := mapObj["currency"]
	if !ok {
		mptIssuanceID, ok := mapObj["mpt_issuance_id"].(string)
		if !ok {
			return nil, ErrInvalidCurrency
		}

		mptIssuanceIDBytes, err := hex.DecodeString(mptIssuanceID)
		if err != nil {
			return nil, err
		}

		i.length = MPTIssuanceIDBytesLength

		return mptIssuanceIDBytes, nil
	}

	currencyCodec := &Currency{}

	currencyBytes, err := currencyCodec.FromJSON(currency)
	if err != nil {
		return nil, err
	}

	if issuerString, okstring := mapObj["issuer"].(string); ok && okstring {
		_, issuerBytes, err := addresscodec.DecodeClassicAddressToAccountID(issuerString)
		if err != nil {
			return nil, err
		}

		return append(currencyBytes, issuerBytes...), nil
	}

	// For XRP or currency-only issues, append 20 bytes of zeros for the issuer
	// to ensure the full 40-byte Issue structure is maintained.
	return append(currencyBytes, XRPBytes...), nil
}

// ToJSON converts a binary Issue representation back to a JSON object.
// It self-determines the length by progressively reading and checking the data:
// - XRP: 20 bytes (currency only, all zeros)
// - IOU: 40 bytes (currency + issuer)
// - MPT: 44 bytes (issuer account + NO_ACCOUNT marker + sequence)
// The opts parameter is ignored as length is determined automatically.
func (i *Issue) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	// Step 1: Read first 20 bytes (currency for XRP/IOU, or issuer account for MPT)
	currencyOrAccount, err := p.ReadBytes(20)
	if err != nil {
		return nil, err
	}

	// Step 2: Check if it's XRP (all zeros)
	if bytes.Equal(currencyOrAccount, XRPBytes) {
		// Consume the next 20 bytes (issuer) which should also be all zeros for XRP
		if _, err := p.ReadBytes(20); err != nil {
			return nil, err
		}
		return map[string]any{
			"currency": "XRP",
		}, nil
	}

	// Step 3: Read next 20 bytes (issuer for IOU, or NO_ACCOUNT marker for MPT)
	issuerOrNoAccount, err := p.ReadBytes(20)
	if err != nil {
		return nil, err
	}

	// Step 4: Check if it's MPT (NO_ACCOUNT marker)
	if bytes.Equal(issuerOrNoAccount, NoAccountBytes) {
		// MPT case - read 4 more bytes for sequence (stored in little-endian)
		sequenceBytes, err := p.ReadBytes(4)
		if err != nil {
			return nil, err
		}

		// Convert sequence from little-endian to big-endian for mpt_issuance_id
		sequence := binary.LittleEndian.Uint32(sequenceBytes)
		seqBE := make([]byte, 4)
		binary.BigEndian.PutUint32(seqBE, sequence)

		// mpt_issuance_id = sequence (BE) + issuer account
		seqBE = append(seqBE, currencyOrAccount...)
		return map[string]any{
			"mpt_issuance_id": strings.ToUpper(hex.EncodeToString(seqBE)),
		}, nil
	}

	// Step 5: IOU case - decode currency and issuer
	// currencyOrAccount contains the currency bytes
	currencyStr := decodeCurrencyBytes(currencyOrAccount)

	// issuerOrNoAccount contains the issuer bytes
	address, err := addresscodec.Encode(issuerOrNoAccount, []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"currency": currencyStr,
		"issuer":   address,
	}, nil
}

// decodeCurrencyBytes decodes a 20-byte currency into its string representation.
func decodeCurrencyBytes(currencyBytes []byte) string {
	if bytes.Equal(currencyBytes, XRPBytes) {
		return "XRP"
	}

	// Check if bytes has exactly 3 non-zero bytes at positions 12-14 (standard currency code)
	nonZeroCount := 0
	var currencyStr string
	for i := 0; i < len(currencyBytes); i++ {
		if currencyBytes[i] != 0 {
			if i >= 12 && i <= 14 {
				nonZeroCount++
				currencyStr += string(currencyBytes[i])
			} else {
				nonZeroCount = 0
				break
			}
		}
	}

	if nonZeroCount == 3 {
		return currencyStr
	}

	// Return hex-encoded currency for non-standard codes
	return strings.ToUpper(hex.EncodeToString(currencyBytes))
}

func (i *Issue) isIssueObject(obj any) bool {
	mapObj, ok := obj.(map[string]any)
	if !ok {
		return false
	}

	nKeys := len(mapObj)

	_, okMptIssuanceID := mapObj["mpt_issuance_id"]
	if nKeys == 1 && okMptIssuanceID {
		return true
	}

	_, okCurrency := mapObj["currency"]
	if nKeys == 1 && okCurrency {
		return true
	}

	_, okIssuer := mapObj["issuer"]
	if nKeys == 2 && okCurrency && okIssuer {
		return true
	}

	return false
}

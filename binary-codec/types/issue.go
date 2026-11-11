package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"regexp"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

const (
	// MPTIssuanceIDBytesLength is the number of bytes for an MPT issuance ID.
	MPTIssuanceIDBytesLength = 24
)

// NoAccountBytes is the marker used to identify MPT issues in the binary format.
// This is the special account ID "0000000000000000000000000000000000000001".
var NoAccountBytes = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}

var (
	// standardCurrencyCodeRegex matches the three bytes of a standard currency code.
	standardCurrencyCodeRegex = regexp.MustCompile(`^` + IOUCodeRegex + `$`)
	// XRPBytes is the serialized byte representation for native XRP (zero-value currency issuer).
	XRPBytes = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

// Issue is the codec for the XRPL Issue type, which identifies an asset without an amount:
// XRP, a token (currency and issuer), or an MPT (issuance ID).
type Issue struct{}

// FromJSON converts an Issue JSON object to its binary representation:
// - XRP: {"currency": "XRP"} to 20 bytes (currency only, all zeros)
// - IOU: {"currency": ..., "issuer": ...} to 40 bytes (currency + issuer)
// - MPT: {"mpt_issuance_id": ...} to 44 bytes (issuer account + NO_ACCOUNT marker + sequence)
// It returns ErrInvalidIssueObject when the object has another shape, and ErrInvalidIssuer
// when the issuer does not agree with the currency.
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
			return nil, ErrInvalidIssueObject
		}

		mptIssuanceIDBytes, err := hex.DecodeString(mptIssuanceID)
		if err != nil {
			return nil, err
		}

		// mpt_issuance_id is 24 bytes: 4 bytes sequence (BE) + 20 bytes issuer account
		// Binary format for MPT Issue is 44 bytes: issuer account (20) + NO_ACCOUNT marker (20) + sequence (LE 4)
		if len(mptIssuanceIDBytes) != MPTIssuanceIDBytesLength {
			return nil, ErrInvalidIssueObject
		}

		// Extract sequence (first 4 bytes, big-endian) and issuer account (last 20 bytes)
		sequenceBE := mptIssuanceIDBytes[0:4]
		issuerAccount := mptIssuanceIDBytes[4:24]

		// Convert sequence from big-endian to little-endian for internal storage
		sequence := binary.BigEndian.Uint32(sequenceBE)
		sequenceLE := make([]byte, 4)
		binary.LittleEndian.PutUint32(sequenceLE, sequence)

		// Encode as: issuerAccount(20) + NO_ACCOUNT(20) + sequence(4, little-endian) = 44 bytes
		result := make([]byte, 0, 44)
		result = append(result, issuerAccount...)
		result = append(result, NoAccountBytes...)
		result = append(result, sequenceLE...)

		return result, nil
	}

	currencyCodec := &Currency{}

	currencyBytes, err := currencyCodec.FromJSON(currency)
	if err != nil {
		return nil, err
	}

	issuer := mapObj["issuer"]
	if bytes.Equal(currencyBytes, XRPBytes) {
		if issuer != nil {
			return nil, ErrInvalidIssuer
		}
		return currencyBytes, nil
	}

	// As in xrpld's issueFromJson, only XRP has no issuer. Without one, a token issue
	// would serialize to 20 bytes and the decoder would read the next 20 bytes as its issuer.
	issuerString, ok := issuer.(string)
	if !ok {
		return nil, ErrInvalidIssuer
	}

	_, issuerBytes, err := addresscodec.DecodeClassicAddressToAccountID(issuerString)
	if err != nil {
		return nil, err
	}

	// As in xrpld's issueFromJson, a token issuer is never ACCOUNT_ZERO or the NO_ACCOUNT
	// marker. The decoder would read the marker as the start of an MPT issue.
	if addresscodec.IsZeroAccountID(issuerBytes) || bytes.Equal(issuerBytes, NoAccountBytes) {
		return nil, ErrInvalidIssuer
	}
	return append(currencyBytes, issuerBytes...), nil
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
		sequenceBE := make([]byte, 4)
		binary.BigEndian.PutUint32(sequenceBE, sequence)

		// mpt_issuance_id = sequence (BE) + issuer account
		// currencyOrAccount contains the issuer account (first 20 bytes we read)
		mptIssuanceID := make([]byte, 0, MPTIssuanceIDBytesLength)
		mptIssuanceID = append(mptIssuanceID, sequenceBE...)
		mptIssuanceID = append(mptIssuanceID, currencyOrAccount...)

		return map[string]any{
			"mpt_issuance_id": hexutil.EncodeToUpperHex(mptIssuanceID),
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
// As in xrpld and xrpl.js, only a standard currency code other than XRP decodes to its
// three characters. Any other value decodes to hex, so that it encodes back to the same bytes.
func decodeCurrencyBytes(currencyBytes []byte) string {
	if bytes.Equal(currencyBytes, XRPBytes) {
		return "XRP"
	}

	code := currencyBytes[12:15]
	if bytes.Equal(currencyBytes[:12], XRPBytes[:12]) &&
		bytes.Equal(currencyBytes[15:], XRPBytes[15:]) &&
		standardCurrencyCodeRegex.Match(code) &&
		string(code) != "XRP" {
		return string(code)
	}

	// Return hex-encoded currency for non-standard codes
	return hexutil.EncodeToUpperHex(currencyBytes)
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

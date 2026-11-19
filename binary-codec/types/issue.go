package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"

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
	// ErrInvalidIssueObject is returned when the JSON object is not a valid Issue.
	ErrInvalidIssueObject = errors.New("invalid issue object")
	// ErrInvalidIssuer is returned when the issuer field is missing or invalid in the Issue JSON.
	ErrInvalidIssuer = errors.New("invalid issuer")
	// XRPBytes is the serialized byte representation for native XRP (zero-value currency issuer).
	XRPBytes = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	// isoXRPBytes is the reserved standard-currency layout of "XRP". rippled reads it as
	// an ordinary currency, so the codec encodes and decodes it as hexadecimal.
	isoXRPBytes = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x58, 0x52, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00}
)

// IsXRPCurrency reports whether currencyBytes is native XRP or the reserved
// standard-currency layout of "XRP". Neither is a valid issued currency.
func IsXRPCurrency(currencyBytes []byte) bool {
	return bytes.Equal(currencyBytes, XRPBytes) || bytes.Equal(currencyBytes, isoXRPBytes)
}

// Issue represents an XRPL Issue field: native XRP, an issued currency and its issuer,
// or an MPT issuance. FromJSON and ToJSON convert through the sealed issue type, so
// the codec never writes or accepts an Issue that rippled's STIssue reader rejects.
type Issue struct{}

// FromJSON converts an Issue JSON object to its binary form. It returns ErrInvalidIssuer
// for an issued currency without an issuer, or with an ACCOUNT_ZERO or ACCOUNT_ONE
// issuer, and ErrInvalidIssueObject for an XRP issue with an issuer or unknown keys.
func (i *Issue) FromJSON(json any) ([]byte, error) {
	obj, ok := json.(map[string]any)
	if !ok {
		return nil, ErrInvalidIssueObject
	}
	parsed, err := parseIssueJSON(obj)
	if err != nil {
		return nil, err
	}
	return parsed.appendIssue(nil), nil
}

// ToJSON converts a binary Issue back to a JSON object. It determines the length from
// the data, as rippled's STIssue reader does:
// - XRP: 20 bytes (currency only, all zeros)
// - IOU: 40 bytes (currency + issuer)
// - MPT: 44 bytes (issuer account + NO_ACCOUNT marker + sequence)
// The opts parameter is ignored.
func (i *Issue) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	parsed, err := readIssue(p)
	if err != nil {
		return nil, err
	}
	return parsed.issueJSON()
}

// issue is a parsed Issue. Only xrpIssue, iouIssue and mptIssue implement it, and their
// constructors apply rippled's STIssue reader rules, so every value encodes to bytes that
// rippled can read.
type issue interface {
	appendIssue(b []byte) []byte
	issueJSON() (map[string]any, error)
}

// iouParts is the currency and issuer of an issued currency, with the rules of rippled's
// STAmount reader. Amount uses it directly, so ACCOUNT_ONE stays a valid Amount issuer.
type iouParts struct{ currency, issuer [20]byte }

// iouIssue is an issued currency in the STIssue layout.
type iouIssue struct{ iouParts }

type xrpIssue struct{}

type mptIssue struct {
	id [MPTIssuanceIDBytesLength]byte
}

// newIOUParts rejects the native all-zero currency and an ACCOUNT_ZERO issuer, as
// rippled's STAmount reader and STIssue isConsistent do. The reserved standard layout of
// "XRP" passes, because rippled reads it as an ordinary currency.
func newIOUParts(currency, issuer []byte) (iouParts, error) {
	switch {
	case len(currency) != CurrencyCodeByteLength:
		return iouParts{}, errInvalidCurrencyCode
	case len(issuer) != addresscodec.AccountAddressLength:
		return iouParts{}, ErrInvalidIssuer
	case bytes.Equal(currency, XRPBytes):
		return iouParts{}, &InvalidCodeError{Disallowed: "XRP"}
	case bytes.Equal(issuer, XRPBytes):
		return iouParts{}, ErrInvalidIssuer
	}
	var parts iouParts
	copy(parts.currency[:], currency)
	copy(parts.issuer[:], issuer)
	return parts, nil
}

// newIOUIssue also rejects an ACCOUNT_ONE issuer, which STIssue reads as the MPT marker.
func newIOUIssue(currency, issuer []byte) (iouIssue, error) {
	parts, err := newIOUParts(currency, issuer)
	if err != nil {
		return iouIssue{}, err
	}
	if bytes.Equal(issuer, NoAccountBytes) {
		return iouIssue{}, ErrInvalidIssuer
	}
	return iouIssue{parts}, nil
}

// newMPTIssue rejects an issuance whose issuer is ACCOUNT_ZERO. STIssue reads 20 leading
// zero bytes as XRP, so that issuance cannot be an Issue.
func newMPTIssue(id []byte) (mptIssue, error) {
	if len(id) != MPTIssuanceIDBytesLength {
		return mptIssue{}, ErrInvalidIssueObject
	}
	if bytes.Equal(id[4:], XRPBytes) {
		return mptIssue{}, ErrInvalidIssuer
	}
	var m mptIssue
	copy(m.id[:], id)
	return m, nil
}

func (p iouParts) append(b []byte) []byte {
	return append(append(b, p.currency[:]...), p.issuer[:]...)
}

func (p iouParts) json() (map[string]any, error) {
	issuer, err := deserializeIssuer(p.issuer[:])
	if err != nil {
		return nil, err
	}
	return map[string]any{"currency": formatCurrencyCode(p.currency[:]), "issuer": issuer}, nil
}

func (xrpIssue) appendIssue(b []byte) []byte { return append(b, XRPBytes...) }

func (xrpIssue) issueJSON() (map[string]any, error) {
	return map[string]any{"currency": "XRP"}, nil
}

func (i iouIssue) appendIssue(b []byte) []byte { return i.append(b) }

func (i iouIssue) issueJSON() (map[string]any, error) { return i.json() }

// appendIssue writes the issuer account, the NO_ACCOUNT marker and the sequence. The
// issuance ID holds the sequence big-endian, and STIssue stores it little-endian.
func (m mptIssue) appendIssue(b []byte) []byte {
	b = append(append(b, m.id[4:]...), NoAccountBytes...)
	return binary.LittleEndian.AppendUint32(b, binary.BigEndian.Uint32(m.id[:4]))
}

func (m mptIssue) issueJSON() (map[string]any, error) {
	return map[string]any{"mpt_issuance_id": hexutil.EncodeToUpperHex(m.id[:])}, nil
}

// parseIssueJSON chooses the issue variant from the field values, not from the number of keys.
func parseIssueJSON(obj map[string]any) (issue, error) {
	if rawID, ok := obj["mpt_issuance_id"]; ok {
		id, isString := rawID.(string)
		if !isString || len(obj) != 1 {
			return nil, ErrInvalidIssueObject
		}
		idBytes, err := hex.DecodeString(id)
		if err != nil {
			return nil, err
		}
		return newMPTIssue(idBytes)
	}

	for key := range obj {
		if key != "currency" && key != "issuer" {
			return nil, ErrInvalidIssueObject
		}
	}
	rawCurrency, ok := obj["currency"]
	if !ok {
		return nil, ErrInvalidIssueObject
	}
	currency, err := (&Currency{}).FromJSON(rawCurrency)
	if err != nil {
		return nil, err
	}

	rawIssuer, hasIssuer := obj["issuer"]
	if bytes.Equal(currency, XRPBytes) {
		if hasIssuer {
			return nil, ErrInvalidIssueObject
		}
		return xrpIssue{}, nil
	}
	issuerText, ok := rawIssuer.(string)
	if !ok {
		return nil, ErrInvalidIssuer
	}
	issuer, err := serializeIssuedCurrencyIssuer(issuerText)
	if err != nil {
		return nil, err
	}
	return newIOUIssue(currency, issuer)
}

// readIssue reads an Issue as rippled's STIssue reader does.
func readIssue(p interfaces.BinaryParser) (issue, error) {
	first, err := p.ReadBytes(20)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(first, XRPBytes) {
		return xrpIssue{}, nil
	}
	second, err := p.ReadBytes(20)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(second, NoAccountBytes) {
		return newIOUIssue(first, second)
	}
	sequence, err := p.ReadBytes(4)
	if err != nil {
		return nil, err
	}
	id := binary.BigEndian.AppendUint32(make([]byte, 0, MPTIssuanceIDBytesLength), binary.LittleEndian.Uint32(sequence))
	return newMPTIssue(append(id, first...))
}

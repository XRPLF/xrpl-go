package types

import (
	"errors"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

const (
	typeAccount  = 0x01
	typeCurrency = 0x10
	typeIssuer   = 0x20

	pathsetEndByte    = 0x00
	pathSeparatorByte = 0xFF
)

// serializePathCurrency serializes a currency code for use in path steps.
// Unlike serializeIssuedCurrencyCode, this allows "XRP" which serializes to 20 zero bytes.
func serializePathCurrency(currency string) ([]byte, error) {
	if currency == "XRP" {
		return make([]byte, 20), nil
	}
	return serializeIssuedCurrencyCode(currency)
}

// PathSet type declaration
type PathSet struct{}

// ErrInvalidPathSet is an error that's thrown when an invalid path set is provided.
var ErrInvalidPathSet = errors.New("invalid type to construct PathSet from. Expected []any of []any")

// FromJSON attempts to serialize a path set from a JSON representation of a slice of paths to a byte array.
// It returns the byte array representation of the path set, or an error if the provided json does not represent a valid path set.
func (p PathSet) FromJSON(json any) ([]byte, error) {

	if _, ok := json.([]any)[0].([]any); !ok {
		return nil, ErrInvalidPathSet
	}

	if !isPathSet(json.([]any)) {
		return nil, ErrInvalidPathSet
	}

	return newPathSet(json.([]any)), nil
}

// ToJSON decodes a path set from a binary representation using a provided binary parser, then translates it to a JSON representation.
// It returns a slice representing the JSON format of the path set, or an error if the path set could not be decoded or if an invalid step is encountered.
func (p PathSet) ToJSON(parser interfaces.BinaryParser, _ ...int) (any, error) {
	var pathSet []any

	for parser.HasMore() {
		peek, err := parser.Peek()
		if err != nil {
			return nil, err
		}

		if peek == pathsetEndByte {
			_, err := parser.ReadByte()
			if err != nil {
				return nil, err
			}
			break
		}

		path, err := parsePath(parser)
		if err != nil {
			return nil, err
		}

		if len(path) > 0 {
			for i, step := range path {
				stepMap, ok := step.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("step is not of type map[string]any")
				}
				// Calculate type by combining flags
				stepType := 0
				if _, ok := stepMap["account"]; ok {
					stepType |= typeAccount
				}
				if _, ok := stepMap["currency"]; ok {
					stepType |= typeCurrency
				}
				if _, ok := stepMap["issuer"]; ok {
					stepType |= typeIssuer
				}
				stepMap["type"] = stepType
				stepMap["type_hex"] = fmt.Sprintf("%016X", stepType)
				path[i] = stepMap
			}
			pathSet = append(pathSet, path)
		}
	}

	return pathSet, nil
}

// isPathSet determines if an array represents a valid path set.
// It checks if the array is either empty or if its first element is a valid path step.
func isPathSet(v []any) bool {
	return len(v) == 0 || len(v[0].([]any)) == 0 || isPathStep(v[0].([]any)[0].(map[string]any))
}

// isPathStep determines if a map represents a valid path step.
// It checks if any of the keys "account", "currency" or "issuer" are present in the map.
func isPathStep(v map[string]any) bool {
	return v["account"] != nil || v["currency"] != nil || v["issuer"] != nil
}

// newPathStep creates a path step from a map representation.
// It generates a byte array representation of the path step, encoding account, currency, and issuer information as appropriate.
func newPathStep(v map[string]any) []byte {

	dataType := 0x00
	b := make([]byte, 0)

	if v["account"] != nil {
		_, account, _ := addresscodec.DecodeClassicAddressToAccountID(v["account"].(string))
		b = append(b, account...)
		dataType |= typeAccount
	}
	if v["currency"] != nil {
		currency, _ := serializePathCurrency(v["currency"].(string))
		b = append(b, currency...)
		dataType |= typeCurrency
	}
	if v["issuer"] != nil {
		_, issuer, _ := addresscodec.DecodeClassicAddressToAccountID(v["issuer"].(string))
		b = append(b, issuer...)
		dataType |= typeIssuer
	}

	return append([]byte{byte(dataType)}, b...)
}

// newPath constructs a path from a slice of path steps.
// It generates a byte array representation of the path, encoding each path step in turn.
func newPath(v []any) []byte {
	b := make([]byte, 0)

	for _, step := range v { // for each step in the path (slice of path steps)
		b = append(b, newPathStep(step.(map[string]any))...) // append the path step to the byte array
	}
	return b
}

// newPathSet constructs a path set from a slice of paths.
// It generates a byte array representation of the path set, encoding each path and adding path separators as appropriate.
func newPathSet(v []any) []byte {

	b := make([]byte, 0)

	for _, path := range v { // for each path in the path set (slice of paths)
		b = append(b, newPath(path.([]any))...) // append the path to the byte array
		b = append(b, pathSeparatorByte)        // between each path, append a path separator byte
	}

	b[len(b)-1] = pathsetEndByte // replace last path separator with path set end byte

	return b

}

// parsePathStep decodes a path step from a binary representation using a provided binary parser.
// It returns a map representing the path step, or an error if the path step could not be decoded.
func parsePathStep(parser interfaces.BinaryParser) (map[string]any, error) {
	dataType, err := parser.ReadByte()
	if err != nil {
		return nil, err
	}

	step := make(map[string]any)

	operations := []struct {
		typeKey byte
		key     string
	}{
		{typeAccount, "account"},
		{typeCurrency, "currency"},
		{typeIssuer, "issuer"},
	}

	for _, op := range operations {
		if dataType&op.typeKey != 0 {
			bytes, err := parser.ReadBytes(20) // AccountID or Currency size
			if err != nil {
				return nil, err
			}

			if op.typeKey == typeCurrency {
				value, err := deserializeCurrencyCode(bytes)
				if err != nil {
					return nil, err
				}
				step[op.key] = value
			} else {
				value, err := addresscodec.Encode(bytes, []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
				if err != nil {
					return nil, err
				}
				step[op.key] = value
			}
		}
	}

	return step, nil
}

// parsePath decodes a path from a binary representation using a provided binary parser.
// It returns a slice representing the path, or an error if the path could not be decoded.
func parsePath(parser interfaces.BinaryParser) ([]any, error) {
	var path []any

	for parser.HasMore() {
		peek, err := parser.Peek()
		if err != nil {
			return nil, err
		}

		if peek == pathsetEndByte {
			break
		}

		if peek == pathSeparatorByte {
			_, err := parser.ReadByte()
			if err != nil {
				return nil, err
			}
			break
		}

		step, err := parsePathStep(parser)
		if err != nil {
			return nil, err
		}
		path = append(path, step)
	}

	return path, nil
}

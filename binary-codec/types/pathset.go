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
var ErrInvalidPathSet = errors.New("invalid path set")

// FromJSON attempts to serialize a path set from a JSON representation of a slice of paths to a byte array.
// It returns the byte array representation of the path set, or an error if the provided json does not represent a valid path set.
func (p PathSet) FromJSON(json any) ([]byte, error) {
	rawPaths, ok := json.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: input is not a []any", ErrInvalidPathSet)
	}
	if len(rawPaths) == 0 {
		return nil, fmt.Errorf("%w: empty path set", ErrInvalidPathSet)
	}

	paths := make([][]map[string]any, 0, len(rawPaths))
	for _, rawPath := range rawPaths {
		rawSteps, ok := rawPath.([]any)
		if !ok {
			return nil, fmt.Errorf("%w: path is not a []any", ErrInvalidPathSet)
		}
		if len(rawSteps) == 0 {
			return nil, fmt.Errorf("%w: empty path", ErrInvalidPathSet)
		}

		steps := make([]map[string]any, 0, len(rawSteps))
		for _, rawStep := range rawSteps {
			step, ok := rawStep.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%w: path step is not a map[string]any", ErrInvalidPathSet)
			}

			if !isPathStep(step) {
				return nil, fmt.Errorf("%w: path step has no account/currency/issuer", ErrInvalidPathSet)
			}
			steps = append(steps, step)
		}
		paths = append(paths, steps)
	}

	return newPathSet(paths)
}

// isPathStep determines if a map represents a valid path step.
// It checks if any of the keys "account", "currency" or "issuer" are present in the map.
func isPathStep(v map[string]any) bool {
	return v["account"] != nil || v["currency"] != nil || v["issuer"] != nil
}

// newPathSet constructs a path set from a non-empty slice of paths.
// It generates a byte array representation of the path set, encoding each path and adding path separators as appropriate.
func newPathSet(v [][]map[string]any) ([]byte, error) {
	b := make([]byte, 0)

	for _, path := range v { // for each path in the path set (slice of paths)
		p, err := newPath(path)
		if err != nil {
			return nil, err
		}
		b = append(b, p...)              // append the path to the byte array
		b = append(b, pathSeparatorByte) // between each path, append a path separator byte
	}

	b[len(b)-1] = pathsetEndByte // replace last path separator with path set end byte

	return b, nil
}

// newPath constructs a path from a non-empty slice of path steps.
// It generates a byte array representation of the path, encoding each path step in turn.
func newPath(v []map[string]any) ([]byte, error) {
	b := make([]byte, 0)

	for _, step := range v { // for each step in the path (slice of path steps)
		s, err := newPathStep(step)
		if err != nil {
			return nil, err
		}
		b = append(b, s...) // append the path step to the byte array
	}
	return b, nil
}

// newPathStep creates a path step from a map representation.
// It generates a byte array representation of the path step, encoding account, currency, and issuer information as appropriate.
func newPathStep(v map[string]any) ([]byte, error) {
	dataType := 0x00
	b := make([]byte, 0)

	if v["account"] != nil {
		accStr, ok := v["account"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: account is not a string", ErrInvalidPathSet)
		}
		_, account, err := addresscodec.DecodeClassicAddressToAccountID(accStr)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid account path step: %w", ErrInvalidPathSet, err)
		}
		b = append(b, account...)
		dataType |= typeAccount
	}
	if v["currency"] != nil {
		curStr, ok := v["currency"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: currency is not a string", ErrInvalidPathSet)
		}
		currency, err := serializePathCurrency(curStr)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid currency path step: %w", ErrInvalidPathSet, err)
		}
		b = append(b, currency...)
		dataType |= typeCurrency
	}
	if v["issuer"] != nil {
		issStr, ok := v["issuer"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: issuer is not a string", ErrInvalidPathSet)
		}
		_, issuer, err := addresscodec.DecodeClassicAddressToAccountID(issStr)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid issuer path step: %w", ErrInvalidPathSet, err)
		}
		b = append(b, issuer...)
		dataType |= typeIssuer
	}

	return append([]byte{byte(dataType)}, b...), nil
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

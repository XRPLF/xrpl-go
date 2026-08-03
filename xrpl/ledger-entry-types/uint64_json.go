package ledger

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// parseUInt64JSON accepts the string representation used by canonical XRPL JSON
// and unsigned JSON integers.js UInt64 codec. maxDigits rejects overlong
// leading-zero strings that would otherwise fit in 64 bits.
func parseUInt64JSON(data []byte, base, maxDigits int, description string) (uint64, error) {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		if len(text) > maxDigits {
			return 0, fmt.Errorf("%s must contain at most %d digits", description, maxDigits)
		}

		value, err := strconv.ParseUint(text, base, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", description, err)
		}
		return value, nil
	}

	var value uint64
	if err := json.Unmarshal(data, &value); err != nil {
		return 0, fmt.Errorf("%s must be a quoted string or an unsigned JSON integer: %w", description, err)
	}
	return value, nil
}

// quotedUInt64 marshals a base-ten ledger UInt64 (rippled's kSmdBaseTen amount
// fields) as a quoted decimal string.
type quotedUInt64 uint64

func (u quotedUInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(u), 10))
}

func (u *quotedUInt64) UnmarshalJSON(data []byte) error {
	value, err := parseUInt64JSON(data, 10, 20, "UInt64 decimal value")
	if err != nil {
		return err
	}
	*u = quotedUInt64(value)
	return nil
}

// hexUInt64 marshals any other ledger UInt64 as a zero-padded hexadecimal string.
type hexUInt64 uint64

func (u hexUInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%016X", uint64(u)))
}

func (u *hexUInt64) UnmarshalJSON(data []byte) error {
	value, err := parseUInt64JSON(data, 16, 16, "UInt64 hexadecimal value")
	if err != nil {
		return err
	}
	*u = hexUInt64(value)
	return nil
}


package builder

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrivateKeyValidation(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "one", value: strings.Repeat("0", 63) + "1", valid: true},
		{name: "curve order minus one", value: "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140", valid: true},
		{name: "zero", value: strings.Repeat("0", privateKeyHexLength)},
		{name: "curve order", value: "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"},
		{name: "curve order plus one", value: "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364142"},
		{name: "invalid hex", value: strings.Repeat("Z", privateKeyHexLength)},
		{name: "wrong length", value: "01"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.valid, isValidPrivateKey(test.value))
		})
	}
}

func TestEncryptionKeyEquality(t *testing.T) {
	const (
		key      = "0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"
		otherKey = "02C6047F9441ED7D6D3045406E95C07CD85C778E4B8CEF3CA7ABAC09B95C709EE5"
	)
	tests := []struct {
		name  string
		left  string
		right string
		equal bool
	}{
		{name: "same decoded key", left: key, right: strings.ToLower(key), equal: true},
		{name: "different valid keys", left: key, right: otherKey},
		{name: "invalid right key", left: key, right: "invalid"},
		{name: "invalid left key", left: "invalid", right: key},
		{name: "same invalid text", left: "invalid", right: "invalid"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.equal, sameEncryptionKey(test.left, test.right))
		})
	}
}

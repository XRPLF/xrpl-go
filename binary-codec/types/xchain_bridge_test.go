package types

import (
	"encoding/hex"
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/stretchr/testify/require"
)

const (
	xChainLockingDoor    = "rGzx83BVoqTYbGn7tiVAnFw7cbxjin13jL"
	xChainLockingDoorHex = "14af80285f637ee4af3c20378f9dfb12511acb8d27"
	xChainIssuingDoor    = "r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV"
	xChainIssuingDoorHex = "14550fc62003e785dc231a1058a05e56e3f09cf4e6"
	xChainXRPIssueHex    = "0000000000000000000000000000000000000000"
	xChainUSDIssueHex    = "0000000000000000000000005553440000000000" + "af80285f637ee4af3c20378f9dfb12511acb8d27"
)

func xChainBridgeJSON(lockingIssue, issuingIssue any) map[string]any {
	return map[string]any{
		"LockingChainDoor":  xChainLockingDoor,
		"LockingChainIssue": lockingIssue,
		"IssuingChainDoor":  xChainIssuingDoor,
		"IssuingChainIssue": issuingIssue,
	}
}

func TestXChainBridge_RoundTrip(t *testing.T) {
	xrp := map[string]any{"currency": "XRP"}
	usd := map[string]any{"currency": "USD", "issuer": xChainLockingDoor}

	tests := []struct {
		name string
		json map[string]any
		hex  string
	}{
		{
			name: "XRP to XRP",
			json: xChainBridgeJSON(xrp, xrp),
			hex:  xChainLockingDoorHex + xChainXRPIssueHex + xChainIssuingDoorHex + xChainXRPIssueHex,
		},
		{
			name: "token to XRP",
			json: xChainBridgeJSON(usd, xrp),
			hex:  xChainLockingDoorHex + xChainUSDIssueHex + xChainIssuingDoorHex + xChainXRPIssueHex,
		},
		{
			name: "XRP to token",
			json: xChainBridgeJSON(xrp, usd),
			hex:  xChainLockingDoorHex + xChainXRPIssueHex + xChainIssuingDoorHex + xChainUSDIssueHex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := (&XChainBridge{}).FromJSON(tt.json)
			require.NoError(t, err)
			require.Equal(t, tt.hex, hex.EncodeToString(encoded))

			decoded, err := (&XChainBridge{}).ToJSON(serdes.NewBinaryParser(encoded, definitions.Get()))
			require.NoError(t, err)
			require.Equal(t, tt.json, decoded)
		})
	}
}

func TestXChainBridge_FromJSON_Errors(t *testing.T) {
	xrp := map[string]any{"currency": "XRP"}
	withExtraField := xChainBridgeJSON(xrp, xrp)
	withExtraField["Extra"] = "value"
	withoutIssuingDoor := xChainBridgeJSON(xrp, xrp)
	delete(withoutIssuingDoor, "IssuingChainDoor")
	withRenamedField := xChainBridgeJSON(xrp, xrp)
	delete(withRenamedField, "IssuingChainDoor")
	withRenamedField["Door"] = xChainIssuingDoor
	withDoorNotString := xChainBridgeJSON(xrp, xrp)
	withDoorNotString["LockingChainDoor"] = 1
	withInvalidDoor := xChainBridgeJSON(xrp, xrp)
	withInvalidDoor["IssuingChainDoor"] = "invalid"

	tests := []struct {
		name        string
		json        any
		expectedErr error
	}{
		{name: "not an object", json: "bridge", expectedErr: errNotValidJSON},
		{name: "extra field", json: withExtraField, expectedErr: errNotValidXChainBridge},
		{name: "missing field", json: withoutIssuingDoor, expectedErr: errNotValidXChainBridge},
		{name: "renamed field", json: withRenamedField, expectedErr: errNotValidXChainBridge},
		{name: "door is not a string", json: withDoorNotString, expectedErr: ErrAccountIDNotString},
		{name: "door is not an address", json: withInvalidDoor, expectedErr: addresscodec.ErrInvalidAddressFormat},
		{name: "issue is an address", json: xChainBridgeJSON(xChainLockingDoor, xrp), expectedErr: ErrInvalidIssueObject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := (&XChainBridge{}).FromJSON(tt.json)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, encoded)
		})
	}
}

func TestXChainBridge_ToJSON_Errors(t *testing.T) {
	valid := xChainLockingDoorHex + xChainXRPIssueHex + xChainIssuingDoorHex + xChainXRPIssueHex

	tests := []struct {
		name string
		hex  string
	}{
		{name: "empty", hex: ""},
		{name: "truncated issue", hex: valid[:len(valid)-2]},
		{name: "door without length prefix", hex: strings.TrimPrefix(valid, "14")},
		{name: "door with a wrong length", hex: "13" + strings.TrimPrefix(valid, "14")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := hex.DecodeString(tt.hex)
			require.NoError(t, err)

			decoded, err := (&XChainBridge{}).ToJSON(serdes.NewBinaryParser(input, definitions.Get()))
			require.ErrorIs(t, err, errNotValidXChainBridge)
			require.Nil(t, decoded)
		})
	}
}

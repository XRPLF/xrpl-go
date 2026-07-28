package binarycodec

import (
	"encoding/hex"
	"encoding/json"
	"maps"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type confidentialMPTWireFixtures struct {
	Source string                       `json:"source"`
	Cases  []confidentialMPTWireFixture `json:"cases"`
}

type confidentialMPTWireFixture struct {
	Name   string         `json:"name"`
	JSON   map[string]any `json:"json"`
	Binary string         `json:"binary"`
}

func loadConfidentialMPTWireFixtures(t *testing.T) confidentialMPTWireFixtures {
	t.Helper()

	data, err := os.ReadFile("testdata/fixtures/confidential-mpt-wire-fields.json")
	require.NoError(t, err)

	var fixtures confidentialMPTWireFixtures
	require.NoError(t, json.Unmarshal(data, &fixtures))
	require.NotEmpty(t, fixtures.Source)
	require.Len(t, fixtures.Cases, 3)

	return fixtures
}

func TestConfidentialMPTWireFieldGoldenFixtures(t *testing.T) {
	fixtures := loadConfidentialMPTWireFixtures(t)

	for _, fixture := range fixtures.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			expected := maps.Clone(fixture.JSON)
			encoded, err := Encode(maps.Clone(expected))
			require.NoError(t, err)
			require.Equal(t, fixture.Binary, encoded)

			decoded, err := Decode(fixture.Binary)
			require.NoError(t, err)
			require.Equal(t, expected, decoded)

			reencoded, err := Encode(decoded)
			require.NoError(t, err)
			require.Equal(t, fixture.Binary, reencoded)
		})
	}
}

func TestLiveRPCNumberTypesRoundTrip(t *testing.T) {
	input := map[string]any{
		"TransactionType":    "MPTokenIssuanceCreate",
		"Flags":              json.Number("160"),
		"Sequence":           json.Number("1"),
		"LastLedgerSequence": json.Number("10"),
		"MaximumAmount":      "1000",
	}

	encoded, err := Encode(maps.Clone(input))
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, "MPTokenIssuanceCreate", decoded["TransactionType"])
	require.Equal(t, uint32(160), decoded["Flags"])
	require.Equal(t, uint32(1), decoded["Sequence"])
	require.Equal(t, uint32(10), decoded["LastLedgerSequence"])
	require.Equal(t, "1000", decoded["MaximumAmount"])

	reencoded, err := Encode(decoded)
	require.NoError(t, err)
	require.Equal(t, encoded, reencoded)
}

func TestMPTBaseTenUInt64GoldenWireValues(t *testing.T) {
	tests := []struct {
		field  string
		header string
	}{
		{field: "MaximumAmount", header: "3018"},
		{field: "OutstandingAmount", header: "3019"},
		{field: "MPTAmount", header: "301A"},
		{field: "LockedAmount", header: "301D"},
		{field: "ConfidentialOutstandingAmount", header: "3020"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			expected := tt.header + "00000000000003E8"
			encoded, err := Encode(map[string]any{tt.field: "1000"})
			require.NoError(t, err)
			require.Equal(t, expected, encoded)

			decoded, err := Decode(expected)
			require.NoError(t, err)
			require.Equal(t, map[string]any{tt.field: "1000"}, decoded)

			reencoded, err := Encode(decoded)
			require.NoError(t, err)
			require.Equal(t, expected, reencoded)
		})
	}
}

func TestConfidentialMPTGoldenWireLayout(t *testing.T) {
	fixtures := loadConfidentialMPTWireFixtures(t)
	byName := make(map[string]confidentialMPTWireFixture, len(fixtures.Cases))
	for _, fixture := range fixtures.Cases {
		byName[fixture.Name] = fixture
	}

	convert, err := hex.DecodeString(byName["convert_blinding_factor"].Binary)
	require.NoError(t, err)
	require.Len(t, convert, 37)
	require.Equal(t, []byte{0x12, 0x00, 0x55}, convert[:3])
	require.Equal(t, []byte{0x50, 0x28}, convert[3:5])
	blindingFactor, err := hex.DecodeString(byName["convert_blinding_factor"].JSON["BlindingFactor"].(string))
	require.NoError(t, err)
	require.Equal(t, blindingFactor, convert[5:37])

	convertBack, err := hex.DecodeString(byName["convert_back_blinding_and_balance_commitment"].Binary)
	require.NoError(t, err)
	require.Len(t, convertBack, 73)
	require.Equal(t, []byte{0x12, 0x00, 0x57}, convertBack[:3])
	require.Equal(t, []byte{0x50, 0x28}, convertBack[3:5])
	require.Equal(t, blindingFactor, convertBack[5:37])
	require.Equal(t, []byte{0x70, 0x2E, 0x21}, convertBack[37:40])

	send, err := hex.DecodeString(byName["send_amount_and_balance_commitments"].Binary)
	require.NoError(t, err)
	require.Len(t, send, 75)
	require.Equal(t, []byte{0x12, 0x00, 0x58}, send[:3])
	require.Equal(t, []byte{0x70, 0x2D, 0x21}, send[3:6])
	require.Equal(t, []byte{0x70, 0x2E, 0x21}, send[39:42])
}

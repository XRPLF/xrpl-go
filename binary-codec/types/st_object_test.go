package types

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestStObject_FromJson(t *testing.T) {
	tt := []struct {
		name        string
		input       any
		output      []byte
		expectedErr error
	}{
		{
			name:        "fail - input is not a map",
			input:       1,
			output:      nil,
			expectedErr: errors.New("not a valid json"),
		},
		// {}
		{
			name: "fail - not found error",
			input: map[string]any{
				"IncorrectField": 89,
				"Flags":          525288,
				"OfferSequence":  1752791,
			},
			output:      nil,
			expectedErr: errors.New("FieldName IncorrectField not found"),
		},
		{
			name: "pass - convert valid Json",
			input: map[string]any{
				"Fee":           "10",
				"Flags":         uint32(524288),
				"OfferSequence": uint32(1752791),
				"TakerGets":     "150000000000",
			},
			output:      []byte{0x22, 0x0, 0x8, 0x0, 0x0, 0x20, 0x19, 0x0, 0x1a, 0xbe, 0xd7, 0x65, 0x40, 0x0, 0x0, 0x22, 0xec, 0xb2, 0x5c, 0x0, 0x68, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa},
			expectedErr: nil,
		},
		{
			name: "pass - convert valid STObject with variable length",
			input: map[string]any{
				"TransactionType":   "Payment",
				"TransactionResult": 0,
				"Fee":               "10",
				"Flags":             uint32(524288),
				"OfferSequence":     uint32(1752791),
				"TakerGets":         "150000000000",
			},
			output:      []byte{0x12, 0x0, 0x0, 0x22, 0x0, 0x8, 0x0, 0x0, 0x20, 0x19, 0x0, 0x1a, 0xbe, 0xd7, 0x65, 0x40, 0x0, 0x0, 0x22, 0xec, 0xb2, 0x5c, 0x0, 0x68, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x3, 0x10, 0x0},
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			serializer := serdes.NewBinarySerializer(serdes.NewFieldIDCodec(definitions.Get()))
			stObject := NewSTObject(serializer)

			got, err := stObject.FromJSON(tc.input)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.output, got)
			}
		})
	}
}

func TestCreateFieldInstanceMapFromJsonXAddressZeroTag(t *testing.T) {
	got, err := createFieldInstanceMapFromJson(map[string]any{
		"Destination": "XV5sbjUmgPpvXv4ixFWZ5ptAYZ6PD2m4Er6SnvjVLpMWPjR",
	})
	require.NoError(t, err)
	require.Equal(t, "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY", got[testutil.GetFieldInstance(t, "Destination")])
	require.Equal(t, uint32(0), got[testutil.GetFieldInstance(t, "DestinationTag")])
}

func TestCreateFieldInstanceMapFromJsonXAddressDuplicateZeroTag(t *testing.T) {
	testcases := []struct {
		name string
		tag  any
	}{
		{
			name: "int",
			tag:  0,
		},
		{
			name: "uint32",
			tag:  uint32(0),
		},
		{
			name: "float64",
			tag:  float64(0),
		},
		{
			name: "json number",
			tag:  json.Number("0"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := createFieldInstanceMapFromJson(map[string]any{
				"Destination":    "XV5sbjUmgPpvXv4ixFWZ5ptAYZ6PD2m4Er6SnvjVLpMWPjR",
				"DestinationTag": tc.tag,
			})
			require.ErrorIs(t, err, ErrDuplicateXAddressTag)
		})
	}
}

func TestCreateFieldInstanceMapFromJsonXAddressNonZeroTag(t *testing.T) {
	t.Run("Destination populates DestinationTag", func(t *testing.T) {
		got, err := createFieldInstanceMapFromJson(map[string]any{
			"Destination": "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGxLBw6rACm2heBxVn",
		})
		require.NoError(t, err)
		require.Equal(t, "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", got[testutil.GetFieldInstance(t, "Destination")])
		require.Equal(t, uint32(22), got[testutil.GetFieldInstance(t, "DestinationTag")])
	})

	t.Run("Account populates SourceTag", func(t *testing.T) {
		got, err := createFieldInstanceMapFromJson(map[string]any{
			"Account": "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGxLBw6rACm2heBxVn",
		})
		require.NoError(t, err)
		require.Equal(t, "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", got[testutil.GetFieldInstance(t, "Account")])
		require.Equal(t, uint32(22), got[testutil.GetFieldInstance(t, "SourceTag")])
	})
}

func TestCreateFieldInstanceMapFromJsonXAddressDuplicateNonZeroTag(t *testing.T) {
	testcases := []struct {
		name string
		tag  any
	}{
		{
			name: "matching tag",
			tag:  uint32(22),
		},
		{
			name: "mismatching tag",
			tag:  uint32(99),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := createFieldInstanceMapFromJson(map[string]any{
				"Destination":    "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGxLBw6rACm2heBxVn",
				"DestinationTag": tc.tag,
			})
			require.ErrorIs(t, err, ErrDuplicateXAddressTag)
		})
	}
}

func TestStObject_ToJson(t *testing.T) {
	defs := definitions.Get()

	testcases := []struct {
		name        string
		malleate    func(t *testing.T) interfaces.BinaryParser
		output      any
		expectedErr error
	}{
		{
			"fail - binary parser read field error",
			func(t *testing.T) interfaces.BinaryParser {
				parser := testutil.NewMockBinaryParser(gomock.NewController(t))
				parser.EXPECT().HasMore().Return(true)
				parser.EXPECT().ReadField().Return(nil, errors.New("read field error"))
				return parser
			},
			nil,
			errors.New("ReadField error: read field error"),
		},
		{
			"pass - convert valid STObject",
			func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0x22, 0x0, 0x8, 0x0, 0x0, 0x20, 0x19, 0x0, 0x1a, 0xbe, 0xd7, 0x65, 0x40, 0x0, 0x0, 0x22, 0xec, 0xb2, 0x5c, 0x0, 0x68, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa}, defs)
			},
			map[string]any{
				"Fee":           "10",
				"Flags":         uint32(524288),
				"OfferSequence": uint32(1752791),
				"TakerGets":     "150000000000",
			},
			nil,
		},
		{
			"pass - convert valid STObject with variable length",
			func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0x12, 0x0, 0x0, 0x22, 0x0, 0x8, 0x0, 0x0, 0x20, 0x19, 0x0, 0x1a, 0xbe, 0xd7, 0x65, 0x40, 0x0, 0x0, 0x22, 0xec, 0xb2, 0x5c, 0x0, 0x68, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x3, 0x10, 0x0}, defs)
			},
			map[string]any{
				"TransactionType":   "Payment",
				"TransactionResult": "tesSUCCESS",
				"Fee":               "10",
				"Flags":             uint32(524288),
				"OfferSequence":     uint32(1752791),
				"TakerGets":         "150000000000",
			},
			nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			parser := tc.malleate(t)
			stObject := NewSTObject(serdes.NewBinarySerializer(serdes.NewFieldIDCodec(definitions.Get())))
			got, err := stObject.ToJSON(parser)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.output, got)
			}
		})
	}
}

func TestGetSortedKeys(t *testing.T) {
	tt := []struct {
		name   string
		input  map[definitions.FieldInstance]any
		output []definitions.FieldInstance
	}{
		{
			name: "pass - get sorted keys",
			input: map[definitions.FieldInstance]any{
				testutil.GetFieldInstance(t, "TransactionType"):   1,
				testutil.GetFieldInstance(t, "TransactionResult"): 0,
				testutil.GetFieldInstance(t, "IndexNext"):         5100000,
				testutil.GetFieldInstance(t, "SourceTag"):         1232,
				testutil.GetFieldInstance(t, "LedgerEntryType"):   1,
			},
			output: []definitions.FieldInstance{
				testutil.GetFieldInstance(t, "LedgerEntryType"),
				testutil.GetFieldInstance(t, "TransactionType"),
				testutil.GetFieldInstance(t, "SourceTag"),
				testutil.GetFieldInstance(t, "IndexNext"),
				testutil.GetFieldInstance(t, "TransactionResult"),
			},
		},
		{
			name: "pass - get sorted keys",
			input: map[definitions.FieldInstance]any{
				testutil.GetFieldInstance(t, "Account"):      "rMBzp8CgpE441cp5PVyA9rpVV7oT8hP3ys",
				testutil.GetFieldInstance(t, "TransferRate"): 4234,
				testutil.GetFieldInstance(t, "Expiration"):   23,
			},
			output: []definitions.FieldInstance{
				testutil.GetFieldInstance(t, "Expiration"),
				testutil.GetFieldInstance(t, "TransferRate"),
				testutil.GetFieldInstance(t, "Account"),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.output, getSortedKeys(tc.input))
		})
	}
}

package types

import (
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
			input: map[string]interface{}{
				"IncorrectField": 89,
				"Flags":          525288,
				"OfferSequence":  1752791,
			},
			output:      nil,
			expectedErr: errors.New("FieldName IncorrectField not found"),
		},
		{
			name: "pass - convert valid Json",
			input: map[string]interface{}{
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
			input: map[string]interface{}{
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
			map[string]interface{}{
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
			map[string]interface{}{
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
		input  map[definitions.FieldInstance]interface{}
		output []definitions.FieldInstance
	}{
		{
			name: "pass - get sorted keys",
			input: map[definitions.FieldInstance]interface{}{
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
			input: map[definitions.FieldInstance]interface{}{
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

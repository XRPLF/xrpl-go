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

func TestSTArrayFromJson(t *testing.T) {
	tt := []struct {
		name           string
		input          any
		expectedOutput []byte
		expectedErr    error
	}{
		{
			name:           "fail - input is not a []any",
			input:          "not a []any",
			expectedErr:    ErrNotSTObjectInSTArray,
			expectedOutput: nil,
		},
		{
			name: "pass - nested stobject test",
			input: []any{
				map[string]any{
					"DeletedNode": map[string]any{
						"FinalFields": map[string]any{
							"ExchangeRate":      "4a0745621d069432",
							"Flags":             uint32(0),
							"RootIndex":         "036D7E923EF22B65E19D95A6365C3373E1E96586E27015074A0745621D069432",
							"TakerGetsCurrency": "0000000000000000000000000000000000000000",
							"TakerGetsIssuer":   "0000000000000000000000000000000000000000",
							"TakerPaysCurrency": "0000000000000000000000004254430000000000",
							"TakerPaysIssuer":   "06A148131B436B2561C85967685B098E050EED4E",
						},
						"LedgerEntryType": "DirectoryNode",
						"LedgerIndex":     "036D7E923EF22B65E19D95A6365C3373E1E96586E27015074A0745621D069432",
					},
				},
				map[string]any{
					"ModifiedNode": map[string]any{
						"FinalFields": map[string]any{
							"Flags":     uint32(0),
							"Owner":     "r68xfQkhFxZrbwo6RRKq728JF2fJYQRE1",
							"RootIndex": "5ED0913938CD6D43BD6450201737394A9991753C4581E5682D61F35048D8FBFC",
						},
						"LedgerEntryType": "DirectoryNode",
						"LedgerIndex":     "5ED0913938CD6D43BD6450201737394A9991753C4581E5682D61F35048D8FBFC",
					},
				},
			},
			expectedOutput: []byte{228, 17, 0, 100, 86, 3, 109, 126, 146, 62, 242, 43, 101, 225, 157, 149, 166, 54, 92, 51, 115, 225, 233, 101, 134, 226, 112, 21, 7, 74, 7, 69, 98, 29, 6, 148, 50, 231, 34, 0, 0, 0, 0, 54, 74, 7, 69, 98, 29, 6, 148, 50, 88, 3, 109, 126, 146, 62, 242, 43, 101, 225, 157, 149, 166, 54, 92, 51, 115, 225, 233, 101, 134, 226, 112, 21, 7, 74, 7, 69, 98, 29, 6, 148, 50, 1, 17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 66, 84, 67, 0, 0, 0, 0, 0, 2, 17, 6, 161, 72, 19, 27, 67, 107, 37, 97, 200, 89, 103, 104, 91, 9, 142, 5, 14, 237, 78, 3, 17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 225, 225, 229, 17, 0, 100, 86, 94, 208, 145, 57, 56, 205, 109, 67, 189, 100, 80, 32, 23, 55, 57, 74, 153, 145, 117, 60, 69, 129, 229, 104, 45, 97, 243, 80, 72, 216, 251, 252, 231, 34, 0, 0, 0, 0, 88, 94, 208, 145, 57, 56, 205, 109, 67, 189, 100, 80, 32, 23, 55, 57, 74, 153, 145, 117, 60, 69, 129, 229, 104, 45, 97, 243, 80, 72, 216, 251, 252, 130, 20, 7, 182, 254, 233, 139, 189, 210, 159, 146, 253, 250, 246, 16, 123, 199, 65, 196, 19, 14, 143, 225, 225, 241},
		},
		{
			name: "pass - large starray test",
			input: []any{
				map[string]any{
					"NFToken": map[string]any{
						"NFTokenID": "000913881D8CE533B1355B905B5C75F85B1F9A5149B94D7BABAF4475000011D9",
						"URI":       "68747470733A2F2F697066732E696F2F697066732F516D57356232715736744D3975615659376E555162774356334D63514B566E6D5659516A6A4368784E65656B476B",
					},
				},
				map[string]any{
					"Signer": map[string]any{
						"NFTokenID": "000913881D8CE533B1355B905B5C75F85B1F9A5149B94D7BAF373F1C00001382",
						"URI":       "68747470733A2F2F697066732E696F2F697066732F516D57356232715736744D3975615659376E555162774356334D63514B566E6D5659516A6A4368784E65656B476B",
					},
				},
			},
			expectedOutput: []byte{236, 90, 0, 9, 19, 136, 29, 140, 229, 51, 177, 53, 91, 144, 91, 92, 117, 248, 91, 31, 154, 81, 73, 185, 77, 123, 171, 175, 68, 117, 0, 0, 17, 217, 117, 67, 104, 116, 116, 112, 115, 58, 47, 47, 105, 112, 102, 115, 46, 105, 111, 47, 105, 112, 102, 115, 47, 81, 109, 87, 53, 98, 50, 113, 87, 54, 116, 77, 57, 117, 97, 86, 89, 55, 110, 85, 81, 98, 119, 67, 86, 51, 77, 99, 81, 75, 86, 110, 109, 86, 89, 81, 106, 106, 67, 104, 120, 78, 101, 101, 107, 71, 107, 225, 224, 16, 90, 0, 9, 19, 136, 29, 140, 229, 51, 177, 53, 91, 144, 91, 92, 117, 248, 91, 31, 154, 81, 73, 185, 77, 123, 175, 55, 63, 28, 0, 0, 19, 130, 117, 67, 104, 116, 116, 112, 115, 58, 47, 47, 105, 112, 102, 115, 46, 105, 111, 47, 105, 112, 102, 115, 47, 81, 109, 87, 53, 98, 50, 113, 87, 54, 116, 77, 57, 117, 97, 86, 89, 55, 110, 85, 81, 98, 119, 67, 86, 51, 77, 99, 81, 75, 86, 110, 109, 86, 89, 81, 106, 106, 67, 104, 120, 78, 101, 101, 107, 71, 107, 225, 241},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			st := STArray{}
			got, err := st.FromJSON(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, got)
			}
		})
	}
}

func TestSTArrayToJson(t *testing.T) {
	defs := definitions.Get()

	tt := []struct {
		name        string
		malleate    func(t *testing.T) interfaces.BinaryParser
		output      []any
		expectedErr error
	}{
		{
			name: "fail - binary parser read bytes out of bounds",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				parser := testutil.NewMockBinaryParser(gomock.NewController(t))
				parser.EXPECT().HasMore().Return(true)
				parser.EXPECT().ReadField().Return(nil, errors.New("read bytes error"))
				return parser
			},
			output:      nil,
			expectedErr: errors.New("read bytes error"),
		},
		{
			name: "pass - large starray",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{228, 17, 0, 100, 86, 3, 109, 126, 146, 62, 242, 43, 101, 225, 157, 149, 166, 54, 92, 51, 115, 225, 233, 101, 134, 226, 112, 21, 7, 74, 7, 69, 98, 29, 6, 148, 50, 231, 34, 0, 0, 0, 0, 54, 74, 7, 69, 98, 29, 6, 148, 50, 88, 3, 109, 126, 146, 62, 242, 43, 101, 225, 157, 149, 166, 54, 92, 51, 115, 225, 233, 101, 134, 226, 112, 21, 7, 74, 7, 69, 98, 29, 6, 148, 50, 1, 17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 66, 84, 67, 0, 0, 0, 0, 0, 2, 17, 6, 161, 72, 19, 27, 67, 107, 37, 97, 200, 89, 103, 104, 91, 9, 142, 5, 14, 237, 78, 3, 17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 225, 225, 229, 17, 0, 100, 86, 94, 208, 145, 57, 56, 205, 109, 67, 189, 100, 80, 32, 23, 55, 57, 74, 153, 145, 117, 60, 69, 129, 229, 104, 45, 97, 243, 80, 72, 216, 251, 252, 231, 34, 0, 0, 0, 0, 88, 94, 208, 145, 57, 56, 205, 109, 67, 189, 100, 80, 32, 23, 55, 57, 74, 153, 145, 117, 60, 69, 129, 229, 104, 45, 97, 243, 80, 72, 216, 251, 252, 130, 20, 7, 182, 254, 233, 139, 189, 210, 159, 146, 253, 250, 246, 16, 123, 199, 65, 196, 19, 14, 143, 225, 225, 241}, defs)
			},
			output: []any{
				map[string]any{
					"DeletedNode": map[string]any{
						"FinalFields": map[string]any{
							"ExchangeRate":      "4A0745621D069432",
							"Flags":             uint32(0),
							"RootIndex":         "036D7E923EF22B65E19D95A6365C3373E1E96586E27015074A0745621D069432",
							"TakerGetsCurrency": "0000000000000000000000000000000000000000",
							"TakerGetsIssuer":   "0000000000000000000000000000000000000000",
							"TakerPaysCurrency": "0000000000000000000000004254430000000000",
							"TakerPaysIssuer":   "06A148131B436B2561C85967685B098E050EED4E",
						},
						"LedgerEntryType": "DirectoryNode",
						"LedgerIndex":     "036D7E923EF22B65E19D95A6365C3373E1E96586E27015074A0745621D069432",
					},
				},
				map[string]any{
					"ModifiedNode": map[string]any{
						"FinalFields": map[string]any{
							"Flags":     uint32(0),
							"Owner":     "r68xfQkhFxZrbwo6RRKq728JF2fJYQRE1",
							"RootIndex": "5ED0913938CD6D43BD6450201737394A9991753C4581E5682D61F35048D8FBFC",
						},
						"LedgerEntryType": "DirectoryNode",
						"LedgerIndex":     "5ED0913938CD6D43BD6450201737394A9991753C4581E5682D61F35048D8FBFC",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "pass - simple starray",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{236, 90, 0, 9, 19, 136, 29, 140, 229, 51, 177, 53, 91, 144, 91, 92, 117, 248, 91, 31, 154, 81, 73, 185, 77, 123, 171, 175, 68, 117, 0, 0, 17, 217, 117, 67, 104, 116, 116, 112, 115, 58, 47, 47, 105, 112, 102, 115, 46, 105, 111, 47, 105, 112, 102, 115, 47, 81, 109, 87, 53, 98, 50, 113, 87, 54, 116, 77, 57, 117, 97, 86, 89, 55, 110, 85, 81, 98, 119, 67, 86, 51, 77, 99, 81, 75, 86, 110, 109, 86, 89, 81, 106, 106, 67, 104, 120, 78, 101, 101, 107, 71, 107, 225, 224, 16, 90, 0, 9, 19, 136, 29, 140, 229, 51, 177, 53, 91, 144, 91, 92, 117, 248, 91, 31, 154, 81, 73, 185, 77, 123, 175, 55, 63, 28, 0, 0, 19, 130, 117, 67, 104, 116, 116, 112, 115, 58, 47, 47, 105, 112, 102, 115, 46, 105, 111, 47, 105, 112, 102, 115, 47, 81, 109, 87, 53, 98, 50, 113, 87, 54, 116, 77, 57, 117, 97, 86, 89, 55, 110, 85, 81, 98, 119, 67, 86, 51, 77, 99, 81, 75, 86, 110, 109, 86, 89, 81, 106, 106, 67, 104, 120, 78, 101, 101, 107, 71, 107, 225, 241}, defs)
			},
			output: []any{
				map[string]any{
					"NFToken": map[string]any{
						"NFTokenID": "000913881D8CE533B1355B905B5C75F85B1F9A5149B94D7BABAF4475000011D9",
						"URI":       "68747470733A2F2F697066732E696F2F697066732F516D57356232715736744D3975615659376E555162774356334D63514B566E6D5659516A6A4368784E65656B476B",
					},
				},
				map[string]any{
					"Signer": map[string]any{
						"NFTokenID": "000913881D8CE533B1355B905B5C75F85B1F9A5149B94D7BAF373F1C00001382",
						"URI":       "68747470733A2F2F697066732E696F2F697066732F516D57356232715736744D3975615659376E555162774356334D63514B566E6D5659516A6A4368784E65656B476B",
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			st := STArray{}
			parser := tc.malleate(t)
			act, err := st.ToJSON(parser)
			if tc.expectedErr != nil {
				require.Error(t, err, tc.expectedErr.Error())
				require.Nil(t, act)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tc.output, act)
			}
		})
	}
}

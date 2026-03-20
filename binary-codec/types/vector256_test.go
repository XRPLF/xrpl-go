package types

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestVector256_FromJson(t *testing.T) {
	testcases := []struct {
		name   string
		input  any
		output []byte
		err    error
	}{
		{
			name:   "fail - invalid input type",
			input:  "invalid input",
			output: nil,
			err:    &ErrInvalidVector256Type{fmt.Sprintf("%T", "invalid input")},
		},
		{
			name:   "pass - valid vector256",
			input:  []string{"73734B611DDA23D3F5F62E20A173B78AB8406AC5015094DA53F53D39B9EDB06C", "73734B611DDA23D3F5F62E20A173B78AB8406AC5015094DA53F53D39B9EDB06C"},
			output: []byte{0x73, 0x73, 0x4B, 0x61, 0x1D, 0xDA, 0x23, 0xD3, 0xF5, 0xF6, 0x2E, 0x20, 0xA1, 0x73, 0xB7, 0x8A, 0xB8, 0x40, 0x6A, 0xC5, 0x01, 0x50, 0x94, 0xDA, 0x53, 0xF5, 0x3D, 0x39, 0xB9, 0xED, 0xB0, 0x6C, 0x73, 0x73, 0x4B, 0x61, 0x1D, 0xDA, 0x23, 0xD3, 0xF5, 0xF6, 0x2E, 0x20, 0xA1, 0x73, 0xB7, 0x8A, 0xB8, 0x40, 0x6A, 0xC5, 0x01, 0x50, 0x94, 0xDA, 0x53, 0xF5, 0x3D, 0x39, 0xB9, 0xED, 0xB0, 0x6C},
			err:    nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := (&Vector256{}).FromJSON(tc.input)
			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.output, got)
			}
		})
	}
}

func TestVector256_ToJson(t *testing.T) {
	defs := definitions.Get()

	testcases := []struct {
		name     string
		malleate func(t *testing.T) (interfaces.BinaryParser, []int)
		output   any
		err      error
	}{
		{
			name: "fail - binary parser read bytes",
			malleate: func(t *testing.T) (interfaces.BinaryParser, []int) {
				parser := testutil.NewMockBinaryParser(gomock.NewController(t))
				parser.EXPECT().ReadBytes(100).AnyTimes().Return(nil, errors.New("read bytes error"))
				return parser, []int{100}
			},
			err: errors.New("read bytes error"),
		},
		{
			name: "pass - valid vector256",
			malleate: func(t *testing.T) (interfaces.BinaryParser, []int) {
				return serdes.NewBinaryParser([]byte{0x73, 0x73, 0x4B, 0x61, 0x1D, 0xDA, 0x23, 0xD3, 0xF5, 0xF6, 0x2E, 0x20, 0xA1, 0x73, 0xB7, 0x8A, 0xB8, 0x40, 0x6A, 0xC5, 0x01, 0x50, 0x94, 0xDA, 0x53, 0xF5, 0x3D, 0x39, 0xB9, 0xED, 0xB0, 0x6C, 0x73, 0x73, 0x4B}, defs), []int{32}
			},
			output: []string{"73734B611DDA23D3F5F62E20A173B78AB8406AC5015094DA53F53D39B9EDB06C"},
			err:    nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			parser, opts := tc.malleate(t)
			got, err := (&Vector256{}).ToJSON(parser, opts...)
			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.output, got)
			}
		})
	}
}

func TestVector256FromValue(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []byte
		expErr   error
	}{
		{
			name:     "test vector256 from value",
			input:    []string{"73734B611DDA23D3F5F62E20A173B78AB8406AC5015094DA53F53D39B9EDB06C", "73734B611DDA23D3F5F62E20A173B78AB8406AC5015094DA53F53D39B9EDB06C"},
			expected: []byte{0x73, 0x73, 0x4B, 0x61, 0x1D, 0xDA, 0x23, 0xD3, 0xF5, 0xF6, 0x2E, 0x20, 0xA1, 0x73, 0xB7, 0x8A, 0xB8, 0x40, 0x6A, 0xC5, 0x01, 0x50, 0x94, 0xDA, 0x53, 0xF5, 0x3D, 0x39, 0xB9, 0xED, 0xB0, 0x6C, 0x73, 0x73, 0x4B, 0x61, 0x1D, 0xDA, 0x23, 0xD3, 0xF5, 0xF6, 0x2E, 0x20, 0xA1, 0x73, 0xB7, 0x8A, 0xB8, 0x40, 0x6A, 0xC5, 0x01, 0x50, 0x94, 0xDA, 0x53, 0xF5, 0x3D, 0x39, 0xB9, 0xED, 0xB0, 0x6C},
			expErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := vector256FromValue(tc.input)

			if tc.expErr != nil {
				require.EqualError(t, tc.expErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, got)
			}
		})
	}
}

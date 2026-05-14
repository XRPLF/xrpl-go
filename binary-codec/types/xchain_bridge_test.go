package types

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
)

func TestXChainBridge_FromJson(t *testing.T) {
	tt := []struct {
		name string
		json any
		want []byte
		err  error
	}{
		{
			name: "valid xchain bridge",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: []byte{83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18},
			err:  nil,
		},
		{
			name: "invalid LockingChainDoor classic address",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p1",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: nil,
			err:  errDecodeClassicAddress,
		},
		{
			name: "invalid LockingChainIssue classic address",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p1",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: nil,
			err:  errDecodeClassicAddress,
		},
		{
			name: "invalid IssuingChainDoor classic address",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p1",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: nil,
			err:  errDecodeClassicAddress,
		},
		{
			name: "invalid IssuingChainIssue classic address",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p1",
			},
			want: nil,
			err:  errDecodeClassicAddress,
		},
		{
			name: "not a valid json",
			json: "not a valid json",
			want: nil,
			err:  errNotValidJSON,
		},
		{
			name: "invalid xchain bridge",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: nil,
			err:  errNotValidXChainBridge,
		},
		{
			name: "LockingChainDoor is not a string",
			json: map[string]any{
				"LockingChainDoor":  123,
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: nil,
			err:  errNotValidXChainBridge,
		},
		{
			name: "LockingChainIssue is not a string",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": 123,
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: nil,
			err:  errNotValidXChainBridge,
		},
		{
			name: "IssuingChainDoor is not a string",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  123,
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			want: nil,
			err:  errNotValidXChainBridge,
		},
		{
			name: "IssuingChainIssue is not a string",
			json: map[string]any{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": 123,
			},
			want: nil,
			err:  errNotValidXChainBridge,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			xcb := &XChainBridge{}
			got, err := xcb.FromJSON(tc.json)
			if !errors.Is(err, tc.err) {
				t.Errorf("FromJson() error = %v, want %v", err.Error(), tc.err.Error())
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("FromJson() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestXChainBridge_ToJson(t *testing.T) {
	tt := []struct {
		name  string
		input []byte
		opts  []int
		want  map[string]string
		err   error
		setup func(t *testing.T) (*XChainBridge, interfaces.BinaryParser)
	}{
		{
			name:  "Valid xchain bridge",
			input: []byte{83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18},
			opts:  []int{80},
			want: map[string]string{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			err: nil,
			setup: func(t *testing.T) (*XChainBridge, interfaces.BinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(80).Return([]byte{83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18}, nil)
				return &XChainBridge{}, mock
			},
		},
		{
			name:  "Valid xchain bridge - real parser",
			input: []byte{83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18},
			opts:  []int{80},
			want: map[string]string{
				"LockingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"LockingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainDoor":  "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
				"IssuingChainIssue": "r3e7qTG44Mg8pHXgxPtyRx286Re5Urtx2p",
			},
			err: nil,
			setup: func(t *testing.T) (*XChainBridge, interfaces.BinaryParser) {
				payload := []byte{83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18}
				return &XChainBridge{}, serdes.NewBinaryParser(payload, definitions.Get())
			},
		},
		{
			name:  "No length prefix",
			input: []byte{83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18},
			opts:  nil,
			want:  nil,
			err:   ErrNoLengthPrefix,
			setup: func(t *testing.T) (*XChainBridge, interfaces.BinaryParser) {
				return &XChainBridge{}, nil
			},
		},
		{
			name:  "ReadBytes error",
			input: []byte{83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18, 83, 223, 129, 195, 127, 70, 21, 146, 66, 247, 202, 145, 99, 224, 159, 4, 64, 41, 204, 18},
			opts:  []int{80},
			want:  nil,
			err:   errReadBytes,
			setup: func(t *testing.T) (*XChainBridge, interfaces.BinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(80).Return([]byte{}, errors.New("errReadBytes"))
				return &XChainBridge{}, mock
			},
		},
		{
			name:  "Short bytes",
			input: nil,
			opts:  []int{80},
			want:  nil,
			err:   errNotValidXChainBridge,
			setup: func(t *testing.T) (*XChainBridge, interfaces.BinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(80).Return(make([]byte, 60), nil)
				return &XChainBridge{}, mock
			},
		},
		{
			name:  "Nil bytes",
			input: nil,
			opts:  []int{80},
			want:  nil,
			err:   errNotValidXChainBridge,
			setup: func(t *testing.T) (*XChainBridge, interfaces.BinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(80).Return(nil, nil)
				return &XChainBridge{}, mock
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			xcb, parser := tc.setup(t)
			got, err := xcb.ToJSON(parser, tc.opts...)
			if !errors.Is(err, tc.err) {
				t.Errorf("ToJson() error = %v, want %v", err.Error(), tc.err.Error())
			} else if tc.err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ToJson() got = %v, want %v", got, tc.want)
			}
		})
	}
}

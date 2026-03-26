//go:build cgo

package proofs_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/confidential/proofs"
	"github.com/stretchr/testify/require"
)

func TestConvertContextHash(t *testing.T) {
	tests := []struct {
		name    string
		account string
		issID   string
		seq     uint32
		wantErr error
	}{
		{"pass - valid inputs", testAccount, testIssuanceID, 1, nil},
		{"fail - invalid address", "notAnAddress", testIssuanceID, 1, proofs.ErrInvalidAddress},
		{"fail - invalid issuance ID", testAccount, "zz", 1, proofs.ErrInvalidIssuanceIDLength},
		{"fail - short issuance ID", testAccount, "0102", 1, proofs.ErrInvalidIssuanceIDLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proofs.ConvertContextHash(tt.account, tt.issID, tt.seq)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptcrypto.HashOutputSize*2)
		})
	}
}

func TestConvertContextHashDeterministic(t *testing.T) {
	h1, err := proofs.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)

	h2, err := proofs.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)

	require.Equal(t, h1, h2, "same inputs should produce the same hash")
}

func TestConvertBackContextHash(t *testing.T) {
	tests := []struct {
		name    string
		account string
		issID   string
		seq     uint32
		ver     uint32
		wantErr error
	}{
		{"pass - valid inputs", testAccount, testIssuanceID, 1, 0, nil},
		{"fail - invalid address", "bad", testIssuanceID, 1, 0, proofs.ErrInvalidAddress},
		{"fail - invalid issuance ID", testAccount, "bad", 1, 0, proofs.ErrInvalidIssuanceIDLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proofs.ConvertBackContextHash(tt.account, tt.issID, tt.seq, tt.ver)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptcrypto.HashOutputSize*2)
		})
	}
}

func TestSendContextHash(t *testing.T) {
	tests := []struct {
		name    string
		account string
		issID   string
		seq     uint32
		dest    string
		ver     uint32
		wantErr error
	}{
		{"pass - valid inputs", testAccount, testIssuanceID, 1, testDest, 0, nil},
		{"fail - invalid account", "bad", testIssuanceID, 1, testDest, 0, proofs.ErrInvalidAddress},
		{"fail - invalid dest", testAccount, testIssuanceID, 1, "bad", 0, proofs.ErrInvalidAddress},
		{"fail - invalid issuance ID", testAccount, "zz", 1, testDest, 0, proofs.ErrInvalidIssuanceIDLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proofs.SendContextHash(tt.account, tt.issID, tt.seq, tt.dest, tt.ver)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptcrypto.HashOutputSize*2)
		})
	}
}

func TestClawbackContextHash(t *testing.T) {
	tests := []struct {
		name    string
		account string
		issID   string
		seq     uint32
		holder  string
		wantErr error
	}{
		{"pass - valid inputs", testAccount, testIssuanceID, 1, testHolder, nil},
		{"fail - invalid account", "bad", testIssuanceID, 1, testHolder, proofs.ErrInvalidAddress},
		{"fail - invalid holder", testAccount, testIssuanceID, 1, "bad", proofs.ErrInvalidAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proofs.ClawbackContextHash(tt.account, tt.issID, tt.seq, tt.holder)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptcrypto.HashOutputSize*2)
		})
	}
}

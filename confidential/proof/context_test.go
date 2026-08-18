//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
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
		{"fail - invalid address", "notAnAddress", testIssuanceID, 1, proof.ErrInvalidAddress},
		{"fail - invalid issuance ID", testAccount, "zz", 1, proof.ErrInvalidIssuanceID},
		{"fail - short issuance ID", testAccount, "0102", 1, proof.ErrInvalidIssuanceID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proof.ConvertContextHash(tt.account, tt.issID, tt.seq)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptsizes.HashOutputSize*2)
		})
	}
}

func TestContextHashReferenceVectors(t *testing.T) {
	tests := []struct {
		name     string
		generate func() (string, error)
		want     string
	}{
		{name: "convert", generate: func() (string, error) { return proof.ConvertContextHash(testAccount, testIssuanceID, 1) }, want: "977ec1bd2ba79215c4f4b90a6fb1ee5e4fb819b96a3800c5b15251d12d4359bd"},
		{name: "convert back", generate: func() (string, error) { return proof.ConvertBackContextHash(testAccount, testIssuanceID, 1, 7) }, want: "cb8e464e73b66c92a508f3a834f174a85b8e39aab466a0caa04a5fb75b4a2cbd"},
		{name: "send", generate: func() (string, error) { return proof.SendContextHash(testAccount, testIssuanceID, 1, testDest, 7) }, want: "6470c7add7960234c584572d7cb6423e807c01feceea69d905c8d8218b3503d9"},
		{name: "clawback", generate: func() (string, error) { return proof.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder) }, want: "e2ef94a2a2ed1386fd8df453be61f282b0a51a843c498749916d2456bd58b1a1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.generate()
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
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
		{"fail - invalid address", "bad", testIssuanceID, 1, 0, proof.ErrInvalidAddress},
		{"fail - invalid issuance ID", testAccount, "bad", 1, 0, proof.ErrInvalidIssuanceID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proof.ConvertBackContextHash(tt.account, tt.issID, tt.seq, tt.ver)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptsizes.HashOutputSize*2)
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
		{"fail - invalid account", "bad", testIssuanceID, 1, testDest, 0, proof.ErrInvalidAddress},
		{"fail - invalid dest", testAccount, testIssuanceID, 1, "bad", 0, proof.ErrInvalidAddress},
		{"fail - invalid issuance ID", testAccount, "zz", 1, testDest, 0, proof.ErrInvalidIssuanceID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proof.SendContextHash(tt.account, tt.issID, tt.seq, tt.dest, tt.ver)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptsizes.HashOutputSize*2)
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
		{"fail - invalid account", "bad", testIssuanceID, 1, testHolder, proof.ErrInvalidAddress},
		{"fail - invalid holder", testAccount, testIssuanceID, 1, "bad", proof.ErrInvalidAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := proof.ClawbackContextHash(tt.account, tt.issID, tt.seq, tt.holder)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, hash, mptsizes.HashOutputSize*2)
		})
	}
}

package hash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		sequence  uint32
		want      string
		wantError bool
	}{
		{
			name:     "calcVaultEntryHash",
			address:  "rDcMtA1XpH5DGwiaqFif2cYCvgk5vxHraS",
			sequence: 18,
			want:     "9C3208D7F99E5644643542518859401A96C93D80CC5F757AF0DF1781046C0A6A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Vault(tt.address, tt.sequence)
			if tt.wantError {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestLoanBroker(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		sequence  uint32
		want      string
		wantError bool
	}{
		{
			name:     "calcLoanBrokerHash",
			address:  "rNTrjogemt4dZD13PaqphezBWSmiApNH4K",
			sequence: 84,
			want:     "E799B84AC949CE2D8F27435C784F15C72E6A23ACA6841BA6D2F37A1E5DA4110F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoanBroker(tt.address, tt.sequence)
			if tt.wantError {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestLoan(t *testing.T) {
	tests := []struct {
		name         string
		loanBrokerID string
		loanSequence uint32
		want         string
		wantError    bool
	}{
		{
			name:         "calcLoanHash",
			loanBrokerID: "AEB642A65066A6E6F03D312713475D958E0B320B74AD1A76B5B2EABB752E52AA",
			loanSequence: 1,
			want:         "E93874AB62125DF2E86FB6C724B261F8E654E0334715C4D7160C0F148CDC9B47",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Loan(tt.loanBrokerID, tt.loanSequence)
			if tt.wantError {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

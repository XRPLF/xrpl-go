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

func TestMPToken(t *testing.T) {
	tests := []struct {
		name       string
		issuanceID string
		holder     string
		want       string
		wantError  bool
	}{
		{
			name:       "pass - valid inputs",
			issuanceID: "000000000000000000000000000000000000000000000001",
			holder:     "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
			want:       "421477BB4C4F7195FD4934C1161BCEE697A3472EAE4E176FEE33DB7A3DD46C3F",
		},
		{
			name:       "pass - different issuance ID",
			issuanceID: "000000000000000000000000000000000000000000000002",
			holder:     "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
			want:       "CF6A7BB8B75ACBE74B0D6D1DF6B446735000DC1D6B512AF0CABE28E551477619",
		},
		{
			name:       "fail - wrong issuance ID length",
			issuanceID: "0001",
			holder:     "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
			wantError:  true,
		},
		{
			name:       "fail - invalid address",
			issuanceID: "000000000000000000000000000000000000000000000001",
			holder:     "invalid",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MPToken(tt.issuanceID, tt.holder)
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

func TestMPTokenIssuance(t *testing.T) {
	tests := []struct {
		name       string
		issuanceID string
		want       string
		wantError  bool
	}{
		{
			name:       "pass - valid issuance ID",
			issuanceID: "000000000000000000000000000000000000000000000001",
			want:       "35AE3B1DD171EC091E8FE05D102B7AA5D9A40AA191EBEE00E4536EB677DF7879",
		},
		{
			name:       "fail - wrong length",
			issuanceID: "0001",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MPTokenIssuance(tt.issuanceID)
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

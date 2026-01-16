package wallet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthorizeChannel(t *testing.T) {
	// SECP256K1
	secpWallet, err := FromSeed("snGHNrPbHrdUcszeuDEigMdC1Lyyd", "")
	require.NoError(t, err)

	// ED25519
	edWallet, err := FromSeed("sEdSuqBPSQaood2DmNYVkwWTn1oQTj2", "")
	require.NoError(t, err)

	validChannelID := "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3"
	validAmount := "1000000"

	tc := []struct {
		name        string
		wallet      Wallet
		channelID   string
		amount      string
		expectedSig string
		expectError bool
	}{
		{
			name:        "pass - succeeds with secp256k1 seed",
			wallet:      secpWallet,
			channelID:   validChannelID,
			amount:      validAmount,
			expectedSig: "304402204E7052F33DDAFAAA55C9F5B132A5E50EE95B2CF68C0902F61DFE77299BC893740220353640B951DCD24371C16868B3F91B78D38B6F3FD1E826413CDF891FA8250AAC",
			expectError: false,
		},
		{
			name:        "pass - succeeds with ed25519 seed",
			wallet:      edWallet,
			channelID:   validChannelID,
			amount:      validAmount,
			expectedSig: "7E1C217A3E4B3C107B7A356E665088B4FBA6464C48C58267BEF64975E3375EA338AE22E6714E3F5E734AE33E6B97AAD59058E1E196C1F92346FC1498D0674404",
			expectError: false,
		},
		{
			name:        "pass - different amounts with secp256k1",
			wallet:      secpWallet,
			channelID:   validChannelID,
			amount:      "5000000",
			expectedSig: "304402202DF006FDE665C8A15628991A946629DDD08F7677E75C54619A96E9872BCC615F02206689262B5F102992346E5D84CA4EC73E947906073E4B2873DCDBEE54AFE948C3",
			expectError: false,
		},
		{
			name:        "pass - different amounts with ed25519",
			wallet:      edWallet,
			channelID:   validChannelID,
			amount:      "5000000",
			expectedSig: "AEEFCF001061F4E0368805B8A56D116EA8B9E4879A69C5B56A5B7E0F6ABD63E63341D56247192104012BC6AAEA71B1C97E466F47DA0736EFAD462481B165FB0E",
			expectError: false,
		},
		{
			name:        "pass - zero amount",
			wallet:      secpWallet,
			channelID:   validChannelID,
			amount:      "0",
			expectedSig: "3044022069888D92E1F4104FAD7BA66D8DA69278E579FE6EDAF32E87ACF481A6383C4AEB02204E0286429FF9842724627A08EAFD1A8356A6B36994DA6F386D2363C5D3AAFE7C",
			expectError: false,
		},
		{
			name:        "fail - fails with invalid channel ID format",
			wallet:      secpWallet,
			channelID:   "invalid-id",
			amount:      validAmount,
			expectedSig: "",
			expectError: true,
		},
		{
			name:        "fail - fails with invalid amount format",
			wallet:      secpWallet,
			channelID:   validChannelID,
			amount:      "invalid-amount",
			expectedSig: "",
			expectError: true,
		},
		{
			name:        "fail - empty channel ID",
			wallet:      secpWallet,
			channelID:   "",
			amount:      validAmount,
			expectedSig: "",
			expectError: true,
		},
		{
			name:        "fail - empty amount",
			wallet:      secpWallet,
			channelID:   validChannelID,
			amount:      "",
			expectedSig: "",
			expectError: true,
		},
		{
			name:        "fail - channel ID too short",
			wallet:      secpWallet,
			channelID:   "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB",
			amount:      validAmount,
			expectedSig: "",
			expectError: true,
		},
		{
			name:        "fail - channel ID too long",
			wallet:      secpWallet,
			channelID:   "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3A",
			amount:      validAmount,
			expectedSig: "",
			expectError: true,
		},
		{
			name:        "fail - channel ID with invalid hex characters",
			wallet:      secpWallet,
			channelID:   "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDBG",
			amount:      validAmount,
			expectedSig: "",
			expectError: true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := AuthorizeChannel(tt.channelID, tt.amount, tt.wallet)
			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, signature)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, signature)
				require.Equal(t, tt.expectedSig, signature)
			}
		})
	}
}

package rpc

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/vault"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestClient_GetVaultInfo(t *testing.T) {
	// Focus on share-field decoding through the client, not a specific on-chain vault.
	mockResponse := `{"result":{"vault":{"shares":{
		"Issuer":"rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
		"LedgerEntryType":"MPTokenIssuance",
		"OutstandingAmount":"1000000",
		"AssetScale":6,
		"MaximumAmount":"9223372036854775807",
		"TransferFee":50000,
		"MPTokenMetadata":"534841524553",
		"LockedAmount":"9007199254740993",
		"ReferenceHolding":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8"
	}}}}`
	mockClient := testutil.JSONRPCMockClient{}
	mockClient.DoFunc = testutil.MockResponse(mockResponse, 200, &mockClient)
	config, err := NewClientConfig("http://testnode/", WithHTTPClient(&mockClient))
	require.NoError(t, err)

	response, err := NewClient(config).GetVaultInfo(&vault.InfoRequest{
		VaultID: "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
	})
	require.NoError(t, err)
	require.Equal(t, vault.Shares{
		Issuer:            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
		LedgerEntryType:   "MPTokenIssuance",
		OutstandingAmount: "1000000",
		AssetScale:        6,
		MaximumAmount:     "9223372036854775807",
		TransferFee:       50000,
		MPTokenMetadata:   "534841524553",
		LockedAmount:      "9007199254740993",
		ReferenceHolding:  types.Hash256("13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8"),
	}, response.Vault.Shares)
}

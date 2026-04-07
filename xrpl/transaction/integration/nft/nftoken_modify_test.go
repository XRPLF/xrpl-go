package integration

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type NFTokenModifyTest struct {
	Name        string
	NFTokenMint *transaction.NFTokenMint
}

func testIntegrationNFTokenModify(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	minter := runner.GetWallet(0)

	originalURI := types.NFTokenURI(hex.EncodeToString([]byte("https://xrpl.org/")))
	updatedURI := types.NFTokenURI(hex.EncodeToString([]byte("https://github.com/XRPLF/xrpl-go")))

	tt := []NFTokenModifyTest{
		{
			Name: "pass - modify NFToken",
			NFTokenMint: &transaction.NFTokenMint{
				BaseTx: transaction.BaseTx{
					Account: minter.GetAddress(),
				},
				NFTokenTaxon: 0,
				URI:          originalURI,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			mintTx := tc.NFTokenMint
			mintTx.SetMutableFlag()
			flatMintTx := mintTx.Flatten()
			_, err := runner.TestTransaction(&flatMintTx, minter, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountNFTs(&account.NFTsRequest{
				Account: minter.GetAddress(),
			})
			require.NoError(t, err)
			require.Len(t, objects.AccountNFTs, 1)
			require.Equal(t, strings.ToLower(objects.AccountNFTs[0].URI.String()), strings.ToLower(originalURI.String()))

			nftID := objects.AccountNFTs[0].NFTokenID
			modifyTokenTx := transaction.NFTokenModify{
				BaseTx: transaction.BaseTx{
					Account: minter.GetAddress(),
				},
				NFTokenID: nftID,
				URI:       updatedURI,
			}

			flatModifyTokenTx := modifyTokenTx.Flatten()
			_, err = runner.TestTransaction(&flatModifyTokenTx, minter, "tesSUCCESS", nil)
			require.NoError(t, err)

			updatedNFTs, err := client.GetAccountNFTs(&account.NFTsRequest{
				Account: minter.GetAddress(),
			})
			require.NoError(t, err)
			require.Len(t, updatedNFTs.AccountNFTs, 1)
			require.Equal(t, strings.ToLower(updatedURI.String()), strings.ToLower(updatedNFTs.AccountNFTs[0].URI.String()))
		})
	}
}

func TestIntegrationNFTokenModify_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationNFTokenModify(t, client)
}

func TestIntegrationNFTokenModify_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationNFTokenModify(t, client)
}

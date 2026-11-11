package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestAMMBidFlatten(t *testing.T) {
	s := AMMBid{
		BaseTx: BaseTx{
			Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			TransactionType: AMMCreateTx,
			Fee:             types.XRPCurrencyAmount(10),
			Sequence:        1234,
			SigningPubKey:   "ghijk",
			TxnSignature:    "A1B2C3D4E5F6",
		},
		Asset:  ledger.Asset{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
		Asset2: ledger.Asset{Currency: "XRP"},
		BidMin: types.XRPCurrencyAmount(100),
		BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
		AuthAccounts: []ledger.AuthAccounts{
			{
				AuthAccount: ledger.AuthAccount{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
			},
			{
				AuthAccount: ledger.AuthAccount{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
			},
		},
	}

	flattened := s.Flatten()

	expected := `{
	"Account":         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"TransactionType": "AMMBid",
	"Fee":             "10",
	"Sequence":        1234,
	"SigningPubKey":   "ghijk",
	"TxnSignature":    "A1B2C3D4E5F6",
	"Asset": {
		"currency": "USD",
		"issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
	},
	"Asset2": {
		"currency": "XRP"
	},
	"BidMin": "100",
	"BidMax": {
		"currency": "USD",
		"issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		"value": "200"
	},
	"AuthAccounts": [
		{
			"AuthAccount": {
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
			}
		},
		{
			"AuthAccount": {
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
			}
		}
	]
}`

	err := testutil.CompareFlattenAndExpected(flattened, []byte(expected))
	if err != nil {
		t.Error(err)
	}
}

func TestAMMBid_Validate(t *testing.T) {
	iou := ledger.Asset{Currency: "USD", Issuer: testAddrDelegate}
	xrp := ledger.Asset{Currency: "XRP"}
	tests := []struct {
		name    string
		ammBid  AMMBid
		wantErr bool
	}{
		{
			name: "fail - zero XRP BidMin",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, BidMin: types.XRPCurrencyAmount(0),
			},
			wantErr: true,
		},
		{
			name: "fail - zero IOU BidMin",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, BidMin: types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "0"},
			},
			wantErr: true,
		},
		{
			name: "fail - zero MPT BidMin",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, BidMin: types.MPTCurrencyAmount{MPTIssuanceID: "00000001ABCDEF0123456789ABCDEF0123456789ABCDEF01", Value: "0"},
			},
			wantErr: true,
		},
		{
			name: "fail - zero XRP BidMax",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, BidMax: types.XRPCurrencyAmount(0),
			},
			wantErr: true,
		},
		{
			name: "fail - zero IOU BidMax",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "0"},
			},
			wantErr: true,
		},
		{
			name: "fail - zero MPT BidMax",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, BidMax: types.MPTCurrencyAmount{MPTIssuanceID: "00000001ABCDEF0123456789ABCDEF0123456789ABCDEF01", Value: "0"},
			},
			wantErr: true,
		},
		{
			name: "pass - omitted bid limits",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp,
			},
		},
		{
			name: "fail - BidMin above BidMax with the same issue",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp,
				BidMin: types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "10"},
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "9.5"},
			},
			wantErr: true,
		},
		{
			name: "pass - BidMin equal to BidMax",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp,
				BidMin: types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "10"},
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "1e1"},
			},
		},
		{
			name: "pass - pointer bid limits",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp,
				BidMin: &types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "9"},
				BidMax: &types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "10"},
			},
		},
		{
			name: "fail - pointer BidMin above BidMax",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp,
				BidMin: &types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "10"},
				BidMax: &types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "9"},
			},
			wantErr: true,
		},
		{
			name: "pass - BidMin above BidMax with different issues",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp,
				BidMin: types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "10"},
				BidMax: types.IssuedCurrencyAmount{Currency: "EUR", Issuer: testAddrDelegate, Value: "5"},
			},
		},
		{
			name: "fail - sender AuthAccount",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, AuthAccounts: []ledger.AuthAccounts{{AuthAccount: ledger.AuthAccount{Account: testAddrAccount}}},
			},
			wantErr: true,
		},
		{
			name: "fail - duplicate AuthAccounts",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, AuthAccounts: []ledger.AuthAccounts{
					{AuthAccount: ledger.AuthAccount{Account: testAddrDelegate}}, {AuthAccount: ledger.AuthAccount{Account: testAddrDelegate}},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - duplicate classic and X-address AuthAccounts",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, AuthAccounts: []ledger.AuthAccounts{
					{AuthAccount: ledger.AuthAccount{Account: testAddrDelegate}}, {AuthAccount: ledger.AuthAccount{Account: xAddr(t, testAddrDelegate, 0, false)}},
				},
			},
			wantErr: true,
		},
		{
			name: "pass - four distinct AuthAccounts",
			ammBid: AMMBid{
				BaseTx: BaseTx{Account: testAddrAccount, TransactionType: AMMBidTx},
				Asset:  iou, Asset2: xrp, AuthAccounts: []ledger.AuthAccounts{
					{AuthAccount: ledger.AuthAccount{Account: testAddrDelegate}},
					{AuthAccount: ledger.AuthAccount{Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"}},
					{AuthAccount: ledger.AuthAccount{Account: "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"}},
					{AuthAccount: ledger.AuthAccount{Account: "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"}},
				},
			},
		},
		{
			name: "pass - valid AMMBid",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
				Asset2: ledger.Asset{Currency: "XRP"},
				BidMin: types.XRPCurrencyAmount(100),
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: testAddrAccount,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "fail - invalid BaseTx AMMBid, Account missing",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
				Asset2: ledger.Asset{Currency: "XRP"},
				BidMin: types.XRPCurrencyAmount(100),
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid AMMBid with more than 4 AuthAccounts",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
				Asset2: ledger.Asset{Currency: "XRP"},
				BidMin: types.XRPCurrencyAmount(100),
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcA",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcB",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcC",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid AMMBid with more an AuthAccount with invalid address",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
				Asset2: ledger.Asset{Currency: "XRP"},
				BidMin: types.XRPCurrencyAmount(100),
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcA",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "invalid",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid AMMBid with invalid Asset, currency empty",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
				Asset2: ledger.Asset{Currency: "XRP"},
				BidMin: types.XRPCurrencyAmount(100),
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid AMMBid with invalid Asset2, issuer empty with currency non empty",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "XRP"},
				Asset2: ledger.Asset{Currency: "USD", Issuer: ""},
				BidMin: types.XRPCurrencyAmount(100),
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid AMMBid with invalid Asset and Asset2, two XRP assets",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "XRP"},
				Asset2: ledger.Asset{Currency: "XRP"},
				BidMin: types.XRPCurrencyAmount(100),
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid AMMBid with invalid BidMin, missing value and issuer",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "XRP"},
				Asset2: ledger.Asset{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
				BidMin: types.IssuedCurrencyAmount{Currency: "USD"}, // missing value and issuer
				BidMax: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid AMMBid with invalid BidMax, missing value and issuer",
			ammBid: AMMBid{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: AMMBidTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Asset:  ledger.Asset{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
				Asset2: ledger.Asset{Currency: "XRP"},
				BidMin: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", Value: "200"},
				BidMax: types.IssuedCurrencyAmount{Currency: "USD"}, // missing value and issuer
				AuthAccounts: []ledger.AuthAccounts{
					{
						AuthAccount: ledger.AuthAccount{
							Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						},
					},
					{
						AuthAccount: ledger.AuthAccount{
							Account: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcE",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.ammBid.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AMMBid.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != !tt.wantErr {
				t.Errorf("AMMBid.Validate() = %v, want %v", valid, !tt.wantErr)
			}
		})
	}
}

func TestAMMBid_TxType(t *testing.T) {
	entry := &AMMBid{}
	assert.Equal(t, AMMBidTx, entry.TxType())
}

func TestAMMBid_RejectsTwoNativeAssets(t *testing.T) {
	const nativeHex = "0000000000000000000000000000000000000000"
	for _, currencies := range [][2]string{{"XRP", "XRP"}, {nativeHex, nativeHex}, {"XRP", nativeHex}, {nativeHex, "XRP"}} {
		t.Run(currencies[0]+"/"+currencies[1], func(t *testing.T) {
			tx := AMMBid{
				BaseTx: BaseTx{Account: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1", TransactionType: AMMBidTx},
				Asset:  ledger.Asset{Currency: currencies[0]},
				Asset2: ledger.Asset{Currency: currencies[1]},
			}
			ok, err := tx.Validate()
			assert.False(t, ok)
			assert.ErrorIs(t, err, ErrAMMAtLeastOneAssetMustBeNonXRP)
		})
	}
}

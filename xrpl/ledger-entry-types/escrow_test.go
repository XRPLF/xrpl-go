package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestEscrow(t *testing.T) {
	var s Object = &Escrow{
		LedgerEntryType:   EscrowEntry,
		Flags:             0,
		Account:           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		Amount:            types.XRPCurrencyAmount(10000),
		CancelAfter:       545440232,
		Condition:         "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
		Destination:       "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		DestinationNode:   "0000000000000000",
		DestinationTag:    23480,
		FinishAfter:       545354132,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq: 28991004,
		SourceTag:         11747,
		TransferRate:      1000,
		IssuerNode:        1234567890,
	}

	j := `{
	"LedgerEntryType": "Escrow",
	"Flags": 0,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Amount": "10000",
	"CancelAfter": 545440232,
	"Condition": "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
	"Destination": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"DestinationNode": "0000000000000000",
	"DestinationTag": 23480,
	"FinishAfter": 545354132,
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004,
	"SourceTag": 11747,
	"TransferRate": 1000,
	"IssuerNode": 1234567890
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestEscrow_EntryType(t *testing.T) {
	s := &Escrow{}
	require.Equal(t, s.EntryType(), EscrowEntry)
}

func TestEscrowMPTAmountSerialization(t *testing.T) {
	var s Object = &Escrow{
		LedgerEntryType: EscrowEntry,
		Flags:           0,
		Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
			Value:         "10000",
		},
		CancelAfter:       545440232,
		Condition:         "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
		Destination:       "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		DestinationNode:   "0000000000000000",
		DestinationTag:    23480,
		FinishAfter:       545354132,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq: 28991004,
		SourceTag:         11747,
		TransferRate:      1000,
		IssuerNode:        1234567890,
	}

	j := `{
	"LedgerEntryType": "Escrow",
	"Flags": 0,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Amount": {
		"mpt_issuance_id": "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
		"value": "10000"
	},
	"CancelAfter": 545440232,
	"Condition": "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
	"Destination": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"DestinationNode": "0000000000000000",
	"DestinationTag": 23480,
	"FinishAfter": 545354132,
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004,
	"SourceTag": 11747,
	"TransferRate": 1000,
	"IssuerNode": 1234567890
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestEscrowIssuedAmountSerialization(t *testing.T) {
	var s Object = &Escrow{
		LedgerEntryType: EscrowEntry,
		Flags:           0,
		Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		Amount: types.IssuedCurrencyAmount{
			Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			Currency: "USD",
			Value:    "10000",
		},
		CancelAfter:       545440232,
		Condition:         "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
		Destination:       "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		DestinationNode:   "0000000000000000",
		DestinationTag:    23480,
		FinishAfter:       545354132,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq: 28991004,
		SourceTag:         11747,
		TransferRate:      1000,
		IssuerNode:        1234567890,
	}

	j := `{
	"LedgerEntryType": "Escrow",
	"Flags": 0,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Amount": {
		"issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		"currency": "USD",
		"value": "10000"
	},
	"CancelAfter": 545440232,
	"Condition": "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
	"Destination": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"DestinationNode": "0000000000000000",
	"DestinationTag": 23480,
	"FinishAfter": 545354132,
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004,
	"SourceTag": 11747,
	"TransferRate": 1000,
	"IssuerNode": 1234567890
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

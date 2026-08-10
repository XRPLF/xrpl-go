package ledger

import (
	"encoding/json"
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
	"SourceTag": 11747
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestEscrow_EntryType(t *testing.T) {
	s := &Escrow{}
	require.Equal(t, EscrowEntry, s.EntryType())
}

func TestEscrowIssuerNodeSerialization(t *testing.T) {
	tests := []struct {
		name         string
		issuerNode   string
		shouldEncode bool
	}{
		{name: "omit empty optional value"},
		{name: "root page", issuerNode: "0", shouldEncode: true},
		{name: "second page returned by rippled", issuerNode: "1", shouldEncode: true},
		{name: "hexadecimal page", issuerNode: "1f", shouldEncode: true},
		{name: "maximum UInt64 page", issuerNode: "FFFFFFFFFFFFFFFF", shouldEncode: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			escrow := &Escrow{
				Amount:     types.XRPCurrencyAmount(1),
				IssuerNode: test.issuerNode,
			}

			encoded, err := json.Marshal(escrow)
			require.NoError(t, err)

			fields := make(map[string]json.RawMessage)
			require.NoError(t, json.Unmarshal(encoded, &fields))
			encodedIssuerNode, ok := fields["IssuerNode"]
			require.Equal(t, test.shouldEncode, ok)
			if test.shouldEncode {
				var issuerNode string
				require.NoError(t, json.Unmarshal(encodedIssuerNode, &issuerNode))
				require.Equal(t, test.issuerNode, issuerNode)
			}

			var decoded Escrow
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			require.Equal(t, test.issuerNode, decoded.IssuerNode)
		})
	}
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
		DestinationNode:   "1",
		DestinationTag:    23480,
		FinishAfter:       545354132,
		OwnerNode:         "1",
		PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq: 28991004,
		SourceTag:         11747,
		TransferRate:      1000,
		IssuerNode:        "1",
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
	"DestinationNode": "1",
	"DestinationTag": 23480,
	"FinishAfter": 545354132,
	"OwnerNode": "1",
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004,
	"SourceTag": 11747,
	"TransferRate": 1000,
	"IssuerNode": "1"
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
			Issuer:   "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			Currency: "USD",
			Value:    "10000",
		},
		CancelAfter:       545440232,
		Condition:         "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
		Destination:       "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		DestinationNode:   "17",
		DestinationTag:    23480,
		FinishAfter:       545354132,
		OwnerNode:         "1f",
		PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq: 28991004,
		SourceTag:         11747,
		TransferRate:      1000,
		IssuerNode:        "1f",
	}

	j := `{
	"LedgerEntryType": "Escrow",
	"Flags": 0,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Amount": {
		"issuer": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
		"currency": "USD",
		"value": "10000"
	},
	"CancelAfter": 545440232,
	"Condition": "A0258020A82A88B2DF843A54F58772E4A3861866ECDB4157645DD9AE528C1D3AEEDABAB6810120",
	"Destination": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"DestinationNode": "17",
	"DestinationTag": 23480,
	"FinishAfter": 545354132,
	"OwnerNode": "1f",
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004,
	"SourceTag": 11747,
	"TransferRate": 1000,
	"IssuerNode": "1f"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

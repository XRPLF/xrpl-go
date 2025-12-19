package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestTxMetadataBuilder_AsTxObjMeta(t *testing.T) {
	batchId := types.BatchID("ABCD1234")

	tests := []struct {
		name     string
		builder  TxMetadataBuilder
		expected TxObjMeta
	}{
		{
			name: "pass - all fields populated",
			builder: TxMetadataBuilder{
				AffectedNodes: []AffectedNode{
					{
						CreatedNode: &CreatedNode{
							LedgerEntryType: "AccountRoot",
							LedgerIndex:     "123",
						},
					},
				},
				PartialDeliveredAmount: "1000000",
				TransactionIndex:       42,
				TransactionResult:      "tesSUCCESS",
				DeliveredAmount:        "1000000",
				ParentBatchID:          &batchId,
			},
			expected: TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						CreatedNode: &CreatedNode{
							LedgerEntryType: "AccountRoot",
							LedgerIndex:     "123",
						},
					},
				},
				PartialDeliveredAmount: "1000000",
				TransactionIndex:       42,
				TransactionResult:      "tesSUCCESS",
				DeliveredAmount:        "1000000",
				ParentBatchID:          &batchId,
			},
		},
		{
			name: "pass - minimal fields",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
			},
			expected: TxObjMeta{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
			},
		},
		{
			name: "pass - nil optional fields",
			builder: TxMetadataBuilder{
				AffectedNodes:          []AffectedNode{},
				PartialDeliveredAmount: nil,
				TransactionIndex:       0,
				TransactionResult:      "tesSUCCESS",
				DeliveredAmount:        nil,
				ParentBatchID:          nil,
			},
			expected: TxObjMeta{
				AffectedNodes:          []AffectedNode{},
				PartialDeliveredAmount: nil,
				TransactionIndex:       0,
				TransactionResult:      "tesSUCCESS",
				DeliveredAmount:        nil,
				ParentBatchID:          nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.AsTxObjMeta()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxMetadataBuilder_AsPaymentMetadata(t *testing.T) {
	batchID := types.BatchID("EFGH5678")

	tests := []struct {
		name     string
		builder  TxMetadataBuilder
		expected PaymentMetadata
	}{
		{
			name: "pass - all fields populated",
			builder: TxMetadataBuilder{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							LedgerEntryType: "AccountRoot",
							LedgerIndex:     "456",
						},
					},
				},
				PartialDeliveredAmount: "5000000",
				TransactionIndex:       10,
				TransactionResult:      "tesSUCCESS",
				DeliveredAmount:        "5000000",
				ParentBatchID:          &batchID,
			},
			expected: PaymentMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes: []AffectedNode{
						{
							ModifiedNode: &ModifiedNode{
								LedgerEntryType: "AccountRoot",
								LedgerIndex:     "456",
							},
						},
					},
					PartialDeliveredAmount: "5000000",
					TransactionIndex:       10,
					TransactionResult:      "tesSUCCESS",
					DeliveredAmount:        "5000000",
					ParentBatchID:          &batchID,
				},
			},
		},
		{
			name: "pass - minimal fields",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
			},
			expected: PaymentMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.AsPaymentMetadata()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxMetadataBuilder_AsNFTokenMintMetadata(t *testing.T) {
	nftokenID := types.NFTokenID("000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007")
	offerID := types.OfferID("9C92E061381C1EF37A8CDE0E8FC35188BFC30B1883825042A64309AC09F4C36D")

	tests := []struct {
		name     string
		builder  TxMetadataBuilder
		expected NFTokenMintMetadata
	}{
		{
			name: "pass - all fields populated",
			builder: TxMetadataBuilder{
				AffectedNodes: []AffectedNode{
					{
						CreatedNode: &CreatedNode{
							LedgerEntryType: "NFTokenPage",
							LedgerIndex:     "789",
						},
					},
				},
				TransactionIndex:  5,
				TransactionResult: "tesSUCCESS",
				NFTokenID:         &nftokenID,
				OfferID:           &offerID,
			},
			expected: NFTokenMintMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes: []AffectedNode{
						{
							CreatedNode: &CreatedNode{
								LedgerEntryType: "NFTokenPage",
								LedgerIndex:     "789",
							},
						},
					},
					TransactionIndex:  5,
					TransactionResult: "tesSUCCESS",
				},
				NFTokenID: &nftokenID,
				OfferID:   &offerID,
			},
		},
		{
			name: "pass - with NFTokenID only",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				NFTokenID:         &nftokenID,
			},
			expected: NFTokenMintMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				NFTokenID: &nftokenID,
			},
		},
		{
			name: "pass - with OfferID only",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				OfferID:           &offerID,
			},
			expected: NFTokenMintMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				OfferID: &offerID,
			},
		},
		{
			name: "pass - nil optional fields",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				NFTokenID:         nil,
				OfferID:           nil,
			},
			expected: NFTokenMintMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				NFTokenID: nil,
				OfferID:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.AsNFTokenMintMetadata()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxMetadataBuilder_AsNFTokenCreateOfferMetadata(t *testing.T) {
	offerID := types.OfferID("68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77")

	tests := []struct {
		name     string
		builder  TxMetadataBuilder
		expected NFTokenCreateOfferMetadata
	}{
		{
			name: "pass - all fields populated",
			builder: TxMetadataBuilder{
				AffectedNodes: []AffectedNode{
					{
						CreatedNode: &CreatedNode{
							LedgerEntryType: "NFTokenOffer",
							LedgerIndex:     "ABC123",
						},
					},
				},
				TransactionIndex:  15,
				TransactionResult: "tesSUCCESS",
				OfferID:           &offerID,
			},
			expected: NFTokenCreateOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes: []AffectedNode{
						{
							CreatedNode: &CreatedNode{
								LedgerEntryType: "NFTokenOffer",
								LedgerIndex:     "ABC123",
							},
						},
					},
					TransactionIndex:  15,
					TransactionResult: "tesSUCCESS",
				},
				OfferID: &offerID,
			},
		},
		{
			name: "pass - with OfferID",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				OfferID:           &offerID,
			},
			expected: NFTokenCreateOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				OfferID: &offerID,
			},
		},
		{
			name: "pass - nil OfferID",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				OfferID:           nil,
			},
			expected: NFTokenCreateOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				OfferID: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.AsNFTokenCreateOfferMetadata()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxMetadataBuilder_AsNFTokenAcceptOfferMetadata(t *testing.T) {
	nftokenID := types.NFTokenID("000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007")

	tests := []struct {
		name     string
		builder  TxMetadataBuilder
		expected NFTokenAcceptOfferMetadata
	}{
		{
			name: "pass - all fields populated",
			builder: TxMetadataBuilder{
				AffectedNodes: []AffectedNode{
					{
						DeletedNode: &DeletedNode{
							LedgerEntryType: "NFTokenOffer",
							LedgerIndex:     "DEF456",
						},
					},
				},
				TransactionIndex:  20,
				TransactionResult: "tesSUCCESS",
				NFTokenID:         &nftokenID,
			},
			expected: NFTokenAcceptOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes: []AffectedNode{
						{
							DeletedNode: &DeletedNode{
								LedgerEntryType: "NFTokenOffer",
								LedgerIndex:     "DEF456",
							},
						},
					},
					TransactionIndex:  20,
					TransactionResult: "tesSUCCESS",
				},
				NFTokenID: &nftokenID,
			},
		},
		{
			name: "pass - with NFTokenID",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				NFTokenID:         &nftokenID,
			},
			expected: NFTokenAcceptOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				NFTokenID: &nftokenID,
			},
		},
		{
			name: "pass - nil NFTokenID",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				NFTokenID:         nil,
			},
			expected: NFTokenAcceptOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				NFTokenID: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.AsNFTokenAcceptOfferMetadata()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxMetadataBuilder_AsNFTokenCancelOfferMetadata(t *testing.T) {
	nftokenID1 := types.NFTokenID("000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007")
	nftokenID2 := types.NFTokenID("000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000008")
	nftokenID3 := types.NFTokenID("000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000009")

	tests := []struct {
		name     string
		builder  TxMetadataBuilder
		expected NFTokenCancelOfferMetadata
	}{
		{
			name: "pass - multiple NFTokenIDs",
			builder: TxMetadataBuilder{
				AffectedNodes: []AffectedNode{
					{
						DeletedNode: &DeletedNode{
							LedgerEntryType: "NFTokenOffer",
							LedgerIndex:     "GHI789",
						},
					},
				},
				TransactionIndex:  25,
				TransactionResult: "tesSUCCESS",
				NFTokenIDs:        []types.NFTokenID{nftokenID1, nftokenID2, nftokenID3},
			},
			expected: NFTokenCancelOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes: []AffectedNode{
						{
							DeletedNode: &DeletedNode{
								LedgerEntryType: "NFTokenOffer",
								LedgerIndex:     "GHI789",
							},
						},
					},
					TransactionIndex:  25,
					TransactionResult: "tesSUCCESS",
				},
				NFTokenIDs: []types.NFTokenID{nftokenID1, nftokenID2, nftokenID3},
			},
		},
		{
			name: "pass - single NFTokenID",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				NFTokenIDs:        []types.NFTokenID{nftokenID1},
			},
			expected: NFTokenCancelOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				NFTokenIDs: []types.NFTokenID{nftokenID1},
			},
		},
		{
			name: "pass - empty NFTokenIDs",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				NFTokenIDs:        []types.NFTokenID{},
			},
			expected: NFTokenCancelOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				NFTokenIDs: []types.NFTokenID{},
			},
		},
		{
			name: "pass - nil NFTokenIDs",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				NFTokenIDs:        nil,
			},
			expected: NFTokenCancelOfferMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				NFTokenIDs: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.AsNFTokenCancelOfferMetadata()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTxMetadataBuilder_AsMPTokenIssuanceCreateMetadata(t *testing.T) {
	mptIssuanceID := types.MPTIssuanceID("MPT1234567890ABCDEF")

	tests := []struct {
		name     string
		builder  TxMetadataBuilder
		expected MPTokenIssuanceCreateMetadata
	}{
		{
			name: "pass - all fields populated",
			builder: TxMetadataBuilder{
				AffectedNodes: []AffectedNode{
					{
						CreatedNode: &CreatedNode{
							LedgerEntryType: "MPTokenIssuance",
							LedgerIndex:     "JKL012",
						},
					},
				},
				TransactionIndex:  30,
				TransactionResult: "tesSUCCESS",
				MPTIssuanceID:     &mptIssuanceID,
			},
			expected: MPTokenIssuanceCreateMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes: []AffectedNode{
						{
							CreatedNode: &CreatedNode{
								LedgerEntryType: "MPTokenIssuance",
								LedgerIndex:     "JKL012",
							},
						},
					},
					TransactionIndex:  30,
					TransactionResult: "tesSUCCESS",
				},
				MPTIssuanceID: &mptIssuanceID,
			},
		},
		{
			name: "pass - with MPTIssuanceID",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				MPTIssuanceID:     &mptIssuanceID,
			},
			expected: MPTokenIssuanceCreateMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				MPTIssuanceID: &mptIssuanceID,
			},
		},
		{
			name: "pass - nil MPTIssuanceID",
			builder: TxMetadataBuilder{
				AffectedNodes:     []AffectedNode{},
				TransactionResult: "tesSUCCESS",
				MPTIssuanceID:     nil,
			},
			expected: MPTokenIssuanceCreateMetadata{
				TxObjMeta: TxObjMeta{
					AffectedNodes:     []AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				MPTIssuanceID: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder.AsMPTokenIssuanceCreateMetadata()
			require.Equal(t, tt.expected, result)
		})
	}
}

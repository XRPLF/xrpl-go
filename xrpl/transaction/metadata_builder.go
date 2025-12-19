package transaction

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// TxMetadataBuilder contains all `meta` transaction response fields and
// enables specific transaction metadata building.
type TxMetadataBuilder struct {
	AffectedNodes []AffectedNode `json:"AffectedNodes"`

	// PartialDeliveredAmount types.CurrencyAmount `json:"DeliveredAmount,omitempty"`
	PartialDeliveredAmount any    `json:"DeliveredAmount,omitempty"`
	TransactionIndex       uint64 `json:"TransactionIndex,omitempty"`
	TransactionResult      string `json:"TransactionResult"`
	// DeliveredAmount        types.CurrencyAmount `json:"delivered_amount,omitempty"`
	DeliveredAmount any `json:"delivered_amount,omitempty"`

	// ParentBatchID is the hash of the parent Batch transaction when this transaction is executed as part of a batch.
	ParentBatchID *types.BatchID `json:"ParentBatchID,omitempty"`

	// rippled 1.11.0 or later.
	// Only available in: NFTokenMintMetadata, NFTokenAcceptOfferMetadata
	NFTokenID *types.NFTokenID `json:"nftoken_id,omitempty"`

	// OfferID is a string of Amount is present.
	// Only available in: NFTokenMintMetadata, NFTokenCreateOfferMetadata
	OfferID *types.OfferID `json:"offer_id,omitempty"`

	// rippled 1.11.0 or later.
	// Only available in: NFTokenCancelOfferMetadata
	NFTokenIDs []types.NFTokenID `json:"nftoken_ids,omitempty"`

	// Only available in: MPTokenIssuanceCreate
	MPTIssuanceID *types.MPTIssuanceID `json:"mpt_issuance_id,omitempty"`
}

// AsPaymentMetadata returns the PaymentMetadata.
func (tmb TxMetadataBuilder) AsPaymentMetadata() PaymentMetadata {
	return PaymentMetadata{
		TxObjMeta: tmb.AsTxObjMeta(),
	}
}

// AsNFTokenMintMetadata returns the AsNFTokenMintMetadata.
func (tmb TxMetadataBuilder) AsNFTokenMintMetadata() NFTokenMintMetadata {
	return NFTokenMintMetadata{
		TxObjMeta: tmb.AsTxObjMeta(),
		NFTokenID: tmb.NFTokenID,
		OfferID:   tmb.OfferID,
	}
}

// AsNFTokenCreateOfferMetadata returns the NFTokenCreateOfferMetadata.
func (tmb TxMetadataBuilder) AsNFTokenCreateOfferMetadata() NFTokenCreateOfferMetadata {
	return NFTokenCreateOfferMetadata{
		TxObjMeta: tmb.AsTxObjMeta(),
		OfferID:   tmb.OfferID,
	}
}

// AsNFTokenAcceptOfferMetadata returns the NFTokenAcceptOfferMetadata.
func (tmb TxMetadataBuilder) AsNFTokenAcceptOfferMetadata() NFTokenAcceptOfferMetadata {
	return NFTokenAcceptOfferMetadata{
		TxObjMeta: tmb.AsTxObjMeta(),
		NFTokenID: tmb.NFTokenID,
	}
}

// AsNFTokenCancelOfferMetadata returns the NFTokenCancelOfferMetadata.
func (tmb TxMetadataBuilder) AsNFTokenCancelOfferMetadata() NFTokenCancelOfferMetadata {
	return NFTokenCancelOfferMetadata{
		TxObjMeta:  tmb.AsTxObjMeta(),
		NFTokenIDs: tmb.NFTokenIDs,
	}
}

// AsMPTokenIssuanceCreateMetadata returns the MPTokenIssuanceCreateMetadata.
func (tmb TxMetadataBuilder) AsMPTokenIssuanceCreateMetadata() MPTokenIssuanceCreateMetadata {
	return MPTokenIssuanceCreateMetadata{
		TxObjMeta:     tmb.AsTxObjMeta(),
		MPTIssuanceID: tmb.MPTIssuanceID,
	}
}

// AsTxObjMeta returns the base TxObjMeta metadata.
func (tmb TxMetadataBuilder) AsTxObjMeta() TxObjMeta {
	return TxObjMeta{
		AffectedNodes:          tmb.AffectedNodes,
		PartialDeliveredAmount: tmb.PartialDeliveredAmount,
		TransactionIndex:       tmb.TransactionIndex,
		TransactionResult:      tmb.TransactionResult,
		DeliveredAmount:        tmb.DeliveredAmount,
		ParentBatchID:          tmb.ParentBatchID,
	}
}

package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNFTokenMint_TxType(t *testing.T) {
	tx := &NFTokenMint{}
	assert.Equal(t, NFTokenMintTx, tx.TxType())
}

func TestNFTokenMint_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*NFTokenMint)
		expected uint32
	}{
		{
			name: "pass - SetBurnableFlag",
			setter: func(n *NFTokenMint) {
				n.SetBurnableFlag()
			},
			expected: tfBurnable,
		},
		{
			name: "pass - SetOnlyXRPFlag",
			setter: func(n *NFTokenMint) {
				n.SetOnlyXRPFlag()
			},
			expected: tfOnlyXRP,
		},
		{
			name: "pass - SetTrustlineFlag",
			setter: func(n *NFTokenMint) {
				n.SetTrustlineFlag()
			},
			expected: tfTrustLine,
		},
		{
			name: "pass - SetTransferableFlag",
			setter: func(n *NFTokenMint) {
				n.SetTransferableFlag()
			},
			expected: tfTransferable,
		},
		{
			name: "pass - SetMutableFlag",
			setter: func(n *NFTokenMint) {
				n.SetMutableFlag()
			},
			expected: tfMutable,
		},
		{
			name: "pass - SetBurnableFlag and SetTransferableFlag",
			setter: func(n *NFTokenMint) {
				n.SetBurnableFlag()
				n.SetTransferableFlag()
			},
			expected: tfBurnable | tfTransferable,
		},
		{
			name: "pass - SetBurnableFlag and SetTransferableFlag and SetOnlyXRPFlag",
			setter: func(n *NFTokenMint) {
				n.SetBurnableFlag()
				n.SetTransferableFlag()
				n.SetOnlyXRPFlag()
			},
			expected: tfBurnable | tfTransferable | tfOnlyXRP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &NFTokenMint{}
			tt.setter(p)
			if p.Flags != tt.expected {
				t.Errorf("Expected Flags to be %d, got %d", tt.expected, p.Flags)
			}
		})
	}
}

func TestNFTokenMint_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		nft      *NFTokenMint
		expected FlatTransaction
	}{
		{
			name: "pass - Flatten with all fields",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				Issuer:       "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				TransferFee:  types.TransferFee(314),
				URI:          "697066733A2F2F62616679626569676479727A74357366703775646D37687537367568377932366E6634646675796C71616266336F636C67747179353566627A6469",
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   "r3Q1i8Y2e5v4Z2u7eFYTEXSwuJYfV2Jpn",
					Value:    "1000",
				},
				Expiration:  types.Expiration(1234567890),
				Destination: "rM8JHG9dzYuPxHEir2qAi998Vsnr3jccUw",
			},
			expected: FlatTransaction{
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "NFTokenMint",
				"Fee":             "10",
				"NFTokenTaxon":    uint32(12345),
				"Issuer":          "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				"TransferFee":     uint16(314),
				"URI":             "697066733A2F2F62616679626569676479727A74357366703775646D37687537367568377932366E6634646675796C71616266336F636C67747179353566627A6469",
				"Amount": map[string]interface{}{
					"currency": "USD",
					"issuer":   "r3Q1i8Y2e5v4Z2u7eFYTEXSwuJYfV2Jpn",
					"value":    "1000",
				},
				"Expiration":  uint32(1234567890),
				"Destination": "rM8JHG9dzYuPxHEir2qAi998Vsnr3jccUw",
			},
		},
		{
			name: "pass - Flatten with minimal fields",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
			},
			expected: FlatTransaction{
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "NFTokenMint",
				"Fee":             "10",
				"NFTokenTaxon":    uint32(12345),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.nft.Flatten())
		})
	}
}

func TestNFTokenMint_Validate(t *testing.T) {
	tests := []struct {
		name       string
		nft        *NFTokenMint
		setter     func(*NFTokenMint)
		wantValid  bool
		wantErr    bool
		errMessage error
	}{
		{
			name: "pass - minimal fields",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid BaseTx fields, missing account",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidAccount,
		},
		{
			name: "fail - transfer fee exceeds max",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				TransferFee:  types.TransferFee(60000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidTransferFee,
		},
		{
			name: "fail - issuer same as account",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				Issuer:       "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrIssuerAccountConflict,
		},
		{
			name: "fail - issuer invalid address",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				Issuer:       "invalidAddress",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidIssuer,
		},
		{
			name: "fail - URI not hexadecimal",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				URI:          "invalidURI",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidURI,
		},
		{
			name: "fail - transfer fee set without transferable flag",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				TransferFee:  types.TransferFee(314),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrTransferFeeRequiresTransferableFlag,
		},
		{
			name: "pass - transfer fee set with transferable flag",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				TransferFee:  types.TransferFee(314),
			},
			setter: func(n *NFTokenMint) {
				n.SetTransferableFlag()
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - all fields",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				TransferFee:  types.TransferFee(314),
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   "rbBGwDkFSkTknJ4GA9nhaJdoDwWqSTpLE",
					Value:    "1000",
				},
				Expiration:  types.Expiration(1234567890),
				Destination: "rM8JHG9dzYuPxHEir2qAi998Vsnr3jccUw",
			},
			setter: func(n *NFTokenMint) {
				n.SetTransferableFlag()
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid Destination",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				TransferFee:  types.TransferFee(314),
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   "rbBGwDkFSkTknJ4GA9nhaJdoDwWqSTpLE",
					Value:    "1000",
				},
				Expiration:  types.Expiration(1234567890),
				Destination: "invalid",
			},
			setter: func(n *NFTokenMint) {
				n.SetTransferableFlag()
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidDestination,
		},
		{
			name: "fail - missing Amount when Expiration or Destination is set",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				Expiration:   types.Expiration(1234567890),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrAmountRequiredWithExpirationOrDestination,
		},
		{
			name: "fail - invalid issuer Amount",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				TransferFee:  types.TransferFee(314),
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   "invalid",
					Value:    "1000",
				},
			},
			setter: func(n *NFTokenMint) {
				n.SetTransferableFlag()
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidIssuer,
		},
		{
			name: "pass - valid Amount with XRP",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				Amount:       types.XRPCurrencyAmount(1000000),
				Destination:  "rM8JHG9dzYuPxHEir2qAi998Vsnr3jccUw",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid Amount with IssuedCurrency",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   "rbBGwDkFSkTknJ4GA9nhaJdoDwWqSTpLE",
					Value:    "100.50",
				},
				Expiration: types.Expiration(1234567890),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid Amount with empty values",
			nft: &NFTokenMint{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: NFTokenMintTx,
					Fee:             types.XRPCurrencyAmount(10),
				},
				NFTokenTaxon: 12345,
				Amount: types.IssuedCurrencyAmount{
					Currency: "",
					Issuer:   "",
					Value:    "",
				},
				Expiration: types.Expiration(1234567890),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidTokenFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setter != nil {
				tt.setter(tt.nft)
			}
			valid, err := tt.nft.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && err != tt.errMessage {
				t.Errorf("Validate() got error message = %v, want error message %v", err, tt.errMessage)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, wantValid %v", valid, tt.wantValid)
			}
		})
	}
}

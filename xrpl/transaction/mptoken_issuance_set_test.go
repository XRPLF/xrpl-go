package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const validMPTIssuanceID = "00070C4495F14B0E44F78A264E41713C64B5F89242540EE2"

// validCompressedKey is the compressed secp256k1 generator point.
const validCompressedKey = "0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"

func validMPTokenIssuanceSet() *MPTokenIssuanceSet {
	return &MPTokenIssuanceSet{
		BaseTx: BaseTx{
			Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			TransactionType: MPTokenIssuanceSetTx,
		},
		MPTokenIssuanceID: validMPTIssuanceID,
		MutableFlags:      types.MutableFlags(TmfMPTSetCanLock),
	}
}

func TestMPTokenIssuanceSet_TxType(t *testing.T) {
	require.Equal(t, MPTokenIssuanceSetTx, (&MPTokenIssuanceSet{}).TxType())
}

func TestMPTokenIssuanceSet_FlattenAndJSON(t *testing.T) {
	tx := validMPTokenIssuanceSet()
	tx.MutableFlags = types.MutableFlags(TmfMPTSetCanLock | TmfMPTSetCanHoldConfidentialBalance)
	tx.MPTokenMetadata = types.MPTokenMetadata("464f4f")
	tx.TransferFee = types.TransferFee(0)
	tx.DomainID = types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9")
	tx.IssuerEncryptionKey = types.EncryptionKey(validCompressedKey)
	tx.AuditorEncryptionKey = types.EncryptionKey(validCompressedKey)

	require.Equal(t, FlatTransaction{
		"Account":              "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		"TransactionType":      "MPTokenIssuanceSet",
		"MPTokenIssuanceID":    validMPTIssuanceID,
		"DomainID":             "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
		"MPTokenMetadata":      "464f4f",
		"TransferFee":          uint16(0),
		"MutableFlags":         uint32(0x41),
		"IssuerEncryptionKey":  validCompressedKey,
		"AuditorEncryptionKey": validCompressedKey,
	}, tx.Flatten())

	holderTx := &MPTokenIssuanceSet{
		BaseTx: BaseTx{
			Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			TransactionType: MPTokenIssuanceSetTx,
			Flags:           TfMPTLock,
		},
		MPTokenIssuanceID: validMPTIssuanceID,
		Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
	}
	require.Equal(t, FlatTransaction{
		"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		"TransactionType":   "MPTokenIssuanceSet",
		"Flags":             uint32(0x01),
		"MPTokenIssuanceID": validMPTIssuanceID,
		"Holder":            "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
	}, holderTx.Flatten())

	encoded, err := json.Marshal(tx)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"Account":"rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		"TransactionType":"MPTokenIssuanceSet",
		"MPTokenIssuanceID":"00070C4495F14B0E44F78A264E41713C64B5F89242540EE2",
		"Holder":null,
		"DomainID":"A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
		"MPTokenMetadata":"464f4f",
		"TransferFee":0,
		"MutableFlags":65,
		"IssuerEncryptionKey":"`+validCompressedKey+`",
		"AuditorEncryptionKey":"`+validCompressedKey+`"
	}`, string(encoded))
}

func TestMPTokenIssuanceSetMutableFlagValues(t *testing.T) {
	require.Equal(t, uint32(0x01), TmfMPTSetCanLock)
	require.Equal(t, uint32(0x02), TmfMPTSetRequireAuth)
	require.Equal(t, uint32(0x04), TmfMPTSetCanEscrow)
	require.Equal(t, uint32(0x08), TmfMPTSetCanTrade)
	require.Equal(t, uint32(0x10), TmfMPTSetCanTransfer)
	require.Equal(t, uint32(0x20), TmfMPTSetCanClawback)
	require.Equal(t, uint32(0x40), TmfMPTSetCanHoldConfidentialBalance)
	require.Equal(t, uint32(0x7F), MPTokenIssuanceSetMutableFlagsMask)
}

func TestMPTokenIssuanceSet_MutableFlagSettersPreserveBits(t *testing.T) {
	tests := []struct {
		name    string
		set     func(*MPTokenIssuanceSet)
		literal uint32
	}{
		{"SetCanLock", (*MPTokenIssuanceSet).SetMPTSetCanLockMutableFlag, 0x01},
		{"SetRequireAuth", (*MPTokenIssuanceSet).SetMPTSetRequireAuthMutableFlag, 0x02},
		{"SetCanEscrow", (*MPTokenIssuanceSet).SetMPTSetCanEscrowMutableFlag, 0x04},
		{"SetCanTrade", (*MPTokenIssuanceSet).SetMPTSetCanTradeMutableFlag, 0x08},
		{"SetCanTransfer", (*MPTokenIssuanceSet).SetMPTSetCanTransferMutableFlag, 0x10},
		{"SetCanClawback", (*MPTokenIssuanceSet).SetMPTSetCanClawbackMutableFlag, 0x20},
		{"SetCanHoldConfidentialBalance", (*MPTokenIssuanceSet).SetMPTSetCanHoldConfidentialBalanceMutableFlag, 0x40},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &MPTokenIssuanceSet{MutableFlags: types.MutableFlags(0x80000000)}
			test.set(tx)
			require.Equal(t, uint32(0x80000000)|test.literal, *tx.MutableFlags)
		})
	}

	tx := &MPTokenIssuanceSet{}
	for _, test := range tests {
		test.set(tx)
	}
	require.Equal(t, uint32(0x7F), *tx.MutableFlags)
}

func TestMPTokenIssuanceSet_ValidateMutableFlags(t *testing.T) {
	tests := []struct {
		name  string
		flags uint32
		ok    bool
	}{
		{"one valid bit", TmfMPTSetCanTrade, true},
		{"all valid bits", MPTokenIssuanceSetMutableFlagsMask, true},
		{"present zero", 0, false},
		{"removed bit 0x80", 0x80, false},
		{"removed legacy clear bit", 0x2000, false},
		{"unknown high bit", 0x80000000, false},
		{"valid and unknown", TmfMPTSetCanLock | 0x100, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			tx.MutableFlags = types.MutableFlags(test.flags)
			ok, err := tx.Validate()
			require.Equal(t, test.ok, ok)
			if test.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrMPTIssuanceSetInvalidMutableFlags)
			}
		})
	}
}

func TestMPTokenIssuanceSet_TransferFeeConfidentialBalance(t *testing.T) {
	tests := []struct {
		name  string
		fee   *uint16
		flags uint32
		ok    bool
	}{
		{"non-zero fee with confidential", types.TransferFee(1), TmfMPTSetCanHoldConfidentialBalance, false},
		{"non-zero fee with confidential and another bit", types.TransferFee(1), TmfMPTSetCanLock | TmfMPTSetCanHoldConfidentialBalance, false},
		{"zero fee with confidential", types.TransferFee(0), TmfMPTSetCanHoldConfidentialBalance, true},
		{"absent fee with confidential", nil, TmfMPTSetCanHoldConfidentialBalance, true},
		{"non-zero fee without confidential", types.TransferFee(1), TmfMPTSetCanTransfer, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			tx.TransferFee = test.fee
			tx.MutableFlags = types.MutableFlags(test.flags)
			ok, err := tx.Validate()
			require.Equal(t, test.ok, ok)
			if test.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrMPTIssuanceSetTransferFeeWithConfidentialBalance)
			}
		})
	}
}

func TestMPTokenIssuanceSet_ValidateIssuanceIDAndHolder(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MPTokenIssuanceSet)
		err    error
	}{
		{"empty issuance ID", func(tx *MPTokenIssuanceSet) { tx.MPTokenIssuanceID = "" }, ErrInvalidMPTokenIssuanceIDSet},
		{"non-hex 48-character issuance ID", func(tx *MPTokenIssuanceSet) {
			tx.MPTokenIssuanceID = "00070C4495F14B0E44F78A264E41713C64B5F89242540EG2"
		}, ErrInvalidMPTokenIssuanceIDSet},
		{"46-character issuance ID", func(tx *MPTokenIssuanceSet) { tx.MPTokenIssuanceID = "00070C4495F14B0E44F78A264E41713C64B5F89242540E" }, ErrInvalidMPTokenIssuanceIDSet},
		{"50-character issuance ID", func(tx *MPTokenIssuanceSet) {
			tx.MPTokenIssuanceID = "00070C4495F14B0E44F78A264E41713C64B5F89242540EE200"
		}, ErrInvalidMPTokenIssuanceIDSet},
		{"oversized issuance ID", func(tx *MPTokenIssuanceSet) {
			tx.MPTokenIssuanceID = "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000"
		}, ErrInvalidMPTokenIssuanceIDSet},
		{"invalid holder", func(tx *MPTokenIssuanceSet) {
			tx.MutableFlags = nil
			tx.Holder = types.Holder("invalid")
		}, ErrInvalidAccount},
		{"holder equals account", func(tx *MPTokenIssuanceSet) {
			tx.MutableFlags = nil
			tx.Holder = types.Holder("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD")
		}, ErrHolderAccountConflict},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			test.mutate(tx)
			ok, err := tx.Validate()
			require.False(t, ok)
			require.ErrorIs(t, err, test.err)
		})
	}

	t.Run("valid standalone holder", func(t *testing.T) {
		tx := validMPTokenIssuanceSet()
		tx.MutableFlags = nil
		tx.Holder = types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2")
		ok, err := tx.Validate()
		require.True(t, ok)
		require.NoError(t, err)
	})
}

func TestMPTokenIssuanceSet_ValidateMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		ok       bool
	}{
		{"exactly 1024 bytes", strings.Repeat("AB", 1024), true},
		{"over 1024 bytes", strings.Repeat("AB", 1025), false},
		{"invalid hex", "not-hex", false},
		{"empty removes metadata", "", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			tx.MutableFlags = nil
			tx.MPTokenMetadata = types.MPTokenMetadata(test.metadata)
			ok, err := tx.Validate()
			require.Equal(t, test.ok, ok)
			if test.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrInvalidMPTokenMetadata)
			}
		})
	}
}

func TestMPTokenIssuanceSet_ValidateEncryptionKeys(t *testing.T) {
	invalidSyntax := "02" + strings.Repeat("GG", 32)
	tests := []struct {
		name   string
		mutate func(*MPTokenIssuanceSet)
		err    error
	}{
		{"issuer invalid syntax", func(tx *MPTokenIssuanceSet) {
			tx.IssuerEncryptionKey = types.EncryptionKey(invalidSyntax)
		}, ErrMPTIssuanceSetInvalidKeyLength},
		{"issuer invalid length", func(tx *MPTokenIssuanceSet) {
			tx.IssuerEncryptionKey = types.EncryptionKey("02AABB")
		}, ErrMPTIssuanceSetInvalidKeyLength},
		{"auditor invalid syntax", func(tx *MPTokenIssuanceSet) {
			tx.IssuerEncryptionKey = types.EncryptionKey(validCompressedKey)
			tx.AuditorEncryptionKey = types.EncryptionKey(invalidSyntax)
		}, ErrMPTIssuanceSetInvalidKeyLength},
		{"auditor invalid length", func(tx *MPTokenIssuanceSet) {
			tx.IssuerEncryptionKey = types.EncryptionKey(validCompressedKey)
			tx.AuditorEncryptionKey = types.EncryptionKey("03AABB")
		}, ErrMPTIssuanceSetInvalidKeyLength},
		{"auditor without issuer", func(tx *MPTokenIssuanceSet) {
			tx.AuditorEncryptionKey = types.EncryptionKey(validCompressedKey)
		}, ErrMPTIssuanceSetAuditorRequiresIssuerKey},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			tx.MutableFlags = nil
			test.mutate(tx)
			ok, err := tx.Validate()
			require.False(t, ok)
			require.ErrorIs(t, err, test.err)
		})
	}

	for _, test := range []struct {
		name   string
		mutate func(*MPTokenIssuanceSet)
	}{
		{"issuer key", func(tx *MPTokenIssuanceSet) {
			tx.IssuerEncryptionKey = types.EncryptionKey(validCompressedKey)
		}},
		{"paired issuer and auditor keys", func(tx *MPTokenIssuanceSet) {
			tx.IssuerEncryptionKey = types.EncryptionKey(validCompressedKey)
			tx.AuditorEncryptionKey = types.EncryptionKey(validCompressedKey)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			tx.MutableFlags = nil
			test.mutate(tx)
			ok, err := tx.Validate()
			require.True(t, ok)
			require.NoError(t, err)
		})
	}
}

func TestMPTokenIssuanceSet_ValidateOperationConflictsAndBounds(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MPTokenIssuanceSet)
		err    error
	}{
		{"no operation", func(tx *MPTokenIssuanceSet) { tx.MutableFlags = nil }, ErrMPTIssuanceSetEmpty},
		{"conflicting lock flags", func(tx *MPTokenIssuanceSet) { tx.MutableFlags = nil; tx.Flags = TfMPTLock | TfMPTUnlock }, ErrMPTokenIssuanceSetFlags},
		{"flags with mutation", func(tx *MPTokenIssuanceSet) { tx.Flags = TfMPTLock }, ErrMPTIssuanceSetFlagsMutuallyExclusive},
		{"holder with mutation", func(tx *MPTokenIssuanceSet) { tx.Holder = types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2") }, ErrMPTIssuanceSetHolderMutuallyExclusive},
		{"holder with key", func(tx *MPTokenIssuanceSet) {
			tx.MutableFlags = nil
			tx.Holder = types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2")
			tx.IssuerEncryptionKey = types.EncryptionKey(validCompressedKey)
		}, ErrMPTIssuanceSetKeyConflict},
		{"fee too high", func(tx *MPTokenIssuanceSet) { tx.TransferFee = types.TransferFee(50001) }, ErrInvalidTransferFee},
		{"invalid domain", func(tx *MPTokenIssuanceSet) { tx.DomainID = types.DomainID("not-hex") }, ErrMPTIssuanceSetDomainIDInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			test.mutate(tx)
			ok, err := tx.Validate()
			require.False(t, ok)
			require.ErrorIs(t, err, test.err)
		})
	}
}

func TestMPTokenIssuanceSet_ZeroValuesAllowed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MPTokenIssuanceSet)
	}{
		{"empty domain removes", func(tx *MPTokenIssuanceSet) { tx.DomainID = types.DomainID("") }},
		{"zero transfer fee", func(tx *MPTokenIssuanceSet) { tx.TransferFee = types.TransferFee(0) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := validMPTokenIssuanceSet()
			tx.MutableFlags = nil
			test.mutate(tx)
			ok, err := tx.Validate()
			require.True(t, ok)
			require.NoError(t, err)
		})
	}
}

func TestMPTokenIssuanceSet_FlagValuesAndSettersPreserveBits(t *testing.T) {
	require.Equal(t, uint32(0x01), TfMPTLock)
	require.Equal(t, uint32(0x02), TfMPTUnlock)

	tests := []struct {
		name    string
		set     func(*MPTokenIssuanceSet)
		literal uint32
	}{
		{"lock", (*MPTokenIssuanceSet).SetMPTLockFlag, 0x01},
		{"unlock", (*MPTokenIssuanceSet).SetMPTUnlockFlag, 0x02},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &MPTokenIssuanceSet{BaseTx: BaseTx{Flags: 0x100}}
			test.set(tx)
			require.Equal(t, uint32(0x100)|test.literal, tx.Flags)
		})
	}
}

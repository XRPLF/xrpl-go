package transaction

import (
	"math"
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	preflightHolder = types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD")
	preflightIssuer = types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
	preflightOther  = types.Address("rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP")
	preflightID     = "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8"
)

type confidentialMPTValidator interface {
	Validate() (bool, error)
}

func preflightBase(account types.Address, txType TxType) BaseTx {
	return BaseTx{Account: account, TransactionType: txType, Fee: 12}
}

// The validPreflight* constructors each return a transaction that validates as-is, so a
// case that mutates one field proves that field is the only reason for the rejection.

func validPreflightConvert() *ConfidentialMPTConvert {
	return &ConfidentialMPTConvert{
		BaseTx:                preflightBase(preflightHolder, ConfidentialMPTConvertTx),
		MPTokenIssuanceID:     preflightID,
		HolderEncryptedAmount: testCiphertext,
		IssuerEncryptedAmount: testCiphertext2,
		BlindingFactor:        strings.Repeat("12", BlindingFactorLen/2),
	}
}

func validPreflightMergeInbox() *ConfidentialMPTMergeInbox {
	return &ConfidentialMPTMergeInbox{
		BaseTx:            preflightBase(preflightHolder, ConfidentialMPTMergeInboxTx),
		MPTokenIssuanceID: preflightID,
	}
}

func validPreflightConvertBack() *ConfidentialMPTConvertBack {
	return &ConfidentialMPTConvertBack{
		BaseTx:                preflightBase(preflightHolder, ConfidentialMPTConvertBackTx),
		MPTokenIssuanceID:     preflightID,
		MPTAmount:             1,
		HolderEncryptedAmount: testCiphertext,
		IssuerEncryptedAmount: testCiphertext2,
		BlindingFactor:        strings.Repeat("12", BlindingFactorLen/2),
		BalanceCommitment:     testCompressedPoint1,
		ZKProof:               strings.Repeat("11", ConvertBackProofLen/2),
	}
}

func validPreflightSend() *ConfidentialMPTSend {
	return &ConfidentialMPTSend{
		BaseTx:                     preflightBase(preflightHolder, ConfidentialMPTSendTx),
		MPTokenIssuanceID:          preflightID,
		Destination:                preflightOther,
		SenderEncryptedAmount:      testCiphertext,
		DestinationEncryptedAmount: testCiphertext2,
		IssuerEncryptedAmount:      testCiphertext3,
		ZKProof:                    strings.Repeat("11", SendProofLen/2),
		AmountCommitment:           testCompressedPoint1,
		BalanceCommitment:          testCompressedPoint3,
	}
}

func validPreflightClawback() *ConfidentialMPTClawback {
	return &ConfidentialMPTClawback{
		BaseTx:            preflightBase(preflightIssuer, ConfidentialMPTClawbackTx),
		MPTokenIssuanceID: preflightID,
		Holder:            preflightHolder,
		MPTAmount:         1,
		ZKProof:           strings.Repeat("11", ClawbackProofLen/2),
	}
}

// preflightCase pairs a fixture with a pointer to its BaseTx, so a case can mutate the
// shared fields without knowing the concrete transaction type.
type preflightCase struct {
	name string
	tx   confidentialMPTValidator
	base *BaseTx
}

// validPreflightCases returns one valid transaction of each confidential MPT type.
func validPreflightCases() []preflightCase {
	convert := validPreflightConvert()
	mergeInbox := validPreflightMergeInbox()
	convertBack := validPreflightConvertBack()
	send := validPreflightSend()
	clawback := validPreflightClawback()
	return []preflightCase{
		{name: "convert", tx: convert, base: &convert.BaseTx},
		{name: "merge inbox", tx: mergeInbox, base: &mergeInbox.BaseTx},
		{name: "convert back", tx: convertBack, base: &convertBack.BaseTx},
		{name: "send", tx: send, base: &send.BaseTx},
		{name: "clawback", tx: clawback, base: &clawback.BaseTx},
	}
}

func TestConfidentialMPTBasePreflight(t *testing.T) {
	for _, test := range validPreflightCases() {
		t.Run(test.name+" rejects transaction-specific flags", func(t *testing.T) {
			test.base.Flags = 1

			valid, err := test.tx.Validate()

			require.False(t, valid)
			require.ErrorIs(t, err, ErrConfidentialMPTInvalidFlags)
		})
	}

	t.Run("universal flags are allowed", func(t *testing.T) {
		tx := validPreflightConvert()
		tx.Flags = types.TfUniversal

		valid, err := tx.Validate()

		require.NoError(t, err)
		require.True(t, valid)
	})
}

// TestConfidentialMPTAcceptDelegate pins that Delegate is a valid BaseTx field on every
// confidential MPT type. Whether a type may actually be delegated is enforced by
// DelegateSet.Validate, not here. See TestConfidentialMPTConvertIsNotDelegatable.
func TestConfidentialMPTAcceptDelegate(t *testing.T) {
	for _, test := range validPreflightCases() {
		t.Run(test.name, func(t *testing.T) {
			test.base.Delegate = preflightOther

			valid, err := test.tx.Validate()

			require.NoError(t, err)
			require.True(t, valid)
		})
	}
}

// TestConfidentialMPTConvertIsNotDelegatable pins the one confidential MPT type rippled
// marks Delegation::NotDelegable, so DelegateSet rejects a permission grant that the
// network would refuse. The other four are Delegable.
func TestConfidentialMPTConvertIsNotDelegatable(t *testing.T) {
	_, nonDelegatable := NonDelegatableTransactionsMap[ConfidentialMPTConvertTx.String()]
	require.True(t, nonDelegatable, "ConfidentialMPTConvert must be non-delegatable")

	for _, txType := range []TxType{
		ConfidentialMPTSendTx,
		ConfidentialMPTConvertBackTx,
		ConfidentialMPTMergeInboxTx,
		ConfidentialMPTClawbackTx,
	} {
		t.Run(txType.String(), func(t *testing.T) {
			_, found := NonDelegatableTransactionsMap[txType.String()]
			require.False(t, found, "%s must remain delegatable", txType)
		})
	}
}

func TestConfidentialMPTIssuanceIDPreflight(t *testing.T) {
	const invalidID = "ABCD"

	convert := validPreflightConvert()
	convert.MPTokenIssuanceID = invalidID
	mergeInbox := validPreflightMergeInbox()
	mergeInbox.MPTokenIssuanceID = invalidID
	convertBack := validPreflightConvertBack()
	convertBack.MPTokenIssuanceID = invalidID
	send := validPreflightSend()
	send.MPTokenIssuanceID = invalidID
	clawback := validPreflightClawback()
	clawback.MPTokenIssuanceID = invalidID

	tests := []struct {
		name string
		tx   confidentialMPTValidator
	}{
		{name: "convert", tx: convert},
		{name: "merge inbox", tx: mergeInbox},
		{name: "convert back", tx: convertBack},
		{name: "send", tx: send},
		{name: "clawback", tx: clawback},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, err := test.tx.Validate()

			require.False(t, valid)
			require.ErrorIs(t, err, ErrConfidentialMPTInvalidIssuanceID)
		})
	}
}

func TestConfidentialMPTIssuerRolePreflight(t *testing.T) {
	issuerXAddress, err := addresscodec.ClassicAddressToXAddress(preflightIssuer.String(), 123, true, false)
	require.NoError(t, err)

	convert := validPreflightConvert()
	convert.Account = preflightIssuer
	mergeInbox := validPreflightMergeInbox()
	mergeInbox.Account = preflightIssuer
	convertBack := validPreflightConvertBack()
	convertBack.Account = preflightIssuer
	sendIssuerAccount := validPreflightSend()
	sendIssuerAccount.Account = preflightIssuer
	sendClassicIssuerDest := validPreflightSend()
	sendClassicIssuerDest.Destination = preflightIssuer
	sendXAddressIssuerDest := validPreflightSend()
	sendXAddressIssuerDest.Destination = types.Address(issuerXAddress)
	clawback := validPreflightClawback()
	clawback.Account = preflightOther

	tests := []struct {
		name    string
		tx      confidentialMPTValidator
		wantErr error
	}{
		{name: "convert rejects issuer", tx: convert, wantErr: ErrConfidentialMPTIssuerNotAllowed},
		{name: "merge inbox rejects issuer", tx: mergeInbox, wantErr: ErrConfidentialMPTIssuerNotAllowed},
		{name: "convert back rejects issuer", tx: convertBack, wantErr: ErrConfidentialMPTIssuerNotAllowed},
		{name: "send rejects issuer account", tx: sendIssuerAccount, wantErr: ErrConfidentialMPTIssuerNotAllowed},
		{name: "send rejects classic issuer destination", tx: sendClassicIssuerDest, wantErr: ErrConfidentialSendDestinationIsIssuer},
		{name: "send rejects X-address issuer destination", tx: sendXAddressIssuerDest, wantErr: ErrConfidentialSendDestinationIsIssuer},
		{name: "clawback requires issuer", tx: clawback, wantErr: ErrConfidentialMPTIssuerRequired},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, err := test.tx.Validate()

			require.False(t, valid)
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestConfidentialMPTAmountPreflight(t *testing.T) {
	// XLS-96 makes a zero-amount convert the opt-in path, so set MPTAmount
	// explicitly rather than relying on the fixture's zero value.
	t.Run("convert permits zero", func(t *testing.T) {
		tx := validPreflightConvert()
		tx.MPTAmount = 0

		valid, err := tx.Validate()

		require.NoError(t, err)
		require.True(t, valid)
	})

	tooLarge := types.MPTPlainAmount(uint64(math.MaxInt64) + 1)

	convertTooLarge := validPreflightConvert()
	convertTooLarge.MPTAmount = tooLarge
	convertBackZero := validPreflightConvertBack()
	convertBackZero.MPTAmount = 0
	convertBackTooLarge := validPreflightConvertBack()
	convertBackTooLarge.MPTAmount = tooLarge
	clawbackZero := validPreflightClawback()
	clawbackZero.MPTAmount = 0
	clawbackTooLarge := validPreflightClawback()
	clawbackTooLarge.MPTAmount = tooLarge

	tests := []struct {
		name    string
		tx      confidentialMPTValidator
		wantErr error
	}{
		{name: "convert rejects amount above maximum", tx: convertTooLarge, wantErr: ErrConfidentialMPTInvalidAmount},
		{name: "convert back rejects zero", tx: convertBackZero, wantErr: ErrConfidentialConvertBackInvalidAmount},
		{name: "convert back rejects amount above maximum", tx: convertBackTooLarge, wantErr: ErrConfidentialConvertBackInvalidAmount},
		{name: "clawback rejects zero", tx: clawbackZero, wantErr: ErrConfidentialClawbackInvalidAmount},
		{name: "clawback rejects amount above maximum", tx: clawbackTooLarge, wantErr: ErrConfidentialClawbackInvalidAmount},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, err := test.tx.Validate()

			require.False(t, valid)
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestConfidentialMPTFlattenIssuanceIDIsString(t *testing.T) {
	flattened := validPreflightConvert().Flatten()
	require.IsType(t, "", flattened["MPTokenIssuanceID"])
	require.Equal(t, preflightID, flattened["MPTokenIssuanceID"])
}

package transaction

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFlattenTransactionType protects the concrete type and the plain string map value.
func TestFlattenTransactionType(t *testing.T) {
	type flatTx interface {
		TxType() TxType
		Flatten() FlatTransaction
	}
	constructors := []func(BaseTx) flatTx{
		func(base BaseTx) flatTx { return &AMMBid{BaseTx: base} },
		func(base BaseTx) flatTx { return &AMMClawback{BaseTx: base} },
		func(base BaseTx) flatTx { return &AMMCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &AMMDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &AMMDeposit{BaseTx: base} },
		func(base BaseTx) flatTx { return &AMMVote{BaseTx: base} },
		func(base BaseTx) flatTx { return &AMMWithdraw{BaseTx: base} },
		func(base BaseTx) flatTx { return &AccountDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &AccountSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &Batch{BaseTx: base} },
		func(base BaseTx) flatTx { return &CheckCancel{BaseTx: base} },
		func(base BaseTx) flatTx { return &CheckCash{BaseTx: base} },
		func(base BaseTx) flatTx { return &CheckCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &Clawback{BaseTx: base} },
		func(base BaseTx) flatTx { return &ConfidentialMPTClawback{BaseTx: base} },
		func(base BaseTx) flatTx { return &ConfidentialMPTConvert{BaseTx: base} },
		func(base BaseTx) flatTx { return &ConfidentialMPTConvertBack{BaseTx: base} },
		func(base BaseTx) flatTx { return &ConfidentialMPTMergeInbox{BaseTx: base} },
		func(base BaseTx) flatTx { return &ConfidentialMPTSend{BaseTx: base} },
		func(base BaseTx) flatTx { return &CredentialAccept{BaseTx: base} },
		func(base BaseTx) flatTx { return &CredentialCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &CredentialDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &DIDDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &DIDSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &DelegateSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &DepositPreauth{BaseTx: base} },
		func(base BaseTx) flatTx { return &EscrowCancel{BaseTx: base} },
		func(base BaseTx) flatTx { return &EscrowCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &EscrowFinish{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanBrokerCoverClawback{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanBrokerCoverDeposit{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanBrokerCoverWithdraw{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanBrokerDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanBrokerSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanManage{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanPay{BaseTx: base} },
		func(base BaseTx) flatTx { return &LoanSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &MPTokenAuthorize{BaseTx: base} },
		func(base BaseTx) flatTx { return &MPTokenIssuanceCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &MPTokenIssuanceDestroy{BaseTx: base} },
		func(base BaseTx) flatTx { return &MPTokenIssuanceSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &NFTokenAcceptOffer{BaseTx: base} },
		func(base BaseTx) flatTx { return &NFTokenBurn{BaseTx: base} },
		func(base BaseTx) flatTx { return &NFTokenCancelOffer{BaseTx: base} },
		func(base BaseTx) flatTx { return &NFTokenCreateOffer{BaseTx: base} },
		func(base BaseTx) flatTx { return &NFTokenMint{BaseTx: base} },
		func(base BaseTx) flatTx { return &NFTokenModify{BaseTx: base} },
		func(base BaseTx) flatTx { return &OfferCancel{BaseTx: base} },
		func(base BaseTx) flatTx { return &OfferCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &OracleDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &OracleSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &Payment{BaseTx: base} },
		func(base BaseTx) flatTx { return &PaymentChannelClaim{BaseTx: base} },
		func(base BaseTx) flatTx { return &PaymentChannelCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &PaymentChannelFund{BaseTx: base} },
		func(base BaseTx) flatTx { return &PermissionedDomainDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &PermissionedDomainSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &SetRegularKey{BaseTx: base} },
		func(base BaseTx) flatTx { return &SignerListSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &SponsorshipSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &SponsorshipTransfer{BaseTx: base} },
		func(base BaseTx) flatTx { return &TicketCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &TrustSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &VaultClawback{BaseTx: base} },
		func(base BaseTx) flatTx { return &VaultCreate{BaseTx: base} },
		func(base BaseTx) flatTx { return &VaultDelete{BaseTx: base} },
		func(base BaseTx) flatTx { return &VaultDeposit{BaseTx: base} },
		func(base BaseTx) flatTx { return &VaultSet{BaseTx: base} },
		func(base BaseTx) flatTx { return &VaultWithdraw{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainAccountCreateCommit{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainAddAccountCreateAttestation{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainAddClaimAttestation{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainClaim{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainCommit{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainCreateBridge{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainCreateClaimID{BaseTx: base} },
		func(base BaseTx) flatTx { return &XChainModifyBridge{BaseTx: base} },
	}
	for _, newTx := range constructors {
		want := newTx(BaseTx{}).TxType().String()
		t.Run(want, func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				stored TxType
			}{
				{name: "empty"},
				{name: "matching", stored: TxType(want)},
				{name: "conflicting", stored: TxType("NotTheConcreteType")},
			} {
				t.Run(tc.name, func(t *testing.T) {
					tx := newTx(BaseTx{TransactionType: tc.stored})
					require.Equal(t, want, tx.Flatten()["TransactionType"])
				})
			}
		})
	}
}

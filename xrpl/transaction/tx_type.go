package transaction

// TxType represents the type of an XRPL transaction.
type TxType string

// nolint // otherwise issues with the Credential transaction types
const (
	AccountSetTx       TxType = "AccountSet"
	AccountDeleteTx    TxType = "AccountDelete"
	AMMBidTx           TxType = "AMMBid"
	AMMClawbackTx      TxType = "AMMClawback"
	AMMCreateTx        TxType = "AMMCreate"
	AMMDeleteTx        TxType = "AMMDelete"
	AMMDepositTx       TxType = "AMMDeposit"
	AMMVoteTx          TxType = "AMMVote"
	AMMWithdrawTx      TxType = "AMMWithdraw"
	BatchTx            TxType = "Batch"
	CheckCancelTx      TxType = "CheckCancel"
	CheckCashTx        TxType = "CheckCash"
	CheckCreateTx      TxType = "CheckCreate"
	ClawbackTx         TxType = "Clawback"
	CredentialAcceptTx TxType = "CredentialAccept"
	CredentialCreateTx TxType = "CredentialCreate"
	CredentialDeleteTx TxType = "CredentialDelete"
	DelegateSetTx      TxType = "DelegateSet"
	DepositPreauthTx   TxType = "DepositPreauth"
	DIDDeleteTx        TxType = "DIDDelete"
	DIDSetTx           TxType = "DIDSet"
	EscrowCancelTx     TxType = "EscrowCancel"
	EscrowCreateTx     TxType = "EscrowCreate"
	EscrowFinishTx     TxType = "EscrowFinish"
	MPTokenAuthorizeTx TxType = "MPTokenAuthorize"
	//nolint:gosec
	MPTokenIssuanceCreateTx TxType = "MPTokenIssuanceCreate"
	//nolint:gosec
	MPTokenIssuanceDestroyTx TxType = "MPTokenIssuanceDestroy"
	//nolint:gosec
	MPTokenIssuanceSetTx                TxType = "MPTokenIssuanceSet"
	NFTokenAcceptOfferTx                TxType = "NFTokenAcceptOffer"
	NFTokenBurnTx                       TxType = "NFTokenBurn"
	NFTokenCancelOfferTx                TxType = "NFTokenCancelOffer"
	NFTokenCreateOfferTx                TxType = "NFTokenCreateOffer"
	NFTokenMintTx                       TxType = "NFTokenMint"
	NFTokenModifyTx                     TxType = "NFTokenModify"
	OfferCreateTx                       TxType = "OfferCreate"
	OfferCancelTx                       TxType = "OfferCancel"
	OracleDeleteTx                      TxType = "OracleDelete"
	OracleSetTx                         TxType = "OracleSet"
	PaymentTx                           TxType = "Payment"
	PaymentChannelClaimTx               TxType = "PaymentChannelClaim"
	PaymentChannelCreateTx              TxType = "PaymentChannelCreate"
	PaymentChannelFundTx                TxType = "PaymentChannelFund"
	PermissionedDomainDeleteTx          TxType = "PermissionedDomainDelete"
	PermissionedDomainSetTx             TxType = "PermissionedDomainSet"
	SetRegularKeyTx                     TxType = "SetRegularKey"
	SignerListSetTx                     TxType = "SignerListSet"
	TrustSetTx                          TxType = "TrustSet"
	TicketCreateTx                      TxType = "TicketCreate"
	HashedTx                            TxType = "HASH"   // TX stored as a string, rather than complete tx obj
	BinaryTx                            TxType = "BINARY" // TX stored as a string, json tagged as 'tx_blob'
	XChainAccountCreateCommitTx         TxType = "XChainAccountCreateCommit"
	XChainAddAccountCreateAttestationTx TxType = "XChainAddAccountCreateAttestation"
	XChainAddClaimAttestationTx         TxType = "XChainAddClaimAttestation"
	XChainCreateBridgeTx                TxType = "XChainCreateBridge"
	XChainCreateClaimIDTx               TxType = "XChainCreateClaimID"
	XChainClaimTx                       TxType = "XChainClaim"
	XChainCommitTx                      TxType = "XChainCommit"
	XChainModifyBridgeTx                TxType = "XChainModifyBridge"
	LoanSetTx                           TxType = "LoanSet"
	LoanDeleteTx                        TxType = "LoanDelete"
	LoanManageTx                        TxType = "LoanManage"
	LoanPayTx                           TxType = "LoanPay"
	LoanBrokerSetTx                     TxType = "LoanBrokerSet"
	LoanBrokerDeleteTx                  TxType = "LoanBrokerDelete"
	LoanBrokerCoverDepositTx            TxType = "LoanBrokerCoverDeposit"
	LoanBrokerCoverWithdrawTx           TxType = "LoanBrokerCoverWithdraw"
	LoanBrokerCoverClawbackTx           TxType = "LoanBrokerCoverClawback"
)

func (t TxType) String() string {
	return string(t)
}

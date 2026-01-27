package transaction

import (
	"errors"
	"fmt"
)

var (
	// ErrDestinationAccountConflict is returned when the Destination matches the Account.
	ErrDestinationAccountConflict = errors.New("destination cannot be the same as the Account")
	// ErrInvalidAccount is returned when the Account field does not meet XRPL address standards.
	ErrInvalidAccount = errors.New("invalid xrpl address for Account")
	// ErrInvalidDelegate is returned when the Delegate field does not meet XRPL address standards.
	ErrInvalidDelegate = errors.New("invalid xrpl address for Delegate")
	// ErrDelegateAccountConflict is returned when the Delegate matches the Account.
	ErrDelegateAccountConflict = errors.New("addresses for Account and Delegate cannot be the same")
	// ErrInvalidCheckID is returned when the CheckID is not a valid 64-character hexadecimal string.
	ErrInvalidCheckID = errors.New("invalid CheckID, must be a valid 64-character hexadecimal string")
	// ErrInvalidCredentialIDs is returned when the CredentialIDs field is empty or not a valid hexadecimal string array.
	ErrInvalidCredentialIDs = errors.New("invalid credential IDs, must be a valid hexadecimal string array")
	// ErrInvalidDestination is returned when the Destination field does not meet XRPL address standards.
	ErrInvalidDestination = errors.New("invalid xrpl address for Destination")
	// ErrInvalidIssuer is returned when the issuer address is an invalid xrpl address.
	ErrInvalidIssuer = errors.New("invalid xrpl address for Issuer")
	// ErrInvalidOwner is returned when the Owner field does not meet XRPL address standards.
	ErrInvalidOwner = errors.New("invalid xrpl address for Owner")
	// ErrInvalidHexPublicKey is returned when the PublicKey is not a valid hexadecimal string.
	ErrInvalidHexPublicKey = errors.New("invalid PublicKey, must be a valid hexadecimal string")
	// ErrInvalidTransactionType is returned when the TransactionType field is invalid or missing.
	ErrInvalidTransactionType = errors.New("invalid or missing TransactionType")
	// ErrInvalidSubject is returned when the Subject field is an invalid xrpl address.
	ErrInvalidSubject = errors.New("invalid xrpl address for Subject")
	// ErrInvalidURI is returned when the URI is not a valid hexadecimal string.
	ErrInvalidURI = errors.New("invalid URI, must be a valid hexadecimal string")
	// ErrOwnerAccountConflict is returned when the owner is the same as the account.
	ErrOwnerAccountConflict = errors.New("owner must be different from the account")

	// ErrInvalidFlags is returned when provided flags for XChainModifyBridge are invalid.
	ErrInvalidFlags = errors.New("invalid flags")

	// xchain

	// ErrInvalidDestinationAddress is returned when the destination address is invalid.
	ErrInvalidDestinationAddress = errors.New("xchain claim: invalid destination address")
	// ErrMissingXChainClaimID is returned when the XChainClaimID is missing.
	ErrMissingXChainClaimID = errors.New("xchain claim: missing XChainClaimID")

	// ErrInvalidXChainClaimID is returned when the XChainClaimID is invalid or missing.
	ErrInvalidXChainClaimID = errors.New("invalid XChainClaimID")

	// ErrInvalidAttestationRewardAccount is returned when the AttestationRewardAccount is not a valid address.
	ErrInvalidAttestationRewardAccount = errors.New("invalid attestation reward account")
	// ErrInvalidAttestationSignerAccount is returned when the AttestationSignerAccount is not a valid address.
	ErrInvalidAttestationSignerAccount = errors.New("invalid attestation signer account")
	// ErrInvalidOtherChainSource is returned when OtherChainSource is not a valid address.
	ErrInvalidOtherChainSource = errors.New("invalid other chain source")
	// ErrInvalidPublicKey is returned when the PublicKey field is empty or invalid.
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrInvalidWasLockingChainSend is returned when WasLockingChainSend is not 0 or 1.
	ErrInvalidWasLockingChainSend = errors.New("invalid was locking chain send")
	// ErrInvalidXChainAccountCreateCount is returned when XChainAccountCreateCount is not a valid unsigned integer.
	ErrInvalidXChainAccountCreateCount = errors.New("invalid x chain account create count")

	// validations

	// ErrEmptyPath is returned when the path is empty.
	ErrEmptyPath = errors.New("path(s) should have at least one path")
	// ErrInvalidTokenCurrency is returned when the token currency is XRP.
	ErrInvalidTokenCurrency = errors.New("invalid or missing token currency, it also cannot have a similar standard code as XRP")
	// ErrInvalidTokenFields is returned when the issued currency object does not have the required fields (currency, issuer and value).
	ErrInvalidTokenFields = errors.New("issued currency object should have 3 fields: currency, issuer, value")
	// ErrInvalidPathStepCombination is returned when the path step is invalid. The fields combination is invalid.
	ErrInvalidPathStepCombination = errors.New("invalid path step, check the valid fields combination at https://xrpl.org/docs/concepts/tokens/fungible-tokens/paths#path-specifications")
	// ErrInvalidTokenValue is returned when the value field is not a valid positive number.
	ErrInvalidTokenValue = errors.New("value field should be a valid positive number")
	// ErrInvalidTokenType is returned when an issued currency is of type XRP.
	ErrInvalidTokenType = errors.New("an issued currency cannot be of type XRP")
	// ErrMissingTokenCurrency is returned when the currency field is missing for an issued currency.
	ErrMissingTokenCurrency = errors.New("currency field is missing for the issued currency")
	// ErrInvalidAssetFields is returned when the asset object does not have the required fields (currency, or currency and issuer).
	ErrInvalidAssetFields = errors.New("asset object should have at least one field 'currency', or two fields 'currency' and 'issuer'")
	// ErrMissingAssetCurrency is returned when the currency field is missing for an asset.
	ErrMissingAssetCurrency = errors.New("currency field is required for an asset")
	// ErrInvalidAssetIssuer is returned when the issuer field is invalid for an asset.
	ErrInvalidAssetIssuer = errors.New("issuer field must be a valid XRPL classic address")

	// validations_xrpl_objects

	// ErrMemoShouldHaveAtLeastOneField is returned when a memo object is empty.
	ErrMemoShouldHaveAtLeastOneField = errors.New("memo object should have at least one field, MemoData, MemoFormat or MemoType")
	// ErrMemoDataShouldBeHex is returned when MemoData is not a hexadecimal string.
	ErrMemoDataShouldBeHex = errors.New("memoData should be a hexadecimal string")
	// ErrMemoFormatShouldBeHex is returned when MemoFormat is not a hexadecimal string.
	ErrMemoFormatShouldBeHex = errors.New("memoFormat should be a hexadecimal string")
	// ErrMemoTypeShouldBeHex is returned when MemoType is not a hexadecimal string.
	ErrMemoTypeShouldBeHex = errors.New("memoType should be a hexadecimal string")
	// ErrSignerShouldHaveThreeFields is returned when a Signer object doesn't have exactly 3 fields.
	ErrSignerShouldHaveThreeFields = errors.New("signers: Signer should have 3 fields: Account, TxnSignature, SigningPubKey")
	// ErrSignerAccountShouldBeString is returned when the Account field in a Signer is not a valid string.
	ErrSignerAccountShouldBeString = errors.New("signers: Account should be a string")
	// ErrSignerTxnSignatureShouldBeNonEmpty is returned when TxnSignature in a Signer is empty.
	ErrSignerTxnSignatureShouldBeNonEmpty = errors.New("signers: TxnSignature should be a non-empty string")
	// ErrSignerSigningPubKeyShouldBeNonEmpty is returned when SigningPubKey in a Signer is empty.
	ErrSignerSigningPubKeyShouldBeNonEmpty = errors.New("signers: SigningPubKey should be a non-empty string")
	// ErrInvalidDomainID is returned when the provided DomainID is invalid.
	ErrInvalidDomainID = errors.New("invalid DomainID value")

	// trust set

	// ErrTrustSetMissingLimitAmount is returned when the LimitAmount field is not set on a TrustSet transaction.
	ErrTrustSetMissingLimitAmount = errors.New("missing field LimitAmount")

	// signer list set

	// ErrInvalidSignerEntries is returned when the number of signer entries is outside the allowed range.
	ErrInvalidSignerEntries = errors.New("invalid number of signer entries")
	// ErrInvalidWalletLocator is returned when a SignerEntry's WalletLocator is not a valid hexadecimal string.
	ErrInvalidWalletLocator = errors.New("invalid WalletLocator in SignerEntry, must be a hexadecimal string")
	// ErrSignerQuorumGreaterThanSumOfSignerWeights is returned when SignerQuorum exceeds sum of all SignerWeights.
	ErrSignerQuorumGreaterThanSumOfSignerWeights = errors.New("signerQuorum must be less than or equal to the sum of all SignerWeights")
	// ErrInvalidQuorumAndEntries is returned when SignerEntries is non-empty while SignerQuorum is zero.
	ErrInvalidQuorumAndEntries = errors.New("signerEntries must be empty when the SignerQuorum is set to 0 to delete a signer list")

	// ErrInvalidRegularKey is returned when the RegularKey field contains an invalid XRPL address.
	ErrInvalidRegularKey = errors.New("invalid xrpl address for the RegularKey field")
	// ErrRegularKeyMatchesAccount is returned when the regular key address matches the account address.
	ErrRegularKeyMatchesAccount = errors.New("regular key must not match the account address")

	// permissioned domain

	// ErrMissingDomainID is returned when the required DomainID field is missing.
	ErrMissingDomainID = errors.New("missing required field: DomainID")

	// payment

	// ErrPartialPaymentFlagRequired is returned when the tfPartialPayment flag is required but not set.
	ErrPartialPaymentFlagRequired = errors.New("tfPartialPayment flag required with DeliverMin")

	// ErrInvalidExpiration indicates the expiration time must be either later than the current time plus the SettleDelay of the channel, or the existing Expiration of the channel.
	ErrInvalidExpiration = errors.New("expiration time must be either later than the current time plus the SettleDelay of the channel, or the existing Expiration of the channel")

	// ErrInvalidChannel is returned when the Channel is not a valid 64-character hexadecimal string.
	ErrInvalidChannel = errors.New("invalid Channel, must be a valid 64-character hexadecimal string")
	// ErrInvalidSignature is returned when the Signature is not a valid hexadecimal string.
	ErrInvalidSignature = errors.New("invalid Signature, must be a valid hexadecimal string")

	// offer

	// ErrTfHybridCannotBeSetWithoutDomainID is returned if a OfferCreate has tfHybrid enabled and no DomainID set.
	ErrTfHybridCannotBeSetWithoutDomainID = errors.New("tfHybrid must have a valid DomainID")

	// nft

	// ErrInvalidTransferFee is returned when the transferFee is not between 0 and 50000 inclusive.
	ErrInvalidTransferFee = errors.New("transferFee must be between 0 and 50000 inclusive")
	// ErrIssuerAccountConflict is returned when the issuer is the same as the account.
	ErrIssuerAccountConflict = errors.New("issuer cannot be the same as the account")
	// ErrTransferFeeRequiresTransferableFlag is returned when the transferFee is set without the tfTransferable flag.
	ErrTransferFeeRequiresTransferableFlag = errors.New("transferFee can only be set if the tfTransferable flag is enabled")
	// ErrAmountRequiredWithExpirationOrDestination is returned when Expiration or Destination is set without Amount.
	ErrAmountRequiredWithExpirationOrDestination = errors.New("amount is required when Expiration or Destination is present")

	// ErrOwnerPresentForSellOffer is returned when the owner is present for a sell offer.
	ErrOwnerPresentForSellOffer = errors.New("owner must not be present for a sell offer")
	// ErrOwnerNotPresentForBuyOffer is returned when the owner is not present for a buy offer.
	ErrOwnerNotPresentForBuyOffer = errors.New("owner must be present for a buy offer")

	// ErrEmptyNFTokenOffers is returned when the NFTokenOffers array contains no entries.
	ErrEmptyNFTokenOffers = errors.New("the NFTokenOffers array must have at least one entry")

	// ErrInvalidNFTokenID is returned when the NFTokenID is not a hexadecimal.
	ErrInvalidNFTokenID = errors.New("invalid NFTokenID, must be a hexadecimal string")

	// ErrNFTokenBrokerFeeZero is returned when NFTokenBrokerFee is zero.
	ErrNFTokenBrokerFeeZero = errors.New("nftoken accept offer: NFTokenBrokerFee cannot be zero")
	// ErrMissingOffer is returned when at least one of NFTokenSellOffer or NFTokenBuyOffer is not set.
	ErrMissingOffer = errors.New("at least one of NFTokenSellOffer or NFTokenBuyOffer must be set")
	// ErrMissingBothOffers is returned when NFTokenBrokerFee is set but neither NFTokenSellOffer nor NFTokenBuyOffer are set (brokered mode).
	ErrMissingBothOffers = errors.New("when NFTokenBrokerFee is set (brokered mode), both NFTokenSellOffer and NFTokenBuyOffer must be set")

	// mpt

	// ErrMPTokenIssuanceSetFlags is returned when both tfMPTLock and tfMPTUnlock flags are enabled simultaneously.
	ErrMPTokenIssuanceSetFlags = errors.New("mptoken issuance set: tfMPTLock and tfMPTUnlock flags cannot both be enabled")

	// ErrInvalidMPTokenIssuanceID is returned when the MPTokenIssuanceID is empty or invalid.
	ErrInvalidMPTokenIssuanceID = errors.New("mptoken issuance destroy: invalid MPTokenIssuanceID")

	// ErrTransferFeeRequiresCanTransfer is returned when TransferFee is set without enabling tfMPTCanTransfer flag.
	ErrTransferFeeRequiresCanTransfer = errors.New("mptoken issuance create: TransferFee cannot be provided without enabling tfMPTCanTransfer flag")
	// ErrInvalidMPTokenMetadata is returned when MPTokenMetadata is not a valid hex string or exceeds size limit.
	ErrInvalidMPTokenMetadata = errors.New("mptoken issuance create: MPTokenMetadata must be a valid hex string and at most 1024 bytes")

	// ErrHolderAccountConflict is returned when the holder account is the same as the issuing account.
	ErrHolderAccountConflict = errors.New("holder must be different from the account")

	// escrow

	// ErrEscrowFinishMissingOwner is returned when the Owner field is missing in an EscrowFinish transaction.
	ErrEscrowFinishMissingOwner = errors.New("escrow finish: missing owner")
	// ErrEscrowFinishMissingOfferSequence is returned when the OfferSequence is zero in an EscrowFinish transaction.
	ErrEscrowFinishMissingOfferSequence = errors.New("escrow finish: missing offer sequence")

	// ErrEscrowCreateInvalidDestinationAddress is returned when the destination address for EscrowCreate is invalid.
	ErrEscrowCreateInvalidDestinationAddress = errors.New("escrow create: invalid destination address")
	// ErrEscrowCreateNoConditionOrFinishAfterSet is returned when both Condition and FinishAfter are unset.
	ErrEscrowCreateNoConditionOrFinishAfterSet = errors.New("escrow create: either Condition or FinishAfter must be specified")

	// ErrEscrowCancelMissingOwner indicates the Owner field is missing when canceling an escrow.
	ErrEscrowCancelMissingOwner = errors.New("escrow cancel: missing owner")
	// ErrEscrowCancelMissingOfferSequence indicates the OfferSequence field is missing when canceling an escrow.
	ErrEscrowCancelMissingOfferSequence = errors.New("escrow cancel: missing offer sequence")

	// did

	// ErrDIDSetMustSetEitherDataOrDIDDocumentOrURI is returned when Data, DIDDocument, and URI are all unset in a DIDSet transaction.
	ErrDIDSetMustSetEitherDataOrDIDDocumentOrURI = errors.New("did set: must set either Data, DIDDocument, or URI")

	// deposit preauth

	// ErrDepositPreauthInvalidAuthorize is returned when the Authorize address is invalid.
	ErrDepositPreauthInvalidAuthorize = errors.New("deposit preauth: invalid Authorize")
	// ErrDepositPreauthInvalidUnauthorize is returned when the Unauthorize address is invalid.
	ErrDepositPreauthInvalidUnauthorize = errors.New("deposit preauth: invalid Unauthorize")
	// ErrDepositPreauthInvalidAuthorizeCredentials is returned when an AuthorizeCredentials entry is invalid.
	ErrDepositPreauthInvalidAuthorizeCredentials = errors.New("deposit preauth: invalid AuthorizeCredentials")
	// ErrDepositPreauthInvalidUnauthorizeCredentials is returned when an UnauthorizeCredentials entry is invalid.
	ErrDepositPreauthInvalidUnauthorizeCredentials = errors.New("deposit preauth: invalid UnauthorizeCredentials")
	// ErrDepositPreauthMustSetOnlyOneField is returned when more than one preauth field is set.
	ErrDepositPreauthMustSetOnlyOneField = errors.New("deposit preauth: must set only one field (Authorize or AuthorizeCredentials or Unauthorize or UnauthorizeCredentials)")
	// ErrDepositPreauthAuthorizeCannotBeSender is returned when Authorize equals the sender's account.
	ErrDepositPreauthAuthorizeCannotBeSender = errors.New("deposit preauth: Authorize cannot be the same as the sender's account")
	// ErrDepositPreauthUnauthorizeCannotBeSender is returned when Unauthorize equals the sender's account.
	ErrDepositPreauthUnauthorizeCannotBeSender = errors.New("deposit preauth: Unauthorize cannot be the same as the sender's account")

	// delegate set

	// ErrDelegateSetAuthorizeAccountConflict is returned when the Authorize account matches the Account.
	ErrDelegateSetAuthorizeAccountConflict = errors.New("authorize account cannot be the same as the Account")
	// ErrDelegateSetPermissionMalformed is returned when the Permissions array is empty or malformed.
	ErrDelegateSetPermissionMalformed = errors.New("permissions array is required and cannot be empty")
	// ErrDelegateSetPermissionsMaxLength is returned when the Permissions array exceeds the maximum length.
	ErrDelegateSetPermissionsMaxLength = errors.New("permissions array cannot exceed maximum length")
	// ErrDelegateSetEmptyPermissionValue is returned when a permission value is empty or undefined.
	ErrDelegateSetEmptyPermissionValue = errors.New("permission value cannot be empty")
	// ErrDelegateSetNonDelegatableTransaction is returned when trying to delegate a non-delegatable transaction type.
	ErrDelegateSetNonDelegatableTransaction = errors.New("cannot delegate non-delegatable transaction types")
	// ErrDelegateSetDuplicatePermissions is returned when the same permission is specified multiple times.
	ErrDelegateSetDuplicatePermissions = errors.New("duplicate permissions are not allowed")

	// credential

	// ErrInvalidCredentialURI is returned when the URI field does not meet the maximum allowed hex-encoded length of 512 characters (256 bytes).
	ErrInvalidCredentialURI = errors.New("credential create: invalid URI, must have a maximum hex string length of 512 characters (256 bytes)")

	// clawback

	// ErrClawbackMissingAmount is returned when the Amount field is not set.
	ErrClawbackMissingAmount = errors.New("clawback: missing field Amount")
	// ErrClawbackInvalidAmount is returned when the Amount is not a valid issued currency.
	ErrClawbackInvalidAmount = errors.New("clawback: invalid Amount")
	// ErrClawbackSameAccount is returned when the clawback account and the token issuer are the same.
	ErrClawbackSameAccount = errors.New("clawback: Account and Amount.issuer cannot be the same")

	// check

	// ErrAmountOrDeliverMinNotProvided is returned when neither Amount nor DeliverMin is provided.
	ErrAmountOrDeliverMinNotProvided = errors.New("check cash: either Amount or DeliverMin must be provided")
	// ErrMutuallyExclusiveAmountDeliverMin is returned when both Amount and DeliverMin are provided.
	ErrMutuallyExclusiveAmountDeliverMin = errors.New("check cash: both Amount and DeliverMin cannot be provided")

	// batch

	// ErrBatchRawTransactionsEmpty is returned when the RawTransactions array is empty or nil.
	// This validates that a batch transaction contains at least one inner transaction to execute.
	ErrBatchRawTransactionsEmpty = errors.New("rawTransactions must be a non-empty array")

	// balance

	errLowLimitIssuerNotFound        = errors.New("low limit issuer not found")
	errHighLimitIssuerNotFound       = errors.New("high limit issuer not found")
	errBalanceCurrencyNotFound       = errors.New("balance currency not found")
	errInvalidBalanceValue           = errors.New("invalid balance value")
	errBalanceNotFound               = errors.New("balance not found")
	errAccountNotFoundForXRPQuantity = errors.New("account not found for XRP quantity")

	// amm

	// ErrAMMAtLeastOneAssetMustBeSet is returned when no deposit asset is specified in the AMM deposit.
	ErrAMMAtLeastOneAssetMustBeSet = errors.New("at least one of the assets must be set")

	// ErrAMMMustSetAmountWithAmount2 is returned when Amount2 is set without Amount.
	ErrAMMMustSetAmountWithAmount2 = errors.New("must set Amount with Amount2")
	// ErrAMMMustSetAmountWithEPrice is returned when EPrice is set without Amount.
	ErrAMMMustSetAmountWithEPrice = errors.New("must set Amount with EPrice")

	// ErrInvalidHolder is returned when the holder is invalid.
	ErrInvalidHolder = errors.New("invalid holder")
	// ErrInvalidAmountIssuer is returned when the amount issuer is invalid.
	ErrInvalidAmountIssuer = errors.New("invalid amount issuer")

	// ErrAMMAtLeastOneAssetMustBeNonXRP is returned when both assets are XRP; at least one asset must be non-XRP.
	ErrAMMAtLeastOneAssetMustBeNonXRP = errors.New("at least one of the assets must be non-XRP")
	// ErrAMMAuthAccountsTooMany is returned when more than four AuthAccount objects are provided.
	ErrAMMAuthAccountsTooMany = errors.New("authAccounts should have at most 4 AuthAccount objects")

	// account

	// ErrAccountSetInvalidSetFlag is returned when SetFlag is outside the valid range (1 to 16).
	ErrAccountSetInvalidSetFlag = errors.New("account set: SetFlag must be an integer between asfRequireDest (1) and asfAllowTrustLineClawback (16)")
	// ErrAccountSetInvalidTickSize is returned when TickSize is outside the valid range (0 to 15 inclusive).
	ErrAccountSetInvalidTickSize = errors.New("account set: TickSize must be an integer between 0 and 15 inclusive")

	// loan

	// ErrLoanSetLoanBrokerIDRequired is returned when LoanBrokerID is not set on a LoanSet transaction.
	ErrLoanSetLoanBrokerIDRequired = errors.New("loanSet: LoanBrokerID is required")
	// ErrLoanSetLoanBrokerIDInvalid is returned when LoanBrokerID is not a valid 64-character hexadecimal string.
	ErrLoanSetLoanBrokerIDInvalid = errors.New("loanSet: LoanBrokerID must be 64 characters hexadecimal string")
	// ErrLoanSetPrincipalRequestedRequired is returned when PrincipalRequested is not set on a LoanSet transaction.
	ErrLoanSetPrincipalRequestedRequired = errors.New("loanSet: PrincipalRequested is required")
	// ErrLoanSetPrincipalRequestedInvalid is returned when PrincipalRequested is not a valid XRPL number.
	ErrLoanSetPrincipalRequestedInvalid = errors.New("loanSet: PrincipalRequested must be a valid XRPL number")
	// ErrLoanSetDataInvalid is returned when Data is not a valid non-empty hex string up to 512 characters.
	ErrLoanSetDataInvalid = errors.New("loanSet: Data must be a valid non-empty hex string up to 512 characters")
	// ErrLoanSetOverpaymentFeeInvalid is returned when OverpaymentFee is outside the valid range (0 to 100000 inclusive).
	ErrLoanSetOverpaymentFeeInvalid = errors.New("loanSet: OverpaymentFee must be between 0 and 100000 inclusive")
	// ErrLoanSetInterestRateInvalid is returned when InterestRate is outside the valid range (0 to 100000 inclusive).
	ErrLoanSetInterestRateInvalid = errors.New("loanSet: InterestRate must be between 0 and 100000 inclusive")
	// ErrLoanSetLateInterestRateInvalid is returned when LateInterestRate is outside the valid range (0 to 100000 inclusive).
	ErrLoanSetLateInterestRateInvalid = errors.New("loanSet: LateInterestRate must be between 0 and 100000 inclusive")
	// ErrLoanSetCloseInterestRateInvalid is returned when CloseInterestRate is outside the valid range (0 to 100000 inclusive).
	ErrLoanSetCloseInterestRateInvalid = errors.New("loanSet: CloseInterestRate must be between 0 and 100000 inclusive")
	// ErrLoanSetOverpaymentInterestRateInvalid is returned when OverpaymentInterestRate is outside the valid range (0 to 100000 inclusive).
	ErrLoanSetOverpaymentInterestRateInvalid = errors.New("loanSet: OverpaymentInterestRate must be between 0 and 100000 inclusive")
	// ErrLoanSetPaymentIntervalInvalid is returned when PaymentInterval is less than 60.
	ErrLoanSetPaymentIntervalInvalid = errors.New("loanSet: PaymentInterval must be greater than or equal to 60")
	// ErrLoanSetGracePeriodInvalid is returned when GracePeriod is greater than PaymentInterval.
	ErrLoanSetGracePeriodInvalid = errors.New("loanSet: GracePeriod must not be greater than PaymentInterval")
	// ErrLoanSetLoanOriginationFeeInvalid is returned when LoanOriginationFee is not a valid XRPL number.
	ErrLoanSetLoanOriginationFeeInvalid = errors.New("loanSet: LoanOriginationFee must be a valid XRPL number")
	// ErrLoanSetLoanServiceFeeInvalid is returned when LoanServiceFee is not a valid XRPL number.
	ErrLoanSetLoanServiceFeeInvalid = errors.New("loanSet: LoanServiceFee must be a valid XRPL number")
	// ErrLoanSetLatePaymentFeeInvalid is returned when LatePaymentFee is not a valid XRPL number.
	ErrLoanSetLatePaymentFeeInvalid = errors.New("loanSet: LatePaymentFee must be a valid XRPL number")
	// ErrLoanSetClosePaymentFeeInvalid is returned when ClosePaymentFee is not a valid XRPL number.
	ErrLoanSetClosePaymentFeeInvalid = errors.New("loanSet: ClosePaymentFee must be a valid XRPL number")

	// ErrLoanDeleteLoanIDRequired is returned when LoanID is not set on a LoanDelete transaction.
	ErrLoanDeleteLoanIDRequired = errors.New("loanDelete: LoanID is required")
	// ErrLoanDeleteLoanIDInvalid is returned when LoanID is not a valid 64-character hexadecimal string.
	ErrLoanDeleteLoanIDInvalid = errors.New("loanDelete: LoanID must be 64 characters hexadecimal string")

	// ErrLoanManageLoanIDRequired is returned when LoanID is not set on a LoanManage transaction.
	ErrLoanManageLoanIDRequired = errors.New("loanManage: LoanID is required")
	// ErrLoanManageLoanIDInvalid is returned when LoanID is not a valid 64-character hexadecimal string.
	ErrLoanManageLoanIDInvalid = errors.New("loanManage: LoanID must be 64 characters hexadecimal string")
	// ErrLoanManageFlagsConflict is returned when tfLoanImpair and tfLoanUnimpair flags are both set.
	ErrLoanManageFlagsConflict = errors.New("loanManage: tfLoanImpair and tfLoanUnimpair cannot both be present")

	// ErrLoanPayLoanIDRequired is returned when LoanID is not set on a LoanPay transaction.
	ErrLoanPayLoanIDRequired = errors.New("loanPay: LoanID is required")
	// ErrLoanPayLoanIDInvalid is returned when LoanID is not a valid 64-character hexadecimal string.
	ErrLoanPayLoanIDInvalid = errors.New("loanPay: LoanID must be 64 characters hexadecimal string")
	// ErrLoanPayAmountRequired is returned when Amount is not set on a LoanPay transaction.
	ErrLoanPayAmountRequired = errors.New("loanPay: Amount is required")

	// ErrLoanBrokerSetVaultIDRequired is returned when VaultID is not set on a LoanBrokerSet transaction.
	ErrLoanBrokerSetVaultIDRequired = errors.New("loanBrokerSet: VaultID is required")
	// ErrLoanBrokerSetVaultIDInvalid is returned when VaultID is not a valid 64-character hexadecimal string.
	ErrLoanBrokerSetVaultIDInvalid = errors.New("loanBrokerSet: VaultID must be 64 characters hexadecimal string")
	// ErrLoanBrokerSetLoanBrokerIDInvalid is returned when LoanBrokerID is not a valid 64-character hexadecimal string.
	ErrLoanBrokerSetLoanBrokerIDInvalid = errors.New("loanBrokerSet: LoanBrokerID must be 64 characters hexadecimal string")
	// ErrLoanBrokerSetDataInvalid is returned when Data is not a valid non-empty hex string up to 512 characters.
	ErrLoanBrokerSetDataInvalid = errors.New("loanBrokerSet: Data must be a valid non-empty hex string up to 512 characters")
	// ErrLoanBrokerSetManagementFeeRateInvalid is returned when ManagementFeeRate is outside the valid range (0 to 10000 inclusive).
	ErrLoanBrokerSetManagementFeeRateInvalid = errors.New("loanBrokerSet: ManagementFeeRate must be between 0 and 10000 inclusive")
	// ErrLoanBrokerSetDebtMaximumInvalid is returned when DebtMaximum is not a valid XRPL number.
	ErrLoanBrokerSetDebtMaximumInvalid = errors.New("loanBrokerSet: DebtMaximum must be a valid XRPL number")
	// ErrLoanBrokerSetDebtMaximumNegative is returned when DebtMaximum is negative.
	ErrLoanBrokerSetDebtMaximumNegative = errors.New("loanBrokerSet: DebtMaximum must be a non-negative value")
	// ErrLoanBrokerSetCoverRateMinimumInvalid is returned when CoverRateMinimum is outside the valid range (0 to 100000 inclusive).
	ErrLoanBrokerSetCoverRateMinimumInvalid = errors.New("loanBrokerSet: CoverRateMinimum must be between 0 and 100000 inclusive")
	// ErrLoanBrokerSetCoverRateLiquidationInvalid is returned when CoverRateLiquidation is outside the valid range (0 to 100000 inclusive).
	ErrLoanBrokerSetCoverRateLiquidationInvalid = errors.New("loanBrokerSet: CoverRateLiquidation must be between 0 and 100000 inclusive")
	// ErrLoanBrokerSetCoverRatesMismatch is returned when CoverRateMinimum and CoverRateLiquidation are not both zero or both non-zero.
	ErrLoanBrokerSetCoverRatesMismatch = errors.New("loanBrokerSet: CoverRateMinimum and CoverRateLiquidation must both be zero or both be non-zero")

	// ErrLoanBrokerDeleteLoanBrokerIDRequired is returned when LoanBrokerID is not set on a LoanBrokerDelete transaction.
	ErrLoanBrokerDeleteLoanBrokerIDRequired = errors.New("loanBrokerDelete: LoanBrokerID is required")
	// ErrLoanBrokerDeleteLoanBrokerIDInvalid is returned when LoanBrokerID is not a valid 64-character hexadecimal string.
	ErrLoanBrokerDeleteLoanBrokerIDInvalid = errors.New("loanBrokerDelete: LoanBrokerID must be 64 characters hexadecimal string")

	// ErrLoanBrokerCoverDepositLoanBrokerIDRequired is returned when LoanBrokerID is not set on a LoanBrokerCoverDeposit transaction.
	ErrLoanBrokerCoverDepositLoanBrokerIDRequired = errors.New("loanBrokerCoverDeposit: LoanBrokerID is required")
	// ErrLoanBrokerCoverDepositLoanBrokerIDInvalid is returned when LoanBrokerID is not a valid 64-character hexadecimal string.
	ErrLoanBrokerCoverDepositLoanBrokerIDInvalid = errors.New("loanBrokerCoverDeposit: LoanBrokerID must be 64 characters hexadecimal string")
	// ErrLoanBrokerCoverDepositAmountRequired is returned when Amount is not set on a LoanBrokerCoverDeposit transaction.
	ErrLoanBrokerCoverDepositAmountRequired = errors.New("loanBrokerCoverDeposit: Amount is required")

	// ErrLoanBrokerCoverWithdrawLoanBrokerIDRequired is returned when LoanBrokerID is not set on a LoanBrokerCoverWithdraw transaction.
	ErrLoanBrokerCoverWithdrawLoanBrokerIDRequired = errors.New("loanBrokerCoverWithdraw: LoanBrokerID is required")
	// ErrLoanBrokerCoverWithdrawLoanBrokerIDInvalid is returned when LoanBrokerID is not a valid 64-character hexadecimal string.
	ErrLoanBrokerCoverWithdrawLoanBrokerIDInvalid = errors.New("loanBrokerCoverWithdraw: LoanBrokerID must be 64 characters hexadecimal string")
	// ErrLoanBrokerCoverWithdrawAmountRequired is returned when Amount is not set on a LoanBrokerCoverWithdraw transaction.
	ErrLoanBrokerCoverWithdrawAmountRequired = errors.New("loanBrokerCoverWithdraw: Amount is required")

	// ErrLoanBrokerCoverClawbackLoanBrokerIDInvalid is returned when LoanBrokerID is not a valid 64-character hexadecimal string.
	ErrLoanBrokerCoverClawbackLoanBrokerIDInvalid = errors.New("loanBrokerCoverClawback: LoanBrokerID must be 64 characters hexadecimal string")
	// ErrLoanBrokerCoverClawbackAmountInvalidType is returned when Amount is not an IssuedCurrencyAmount or MPTCurrencyAmount.
	ErrLoanBrokerCoverClawbackAmountInvalidType = errors.New("loanBrokerCoverClawback: Amount must be an IssuedCurrencyAmount or MPTCurrencyAmount")
	// ErrLoanBrokerCoverClawbackAmountNegative is returned when Amount is negative.
	ErrLoanBrokerCoverClawbackAmountNegative = errors.New("loanBrokerCoverClawback: Amount must be >= 0")
	// ErrLoanBrokerCoverClawbackLoanBrokerIDOrAmountRequired is returned when neither LoanBrokerID nor Amount is provided.
	ErrLoanBrokerCoverClawbackLoanBrokerIDOrAmountRequired = errors.New("loanBrokerCoverClawback: Either LoanBrokerID or Amount is required")
)

// ErrAMMTradingFeeTooHigh is returned when the AMM trading fee exceeds the maximum allowed.
type ErrAMMTradingFeeTooHigh struct {
	Value uint16
	Limit uint64
}

// Error implements the error interface for ErrAMMTradingFeeToHigh
func (e ErrAMMTradingFeeTooHigh) Error() string {
	return fmt.Sprintf("AMM trading fee exceeds maximum allowed: got %d, must be less or equal than %d", e.Value, e.Limit)
}

// ErrOracleProviderLength is returned when the Provider field exceeds OracleSetProviderMaxLength bytes.
type ErrOracleProviderLength struct {
	Length int
	Limit  int
}

// Error implements the error interface for ErrOracleProviderLength
func (e ErrOracleProviderLength) Error() string {
	return fmt.Sprintf("provider length exceeds maximum: got %d bytes, max %d", e.Length, e.Limit)
}

// ErrOraclePriceDataSeriesItems is returned when the number of PriceDataSeries items exceeds the maximum allowed.
type ErrOraclePriceDataSeriesItems struct {
	Length int
	Limit  int
}

// Error implements the error interface for ErrOraclePriceDataSeriesItems.
func (e ErrOraclePriceDataSeriesItems) Error() string {
	return fmt.Sprintf("oracle price data series items exceed maximum allowed: ot %d, max %d", e.Length, e.Limit)
}

// ErrTicketCreateInvalidTicketCount is returned when the ticket count is outside the valid range.
type ErrTicketCreateInvalidTicketCount struct {
	TicketCount    uint32
	MinTicketCount uint32
	MaxTicketCount uint32
}

// Error implements the error interface for ErrTicketCreateInvalidTicketCount.
func (e ErrTicketCreateInvalidTicketCount) Error() string {
	return fmt.Sprintf("invalid ticket count %d: must be between %d and %d", e.TicketCount, e.MinTicketCount, e.MaxTicketCount)
}

// ErrTransactionInvalidField is returned when a field has an invalid value.
type ErrTransactionInvalidField struct {
	Type  string
	Field string
}

// Error implements the error interface for ErrTransactionInvalidField
func (e ErrTransactionInvalidField) Error() string {
	return fmt.Sprintf("%s invalid field: %s", e.Type, e.Field)
}

// ErrMissingField is returned when a required field is missing.
type ErrMissingField struct {
	Field string
}

// Error implements the error interface for ErrMissingField
func (e ErrMissingField) Error() string {
	return fmt.Sprintf("missing field required: %s", e.Field)
}

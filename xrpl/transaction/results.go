package transaction

// TxResult represents the result code of a transaction
type TxResult string

//revive:disable:var-naming,exported
// #nosec G101

const (
	// tec codes ⬇️ - https://xrpl.org/docs/references/protocol/transactions/transaction-results/tec-codes

	// These codes indicate that the transaction failed, but it was applied to a ledger to apply the transaction cost. They have numerical values in the range 100 to 199. It is recommended to use the text code, not the numeric value.

	// Transactions with tec codes destroy the XRP paid as a transaction cost, and consume a sequence number. For the most part, the transactions take no other action, but there are some exceptions. For example, a transaction that results in tecOVERSIZE still cleans up some unfunded offers. Always look at the transaction metadata to see precisely what a transaction did.
	// ------------------------------------------------------------------------------------------------

	// The transaction failed because the operation is not allowed on Automated Market Maker (AMM) accounts. (Added by the AMM amendment)
	TecAMM_ACCOUNT TxResult = "tecAMM_ACCOUNT"

	// The AMMCreate transaction failed because the sender does not have enough of the specified assets to fund it. (Added by the AMM amendment)
	TecAMM_UNFUNDED TxResult = "tecAMM_UNFUNDED"

	// The AMMDeposit or AMMWithdraw transaction failed because either the AMM or the user does not hold enough of one of the specified assets. (Added by the AMM amendment)
	TecAMM_BALANCE TxResult = "tecAMM_BALANCE"

	// The AMM-related transaction failed because the AMM has no assets in its pool. In this state, you can only delete the AMM or fund it with a new deposit. (Added by the AMM amendment)
	TecAMM_EMPTY TxResult = "tecAMM_EMPTY"

	// The AMM-related transaction failed due to asset availability or pricing constraints. (Added by the AMM amendment)
	TecAMM_FAILED TxResult = "tecAMM_FAILED"

	// The AMM-related transaction failed due to insufficient LP Tokens or problems with rounding. (Added by the AMM amendment)
	TecAMM_INVALID_TOKENS TxResult = "tecAMM_INVALID_TOKENS"

	// The transaction was meant to operate on an AMM with empty asset pools, but the specified AMM currently holds assets. (Added by the AMM amendment)
	TecAMM_NOT_EMPTY TxResult = "tecAMM_NOT_EMPTY"

	// The transaction tried to accept an offer placed by the same account to buy or sell an NFToken. (Added by the NonFungibleTokensV1_1 amendment)
	TecCANT_ACCEPT_OWN_NFTOKEN_OFFER TxResult = "tecCANT_ACCEPT_OWN_NFTOKEN_OFFER"

	// Unspecified failure, with transaction cost destroyed.
	TecCLAIM TxResult = "tecCLAIM"

	// This EscrowCreate or EscrowFinish transaction contained a malformed or mismatched crypto-condition.
	TecCRYPTOCONDITION_ERROR TxResult = "tecCRYPTOCONDITION_ERROR"

	// The transaction tried to add an object to an account's owner directory, but the directory is full.
	TecDIR_FULL TxResult = "tecDIR_FULL"

	// The transaction tried to create an object that already exists.
	TecDUPLICATE TxResult = "tecDUPLICATE"

	// The Payment transaction omitted a destination tag, but the destination account requires one.
	TecDST_TAG_NEEDED TxResult = "tecDST_TAG_NEEDED"

	// The transaction tried to create a DID entry with no contents. (Added by the DID amendment)
	TecEMPTY_DID TxResult = "tecEMPTY_DID"

	// The transaction tried to create an object whose provided Expiration time has already passed.
	TecEXPIRED TxResult = "tecEXPIRED"

	// An unspecified error occurred when processing the transaction.
	TecFAILED_PROCESSING TxResult = "tecFAILED_PROCESSING"

	// The OfferCreate transaction failed because one or both of the assets involved are subject to a global freeze.
	TecFROZEN TxResult = "tecFROZEN"

	// The AccountDelete transaction failed because the account owns objects that cannot be deleted.
	TecHAS_OBLIGATIONS TxResult = "tecHAS_OBLIGATIONS"

	// The transaction failed due to insufficient XRP to create a new trust line.
	TecINSUF_RESERVE_LINE TxResult = "tecINSUF_RESERVE_LINE"

	// The transaction failed due to insufficient XRP to create a new Offer.
	TecINSUF_RESERVE_OFFER TxResult = "tecINSUF_RESERVE_OFFER"

	// The transaction failed because the sending account does not have enough XRP to pay the transaction cost specified.
	TecINSUFF_FEE TxResult = "tecINSUFF_FEE"

	// One of the accounts involved does not hold enough of a necessary asset. (Added by the NonFungibleTokensV1_1 amendment)
	TecINSUFFICIENT_FUNDS TxResult = "tecINSUFFICIENT_FUNDS"

	// The amount specified is not enough to pay all fees involved in the transaction. (Added by the NonFungibleTokensV1_1 amendment)
	TecINSUFFICIENT_PAYMENT TxResult = "tecINSUFFICIENT_PAYMENT"

	// The transaction would increase the reserve requirement higher than the sending account's balance.
	TecINSUFFICIENT_RESERVE TxResult = "tecINSUFFICIENT_RESERVE"

	// Unspecified internal error, with transaction cost applied. Should not normally be returned.
	TecINTERNAL TxResult = "tecINTERNAL"

	// An invariant check failed when executing this transaction. (EnforceInvariants amendment)
	TecINVARIANT_FAILED TxResult = "tecINVARIANT_FAILED"

	// The OfferCreate transaction specified TfFillOrKill and could not be filled.
	TecKILLED TxResult = "tecKILLED"

	// A sequence number field is already at its maximum. (Added by NonFungibleTokensV1_1 amendment)
	TecMAX_SEQUENCE_REACHED TxResult = "tecMAX_SEQUENCE_REACHED"

	// Transaction tried to cause changes requiring the master key.
	TecNEED_MASTER_KEY TxResult = "tecNEED_MASTER_KEY"

	// NFTokenAcceptOffer attempted to match incompatible buy/sell offers.
	TecNFTOKEN_BUY_SELL_MISMATCH TxResult = "tecNFTOKEN_BUY_SELL_MISMATCH"

	// Offer type mismatch in NFTokenAcceptOffer transaction.
	TecNFTOKEN_OFFER_TYPE_MISMATCH TxResult = "tecNFTOKEN_OFFER_TYPE_MISMATCH"

	// The transaction tried to remove the only available method of authorizing transactions. (Prior to rippled 0.30.0, this was called tecMASTER_DISABLED.)
	TecNO_ALTERNATIVE_KEY TxResult = "tecNO_ALTERNATIVE_KEY"

	// The transaction failed because it needs to add a balance on a trust line to an account with the LsfRequireAuth flag enabled, and that trust line has not been authorized.
	TecNO_AUTH TxResult = "tecNO_AUTH"

	// The account on the receiving end of the transaction does not exist.
	TecNO_DST TxResult = "tecNO_DST"

	// The account on the receiving end does not exist, and the transaction does not send enough XRP to create it.
	TecNO_DST_INSUF_XRP TxResult = "tecNO_DST_INSUF_XRP"

	// The transaction tried to modify a ledger object, but the specified object does not exist.
	TecNO_ENTRY TxResult = "tecNO_ENTRY"

	// The account specified as issuer of a currency does not exist.
	TecNO_ISSUER TxResult = "tecNO_ISSUER"

	// The TakerPays field specifies an asset requiring authorization, but the account lacks a trust line.
	TecNO_LINE TxResult = "tecNO_LINE"

	// Insufficient XRP reserve to create a new trust line; the counterparty has no trust line.
	TecNO_LINE_INSUF_RESERVE TxResult = "tecNO_LINE_INSUF_RESERVE"

	// The transaction attempted to set a trust line to its default state, but it didn't exist.
	TecNO_LINE_REDUNDANT TxResult = "tecNO_LINE_REDUNDANT"

	// The sender lacks permission to execute the requested operation.
	TecNO_PERMISSION TxResult = "tecNO_PERMISSION"

	// Attempted to disable the master key, but no alternative method exists.
	TecNO_REGULAR_KEY TxResult = "tecNO_REGULAR_KEY"

	// No available directory page to hold the minted/acquired NFToken. (Added by the NonFungibleTokensV1_1 amendment)
	TecNO_SUITABLE_NFTOKEN_PAGE TxResult = "tecNO_SUITABLE_NFTOKEN_PAGE"

	// The referenced Escrow or PayChannel ledger object doesn't exist or has already been deleted.
	TecNO_TARGET TxResult = "tecNO_TARGET"

	// An object specified in this transaction did not exist in the ledger. (Added by NonFungibleTokensV1_1 amendment)
	TecOBJECT_NOT_FOUND TxResult = "tecOBJECT_NOT_FOUND"

	// The transaction created excessively large metadata.
	TecOVERSIZE TxResult = "tecOVERSIZE"

	// The transaction cannot succeed because the sender already owns ledger objects.
	TecOWNERS TxResult = "tecOWNERS"

	// Provided paths lack enough liquidity to send any amount at all.
	TecPATH_DRY TxResult = "tecPATH_DRY"

	// Provided paths lack enough liquidity to send the full requested amount.
	TecPATH_PARTIAL TxResult = "tecPATH_PARTIAL"

	// Account deletion failed due to a sequence number being too recent.
	TecTOO_SOON TxResult = "tecTOO_SOON"

	// Insufficient XRP to cover the transaction amount plus reserve requirements.
	TecUNFUNDED TxResult = "tecUNFUNDED"

	// DEPRECATED.
	TecUNFUNDED_ADD TxResult = "tecUNFUNDED_ADD"

	// Attempted payment exceeds the sender’s XRP holdings.
	TecUNFUNDED_PAYMENT TxResult = "tecUNFUNDED_PAYMENT"

	// Offer creation failed due to lack of the TakerGets currency.
	TecUNFUNDED_OFFER TxResult = "tecUNFUNDED_OFFER"

	// ------------------------------------------------------------------------------------------------
	// tef codes ⬇️ - https://xrpl.org/docs/references/protocol/transactions/transaction-results/tef-codes
	//
	// These codes indicate that the transaction failed and was not included in a ledger, but the transaction could have succeeded in some theoretical ledger.
	// Typically this means that the transaction can no longer succeed in any future ledger. They have numerical values in the range -199 to -100. The exact code for any given error is subject to change, so don't rely on it.
	// ------------------------------------------------------------------------------------------------

	// The sequence number of the transaction is lower than the current sequence number of the account sending the transaction.
	//revive:disable-next-line:var-naming
	// The same exact transaction has already been applied.
	TefALREADY TxResult = "tefALREADY"

	// DEPRECATED.
	TefBAD_ADD_AUTH TxResult = "tefBAD_ADD_AUTH"

	// The key used to sign this account is not authorized to modify this account.
	TefBAD_AUTH TxResult = "tefBAD_AUTH"

	// The single signature provided to authorize this transaction does not match the master key, but no regular key is associated with this address.
	TefBAD_AUTH_MASTER TxResult = "tefBAD_AUTH_MASTER"

	// The ledger was discovered in an unexpected state while processing the transaction.
	TefBAD_LEDGER TxResult = "tefBAD_LEDGER"

	// The transaction was multi-signed, but signatures did not meet quorum.
	TefBAD_QUORUM TxResult = "tefBAD_QUORUM"

	// The transaction was multi-signed but included a signature not part of the SignerList.
	TefBAD_SIGNATURE TxResult = "tefBAD_SIGNATURE"

	// DEPRECATED.
	TefCREATED TxResult = "tefCREATED"

	// The server entered an unexpected state processing the transaction.
	TefEXCEPTION TxResult = "tefEXCEPTION"

	// Unspecified failure in applying the transaction.
	TefFAILURE TxResult = "tefFAILURE"

	// The server entered an unexpected internal state when applying the transaction.
	TefINTERNAL TxResult = "tefINTERNAL"

	// An invariant check failed when claiming the transaction cost.
	TefINVARIANT_FAILED TxResult = "tefINVARIANT_FAILED"

	// The transaction was signed with the master key, but the account has master key disabled.
	TefMASTER_DISABLED TxResult = "tefMASTER_DISABLED"

	// Transaction included LastLedgerSequence, but the current ledger sequence number exceeds it.
	TefMAX_LEDGER TxResult = "tefMAX_LEDGER"

	// Attempted transfer of non-transferable NFToken.
	TefNFTOKEN_IS_NOT_TRANSFERABLE TxResult = "tefNFTOKEN_IS_NOT_TRANSFERABLE"

	// TrustSet transaction tried to authorize a trust line unnecessarily.
	TefNO_AUTH_REQUIRED TxResult = "tefNO_AUTH_REQUIRED"

	// The transaction attempted to use a Ticket that cannot exist.
	TefNO_TICKET TxResult = "tefNO_TICKET"

	// The transaction was multi-signed, but the sending account has no SignerList.
	TefNOT_MULTI_SIGNING TxResult = "tefNOT_MULTI_SIGNING"

	// The transaction's sequence number is lower than the account's current sequence number.
	TefPAST_SEQ TxResult = "tefPAST_SEQ"

	// The transaction would affect too many objects in the ledger.
	TefTOO_BIG TxResult = "tefTOO_BIG"

	// AccountTxnID does not match the account's previous transaction.
	TefWRONG_PRIOR TxResult = "tefWRONG_PRIOR"

	// ------------------------------------------------------------------------------------------------
	// tel codes ⬇️ - https://xrpl.org/docs/references/protocol/transactions/transaction-results/tel-codes
	// These codes indicate an error in the local server processing the transaction; it is possible that another server with a different configuration or load level could process the transaction successfully.
	// They have numerical values in the range -399 to -300. The exact code for any given error is subject to change, so don't rely on it.
	// ------------------------------------------------------------------------------------------------

	// The domain value specified by the transaction is invalid or too long.
	TelBAD_DOMAIN TxResult = "telBAD_DOMAIN"

	// Transaction contains too many paths to process.
	TelBAD_PATH_COUNT TxResult = "telBAD_PATH_COUNT"

	// The public key value specified by the transaction is invalid or incorrect in length.
	TelBAD_PUBLIC_KEY TxResult = "telBAD_PUBLIC_KEY"

	// Transaction not queued due to queuing restrictions.
	TelCAN_NOT_QUEUE TxResult = "telCAN_NOT_QUEUE"

	// Transaction not queued because potential costs exceed account balance.
	TelCAN_NOT_QUEUE_BALANCE TxResult = "telCAN_NOT_QUEUE_BALANCE"

	// Transaction not queued because it would block existing queued transactions.
	TelCAN_NOT_QUEUE_BLOCKS TxResult = "telCAN_NOT_QUEUE_BLOCKS"

	// Transaction not queued because it is blocked by transactions ahead of it.
	TelCAN_NOT_QUEUE_BLOCKED TxResult = "telCAN_NOT_QUEUE_BLOCKED"

	// Transaction not queued due to insufficient fee increase.
	TelCAN_NOT_QUEUE_FEE TxResult = "telCAN_NOT_QUEUE_FEE"

	// Transaction not queued because the transaction queue is full.
	TelCAN_NOT_QUEUE_FULL TxResult = "telCAN_NOT_QUEUE_FULL"

	// An unspecified error occurred processing the transaction.
	TelFAILED_PROCESSING TxResult = "telFAILED_PROCESSING"

	// Fee insufficient based on current server load.
	TelINSUF_FEE_P TxResult = "telINSUF_FEE_P"

	// Unspecified local error occurred.
	TelLOCAL_ERROR TxResult = "telLOCAL_ERROR"

	// NetworkID field specified incorrectly based on current network rules.
	TelNETWORK_ID_MAKES_TX_NON_CANONICAL TxResult = "telNETWORK_ID_MAKES_TX_NON_CANONICAL"

	// TfPartialPayment improperly used in an XRP payment funding a new account.
	TelNO_DST_PARTIAL TxResult = "telNO_DST_PARTIAL"

	// Transaction missing required NetworkID field.
	TelREQUIRES_NETWORK_ID TxResult = "telREQUIRES_NETWORK_ID"

	// Transaction specifies incorrect NetworkID value for the current network.
	TelWRONG_NETWORK TxResult = "telWRONG_NETWORK"

	// ------------------------------------------------------------------------------------------------
	// tem codes ⬇️ - https://xrpl.org/docs/references/protocol/transactions/transaction-results/tem-codes
	//
	// These codes indicate that the transaction was malformed, and cannot succeed according to the XRP Ledger protocol.
	// They have numerical values in the range -299 to -200. The exact code for any given error is subject to change, so don't rely on it.
	// ------------------------------------------------------------------------------------------------

	// The transaction incorrectly specified one or more assets. (Added by the AMM amendment)
	TemBAD_AMM_TOKENS TxResult = "temBAD_AMM_TOKENS"

	// An amount specified by the transaction was invalid, possibly negative.
	TemBAD_AMOUNT TxResult = "temBAD_AMOUNT"

	// Key used for signing doesn't match master key, and no Regular Key is set.
	TemBAD_AUTH_MASTER TxResult = "temBAD_AUTH_MASTER"

	// Currency field improperly specified.
	TemBAD_CURRENCY TxResult = "temBAD_CURRENCY"

	// Expiration value improperly specified or missing.
	TemBAD_EXPIRATION TxResult = "temBAD_EXPIRATION"

	// Fee value improperly specified.
	TemBAD_FEE TxResult = "temBAD_FEE"

	// Issuer field improperly specified.
	TemBAD_ISSUER TxResult = "temBAD_ISSUER"

	// LimitAmount value improperly specified.
	TemBAD_LIMIT TxResult = "temBAD_LIMIT"

	// NFTokenMint TransferFee improperly specified. (Added by NonFungibleTokensV1_1 amendment)
	TemBAD_NFTOKEN_TRANSFER_FEE TxResult = "temBAD_NFTOKEN_TRANSFER_FEE"

	// Invalid offer specified in OfferCreate.
	TemBAD_OFFER TxResult = "temBAD_OFFER"

	// Payment Paths improperly specified.
	TemBAD_PATH TxResult = "temBAD_PATH"

	// Payment Paths flagged as a loop.
	TemBAD_PATH_LOOP TxResult = "temBAD_PATH_LOOP"

	// TfLimitQuality improperly used in direct XRP-to-XRP payment.
	TemBAD_SEND_XRP_LIMIT TxResult = "temBAD_SEND_XRP_LIMIT"

	// SendMax improperly included in direct XRP-to-XRP payment.
	TemBAD_SEND_XRP_MAX TxResult = "temBAD_SEND_XRP_MAX"

	// TfRippleNotDirect improperly used in direct XRP-to-XRP payment.
	TemBAD_SEND_XRP_NO_DIRECT TxResult = "temBAD_SEND_XRP_NO_DIRECT"

	// TfPartialPayment improperly used in direct XRP-to-XRP payment.
	TemBAD_SEND_XRP_PARTIAL TxResult = "temBAD_SEND_XRP_PARTIAL"

	// Paths improperly included in direct XRP-to-XRP payment.
	TemBAD_SEND_XRP_PATHS TxResult = "temBAD_SEND_XRP_PATHS"

	// Sequence number references a future transaction.
	TemBAD_SEQUENCE TxResult = "temBAD_SEQUENCE"

	// Signature missing or improperly formed.
	TemBAD_SIGNATURE TxResult = "temBAD_SIGNATURE"

	// Source account address improperly formed.
	TemBAD_SRC_ACCOUNT TxResult = "temBAD_SRC_ACCOUNT"

	// TransferRate field improperly formatted.
	TemBAD_TRANSFER_RATE TxResult = "temBAD_TRANSFER_RATE"

	// DepositPreauth transaction attempted to preauthorize self.
	TemCANNOT_PREAUTH_SELF TxResult = "temCANNOT_PREAUTH_SELF"

	// Destination address matches sending account.
	TemDST_IS_SRC TxResult = "temDST_IS_SRC"

	// Transaction improperly omitted destination.
	TemDST_NEEDED TxResult = "temDST_NEEDED"

	// Transaction is otherwise invalid.
	TemINVALID TxResult = "temINVALID"

	// TicketCount field specifies invalid number of tickets.
	TemINVALID_COUNT TxResult = "temINVALID_COUNT"

	// Transaction includes invalid or contradictory Flag.
	TemINVALID_FLAG TxResult = "temINVALID_FLAG"

	// Unspecified problem with transaction format.
	TemMALFORMED TxResult = "temMALFORMED"

	// Transaction would accomplish nothing.
	TemREDUNDANT TxResult = "temREDUNDANT"

	// Removed in: rippled 0.28.0.
	TemREDUNDANT_SEND_MAX TxResult = "temREDUNDANT_SEND_MAX"

	// Payment transaction includes empty Paths.
	TemRIPPLE_EMPTY TxResult = "temRIPPLE_EMPTY"

	// Invalid SignerWeight in SignerListSet.
	TemBAD_WEIGHT TxResult = "temBAD_WEIGHT"

	// Invalid signer in SignerListSet.
	TemBAD_SIGNER TxResult = "temBAD_SIGNER"

	// Invalid SignerQuorum value in SignerListSet.
	TemBAD_QUORUM TxResult = "temBAD_QUORUM"

	// Used internally only, never returned.
	TemUNCERTAIN TxResult = "temUNCERTAIN"

	// Used internally only, never returned.
	TemUNKNOWN TxResult = "temUNKNOWN"

	// Transaction requires disabled logic or amendment.
	TemDISABLED TxResult = "temDISABLED"

	// ------------------------------------------------------------------------------------------------
	// ter codes ⬇️ - https://xrpl.org/docs/references/protocol/transactions/transaction-results/ter-codes
	//
	// These codes indicate that the transaction has not been applied yet, and generally will be automatically retried by the server that returned the result code. The transaction could apply successfully in the future; for example, if a certain other transaction applies first.
	// These codes have numerical values in the range -99 to -1, but the exact code for any given error is subject to change, so don't rely on it.
	// ------------------------------------------------------------------------------------------------

	// DEPRECATED.
	TerFUNDS_SPENT TxResult = "terFUNDS_SPENT"

	// Sending account has insufficient XRP to pay the specified fee.
	TerINSUF_FEE_B TxResult = "terINSUF_FEE_B"

	// Used internally only, never returned.
	TerLAST TxResult = "terLAST"

	// Sending address is not yet funded in the ledger.
	TerNO_ACCOUNT TxResult = "terNO_ACCOUNT"

	// Asset pair specified does not have an existing AMM instance. (Added by AMM amendment)
	TerNO_AMM TxResult = "terNO_AMM"

	// Attempted to add unauthorized currency to trust line.
	TerNO_AUTH TxResult = "terNO_AUTH"

	// Used internally only, never returned.
	TerNO_LINE TxResult = "terNO_LINE"

	// Transaction cannot succeed due to rippling settings.
	TerNO_RIPPLE TxResult = "terNO_RIPPLE"

	// Transaction requires sender to have nonzero owners count.
	TerOWNERS TxResult = "terOWNERS"

	// Sequence number is higher than the sender's current sequence number.
	TerPRE_SEQ TxResult = "terPRE_SEQ"

	// Attempted to use Ticket not yet existing in the ledger, though it could still be created.
	TerPRE_TICKET TxResult = "terPRE_TICKET"

	// Transaction met load-scaled cost but queued for future ledger due to open ledger cost.
	TerQUEUED TxResult = "terQUEUED"

	// Unspecified retriable error.
	TerRETRY TxResult = "terRETRY"

	// Transaction submitted but not yet applied.
	TerSUBMITTED TxResult = "terSUBMITTED"

	// ------------------------------------------------------------------------------------------------
	// Success results - https://xrpl.org/docs/references/protocol/transactions/transaction-results/tes-success
	// ------------------------------------------------------------------------------------------------

	// The transaction was applied and forwarded to other servers.
	// If this appears in a validated ledger, then the transaction's success is final.
	TesSUCCESS TxResult = "tesSUCCESS"
)

//revive:enable:var-naming,exported

// String returns the string representation of the result
func (t TxResult) String() string {
	return string(t)
}

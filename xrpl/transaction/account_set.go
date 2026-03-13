package transaction

import (
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	//
	// Account Set Flags
	//

	// AsfRequireDest requires a destination tag to send transactions to this account.
	AsfRequireDest uint32 = 1
	// AsfRequireAuth requires authorization for users to hold balances issued by this address.
	// Can only be enabled if the address has no trust lines connected to it.
	AsfRequireAuth uint32 = 2
	// AsfDisallowXRP indicates XRP should not be sent to this account.
	AsfDisallowXRP uint32 = 3
	// AsfDisableMaster disallows use of the master key pair. Can only be enabled if the account
	// has configured another way to sign transactions, such as a Regular Key or a
	// Signer List.
	AsfDisableMaster uint32 = 4
	// AsfAccountTxnID tracks the ID of this account's most recent transaction. Required for
	// AccountTxnID.
	AsfAccountTxnID uint32 = 5
	// AsfNoFreeze permanently gives up the ability to freeze individual trust lines or
	// disable Global Freeze. This flag can never be disabled after being enabled.
	AsfNoFreeze uint32 = 6
	// AsfGlobalFreeze freezes all assets issued by this account.
	AsfGlobalFreeze uint32 = 7
	// AsfDefaultRipple enables rippling on this account's trust lines by default.
	AsfDefaultRipple uint32 = 8
	// AsfDepositAuth enables Deposit Authorization on this account.
	AsfDepositAuth uint32 = 9
	// AsfAuthorizedNFTokenMinter allows another account to mint and burn tokens on behalf of this account.
	// To remove an authorized minter, enable this flag and omit the NFTokenMinter field.
	AsfAuthorizedNFTokenMinter uint32 = 10
	// AsfDisallowIncomingNFTokenOffer disallows other accounts from creating incoming NFTOffers (note: asf 11 is reserved for Hooks amendment)
	AsfDisallowIncomingNFTokenOffer uint32 = 12
	// AsfDisallowIncomingCheck disallows other accounts from creating incoming Checks
	AsfDisallowIncomingCheck uint32 = 13
	// AsfDisallowIncomingPayChan disallows other accounts from creating incoming Payment Channels
	AsfDisallowIncomingPayChan uint32 = 14
	// AsfDisallowIncomingTrustLine disallows other accounts from creating incoming TrustLines
	AsfDisallowIncomingTrustLine uint32 = 15
	// AsfAllowTrustLineClawback permanently gains the ability to claw back issued IOUs
	AsfAllowTrustLineClawback uint32 = 16
	// AsfAllowTrustLineLocking allows issuers to use their IOUs as escrow amounts
	AsfAllowTrustLineLocking uint32 = 17

	//
	// Transaction Flags
	//

	// TfRequireDestTag the same as SetFlag: AsfRequireDest.
	TfRequireDestTag uint32 = 65536 // 0x00010000
	// TfOptionalDestTag the same as ClearFlag: AsfRequireDestTag.
	TfOptionalDestTag uint32 = 131072 // 0x00020000
	// TfRequireAuth the same as SetFlag: AsfRequireAuth.
	TfRequireAuth uint32 = 262144 // 0x00040000
	// TfOptionalAuth the same as ClearFlag: AsfRequireAuth.
	TfOptionalAuth uint32 = 524288 // 0x00080000
	// TfDisallowXRP the same as SetFlag: AsfDisallowXRP.
	TfDisallowXRP uint32 = 1048576 // 0x00100000
	// TfAllowXRP the same as ClearFlag: AsfDisallowXRP.
	TfAllowXRP uint32 = 2097152 // 0x00200000

	// MinTickSize is the minimum tick size to use for offers involving a currency issued by this address.
	// Valid values are 3 to 15 inclusive, or 0 to disable.
	MinTickSize = 3

	// MaxTickSize is the maximum tick size to use for offers involving a currency issued by this address.
	// Valid values are 3 to 15 inclusive, or 0 to disable.
	MaxTickSize = 15
)

// An AccountSet transaction modifies the properties of an account in the XRP
// Ledger.
type AccountSet struct {
	BaseTx
	// ClearFlag: AsfRequireDestTag, AsfOptionalDestTag, AsfRequireAuth, AsfOptionalAuth, AsfDisallowXRP, AsfAllowXRP
	ClearFlag uint32 `json:",omitempty"`
	// The domain that owns this account, as a string of hex representing the.
	// ASCII for the domain in lowercase.
	Domain *string `json:",omitempty"`
	// An arbitrary 128-bit value. Conventionally, clients treat this as the md5 hash of an email address to use for displaying a Gravatar image.
	EmailHash *types.Hash128 `json:",omitempty"`
	// Public key for sending encrypted messages to this account.
	MessageKey *string `json:",omitempty"`
	// Sets an alternate account that is allowed to mint NFTokens on this
	// account's behalf using NFTokenMint's `Issuer` field.
	NFTokenMinter *string `json:",omitempty"`
	// Integer flag to enable for this account.
	SetFlag uint32 `json:",omitempty"`
	// The fee to charge when users transfer this account's issued currencies,
	// represented as billionths of a unit. Cannot be more than 2000000000 or less
	// than 1000000000, except for the special case 0 meaning no fee.
	TransferRate *uint32 `json:",omitempty"`
	// Tick size to use for offers involving a currency issued by this address.
	// The exchange rates of those offers is rounded to this many significant
	// digits. Valid values are 3 to 15 inclusive, or 0 to disable.
	TickSize *uint8 `json:",omitempty"`
	// (Optional) An arbitrary 256-bit value. If specified, the value is stored as
	// part of the account but has no inherent meaning or requirements.
	WalletLocator *types.Hash256 `json:",omitempty"`
	// (Optional) Not used. This field is valid in AccountSet transactions but does nothing.
	WalletSize *uint32 `json:",omitempty"`
}

// TxType returns the type of the transaction (AccountSet).
func (*AccountSet) TxType() TxType {
	return AccountSetTx
}

// Flatten returns the flattened map of the AccountSet transaction.
func (s *AccountSet) Flatten() FlatTransaction {
	flattened := s.BaseTx.Flatten()

	flattened["TransactionType"] = "AccountSet"

	if s.ClearFlag != 0 {
		flattened["ClearFlag"] = s.ClearFlag
	}
	if s.Domain != nil {
		flattened["Domain"] = *s.Domain
	}
	if s.EmailHash != nil {
		flattened["EmailHash"] = s.EmailHash.String()
	}
	if s.MessageKey != nil {
		flattened["MessageKey"] = *s.MessageKey
	}
	if s.NFTokenMinter != nil {
		flattened["NFTokenMinter"] = *s.NFTokenMinter
	}
	if s.SetFlag != 0 {
		flattened["SetFlag"] = s.SetFlag
	}
	if s.TransferRate != nil {
		flattened["TransferRate"] = *s.TransferRate
	}
	if s.TickSize != nil {
		flattened["TickSize"] = *s.TickSize
	}
	if s.WalletLocator != nil {
		flattened["WalletLocator"] = s.WalletLocator.String()
	}
	if s.WalletSize != nil {
		flattened["WalletSize"] = *s.WalletSize
	}

	return flattened
}

// -----------------------------------
// -------------- FLAGS --------------
// -----------------------------------

// SetRequireDestTag sets the require destination tag flag.
func (s *AccountSet) SetRequireDestTag() {
	s.Flags |= TfRequireDestTag
}

// SetRequireAuth sets the require auth flag.
func (s *AccountSet) SetRequireAuth() {
	s.Flags |= TfRequireAuth
}

// SetDisallowXRP sets the disallow XRP flag.
func (s *AccountSet) SetDisallowXRP() {
	s.Flags |= TfDisallowXRP
}

// SetOptionalDestTag sets the optional destination tag flag.
func (s *AccountSet) SetOptionalDestTag() {
	s.Flags |= TfOptionalDestTag
}

// SetOptionalAuth sets the optional auth flag.
func (s *AccountSet) SetOptionalAuth() {
	s.Flags |= TfOptionalAuth
}

// SetAllowXRP sets the allow XRP flag.
func (s *AccountSet) SetAllowXRP() {
	s.Flags |= TfAllowXRP
}

// SetAsfRequireDest sets the require destination tag flag.
func (s *AccountSet) SetAsfRequireDest() {
	s.SetFlag = AsfRequireDest
}

// ClearAsfRequireDest clears the require destination tag flag.
func (s *AccountSet) ClearAsfRequireDest() {
	s.ClearFlag = AsfRequireDest
}

// SetAsfRequireAuth sets the require authorization flag.
func (s *AccountSet) SetAsfRequireAuth() {
	s.SetFlag = AsfRequireAuth
}

// ClearAsfRequireAuth clears the require authorization flag.
func (s *AccountSet) ClearAsfRequireAuth() {
	s.ClearFlag = AsfRequireAuth
}

// SetAsfDisallowXRP sets the disallow XRP flag.
func (s *AccountSet) SetAsfDisallowXRP() {
	s.SetFlag = AsfDisallowXRP
}

// ClearAsfDisallowXRP clears the disallow XRP flag.
func (s *AccountSet) ClearAsfDisallowXRP() {
	s.ClearFlag = AsfDisallowXRP
}

// SetAsfDisableMaster sets the disable master key flag.
func (s *AccountSet) SetAsfDisableMaster() {
	s.SetFlag = AsfDisableMaster
}

// ClearAsfDisableMaster clears the disable master key flag.
func (s *AccountSet) ClearAsfDisableMaster() {
	s.ClearFlag = AsfDisableMaster
}

// SetAsfAccountTxnID sets the account transaction ID flag.
func (s *AccountSet) SetAsfAccountTxnID() {
	s.SetFlag = AsfAccountTxnID
}

// ClearAsfAccountTxnID clears the account transaction ID flag.
func (s *AccountSet) ClearAsfAccountTxnID() {
	s.ClearFlag = AsfAccountTxnID
}

// SetAsfNoFreeze sets the no freeze flag.
func (s *AccountSet) SetAsfNoFreeze() {
	s.SetFlag = AsfNoFreeze
}

// ClearAsfNoFreeze clears the no freeze flag.
func (s *AccountSet) ClearAsfNoFreeze() {
	s.ClearFlag = AsfNoFreeze
}

// SetAsfGlobalFreeze sets the global freeze flag.
func (s *AccountSet) SetAsfGlobalFreeze() {
	s.SetFlag = AsfGlobalFreeze
}

// ClearAsfGlobalFreeze clears the global freeze flag.
func (s *AccountSet) ClearAsfGlobalFreeze() {
	s.ClearFlag = AsfGlobalFreeze
}

// SetAsfDefaultRipple sets the default ripple flag.
func (s *AccountSet) SetAsfDefaultRipple() {
	s.SetFlag = AsfDefaultRipple
}

// ClearAsfDefaultRipple clears the default ripple flag.
func (s *AccountSet) ClearAsfDefaultRipple() {
	s.ClearFlag = AsfDefaultRipple
}

// SetAsfDepositAuth sets the deposit authorization flag.
func (s *AccountSet) SetAsfDepositAuth() {
	s.SetFlag = AsfDepositAuth
}

// ClearAsfDepositAuth clears the deposit authorization flag.
func (s *AccountSet) ClearAsfDepositAuth() {
	s.ClearFlag = AsfDepositAuth
}

// SetAsfAuthorizedNFTokenMinter sets the authorized NFToken minter flag.
func (s *AccountSet) SetAsfAuthorizedNFTokenMinter() {
	s.SetFlag = AsfAuthorizedNFTokenMinter
}

// ClearAsfAuthorizedNFTokenMinter clears the authorized NFToken minter flag.
func (s *AccountSet) ClearAsfAuthorizedNFTokenMinter() {
	s.ClearFlag = AsfAuthorizedNFTokenMinter
}

// SetAsfDisallowIncomingNFTokenOffer sets the disallow incoming NFToken offer flag.
func (s *AccountSet) SetAsfDisallowIncomingNFTokenOffer() {
	s.SetFlag = AsfDisallowIncomingNFTokenOffer
}

// ClearAsfDisallowIncomingNFTokenOffer clears the disallow incoming NFToken offer flag.
func (s *AccountSet) ClearAsfDisallowIncomingNFTokenOffer() {
	s.ClearFlag = AsfDisallowIncomingNFTokenOffer
}

// SetAsfDisallowIncomingCheck sets the disallow incoming check flag.
func (s *AccountSet) SetAsfDisallowIncomingCheck() {
	s.SetFlag = AsfDisallowIncomingCheck
}

// ClearAsfDisallowIncomingCheck clears the disallow incoming check flag.
func (s *AccountSet) ClearAsfDisallowIncomingCheck() {
	s.ClearFlag = AsfDisallowIncomingCheck
}

// SetAsfDisallowIncomingPayChan sets the disallow incoming payment channel flag.
func (s *AccountSet) SetAsfDisallowIncomingPayChan() {
	s.SetFlag = AsfDisallowIncomingPayChan
}

// ClearAsfDisallowIncomingPayChan clears the disallow incoming payment channel flag.
func (s *AccountSet) ClearAsfDisallowIncomingPayChan() {
	s.ClearFlag = AsfDisallowIncomingPayChan
}

// SetAsfDisallowIncomingTrustLine sets the disallow incoming trust line flag.
func (s *AccountSet) SetAsfDisallowIncomingTrustLine() {
	s.SetFlag = AsfDisallowIncomingTrustLine
}

// ClearAsfDisallowIncomingTrustLine clears the disallow incoming trust line flag.
func (s *AccountSet) ClearAsfDisallowIncomingTrustLine() {
	s.ClearFlag = AsfDisallowIncomingTrustLine
}

// SetAsfAllowTrustLineClawback sets the allow trust line clawback flag.
func (s *AccountSet) SetAsfAllowTrustLineClawback() {
	s.SetFlag = AsfAllowTrustLineClawback
}

// ClearAsfAllowTrustLineClawback clears the allow trust line clawback flag.
func (s *AccountSet) ClearAsfAllowTrustLineClawback() {
	s.ClearFlag = AsfAllowTrustLineClawback
}

// SetAsfAllowTrustLineLocking sets the allow trust line locking flag.
func (s *AccountSet) SetAsfAllowTrustLineLocking() {
	s.SetFlag = AsfAllowTrustLineLocking
}

// ClearAsfAllowTrustLineLocking clears the allow trust line locking flag.
func (s *AccountSet) ClearAsfAllowTrustLineLocking() {
	s.ClearFlag = AsfAllowTrustLineLocking
}

// Validate the AccountSet transaction fields.
func (s *AccountSet) Validate() (bool, error) {
	flatten := s.Flatten()

	// validate the base transaction
	_, err := s.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "ClearFlag", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "Domain", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "EmailHash", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "MessageKey", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "SetFlag", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "TransferRate", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "TickSize", typecheck.IsUint8)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "NFTokenMinter", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "WalletLocator", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flatten, "WalletSize", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	// check if SetFlag is within the valid range
	if s.SetFlag != 0 {
		if s.SetFlag < AsfRequireDest || s.SetFlag > AsfAllowTrustLineLocking {
			return false, ErrAccountSetInvalidSetFlag
		}
	}

	// check if TickSize is within the valid range
	if s.TickSize != nil && *s.TickSize != 0 && (*s.TickSize < MinTickSize || *s.TickSize > MaxTickSize) {
		return false, ErrAccountSetInvalidTickSize
	}

	return true, nil
}

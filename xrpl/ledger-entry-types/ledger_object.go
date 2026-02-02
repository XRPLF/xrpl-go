package ledger

// EntryType represents the type of a ledger entry as a string identifier.
type EntryType string

// EntryType constants define all supported ledger entry types.
const (
	AccountRootEntry                     EntryType = "AccountRoot"
	AmendmentsEntry                      EntryType = "Amendments"
	AMMEntry                             EntryType = "AMM"
	BridgeEntry                          EntryType = "Bridge"
	CheckEntry                           EntryType = "Check"
	CredentialEntry                      EntryType = "Credential"
	DelegateEntry                        EntryType = "Delegate"
	DepositPreauthObjEntry               EntryType = "DepositPreauth"
	DIDEntry                             EntryType = "DID"
	DirectoryNodeEntry                   EntryType = "DirectoryNode"
	EscrowEntry                          EntryType = "Escrow"
	FeeSettingsEntry                     EntryType = "FeeSettings"
	LedgerHashesEntry                    EntryType = "LedgerHashes"
	LoanEntry                            EntryType = "Loan"
	LoanBrokerEntry                      EntryType = "LoanBroker"
	MPTokenEntry                         EntryType = "MPToken"
	MPTokenIssuanceEntry                 EntryType = "MPTokenIssuance" // #nosec G101
	NegativeUNLEntry                     EntryType = "NegativeUNL"
	NFTokenOfferEntry                    EntryType = "NFTokenOffer"
	NFTokenPageEntry                     EntryType = "NFTokenPage"
	OfferEntry                           EntryType = "Offer"
	OracleEntry                          EntryType = "Oracle"
	PayChannelEntry                      EntryType = "PayChannel"
	PermissionedDomainEntry              EntryType = "PermissionedDomain"
	RippleStateEntry                     EntryType = "RippleState"
	SignerListEntry                      EntryType = "SignerList"
	TicketEntry                          EntryType = "Ticket"
	XChainOwnedClaimIDEntry              EntryType = "XChainOwnedClaimID"
	XChainOwnedCreateAccountClaimIDEntry EntryType = "XChainOwnedCreateAccountClaimID"
)

// FlatLedgerObject represents a generic ledger entry as a flat map of field names to values.
type FlatLedgerObject map[string]interface{}

// EntryType returns the LedgerEntryType string stored in this flat object.
func (f FlatLedgerObject) EntryType() EntryType {
	return EntryType(f["LedgerEntryType"].(string))
}

// Object represents a generic ledger entry object with an EntryType method.
type Object interface {
	EntryType() EntryType
}

// EmptyLedgerObject returns a new empty ledger object matching the given entry type string.
// Returns an error if the entry type is unrecognized.
func EmptyLedgerObject(t string) (Object, error) {
	switch EntryType(t) {
	case AccountRootEntry:
		return &AccountRoot{}, nil
	case AmendmentsEntry:
		return &Amendments{}, nil
	case AMMEntry:
		return &AMM{}, nil
	case BridgeEntry:
		return &Bridge{}, nil
	case CheckEntry:
		return &Check{}, nil
	case CredentialEntry:
		return &Credential{}, nil
	case DelegateEntry:
		return &Delegate{}, nil
	case DepositPreauthObjEntry:
		return &DepositPreauthObj{}, nil
	case DIDEntry:
		return &DID{}, nil
	case DirectoryNodeEntry:
		return &DirectoryNode{}, nil
	case EscrowEntry:
		return &Escrow{}, nil
	case FeeSettingsEntry:
		return &FeeSettings{}, nil
	case MPTokenEntry:
		return &MPToken{}, nil
	case MPTokenIssuanceEntry:
		return &MPTokenIssuance{}, nil
	case LedgerHashesEntry:
		return &Hashes{}, nil
	case LoanEntry:
		return &Loan{}, nil
	case LoanBrokerEntry:
		return &LoanBroker{}, nil
	case NegativeUNLEntry:
		return &NegativeUNL{}, nil
	case NFTokenOfferEntry:
		return &NFTokenOffer{}, nil
	case NFTokenPageEntry:
		return &NFTokenPage{}, nil
	case OfferEntry:
		return &Offer{}, nil
	case OracleEntry:
		return &Oracle{}, nil
	case PayChannelEntry:
		return &PayChannel{}, nil
	case PermissionedDomainEntry:
		return &PermissionedDomain{}, nil
	case RippleStateEntry:
		return &RippleState{}, nil
	case SignerListEntry:
		return &SignerList{}, nil
	case TicketEntry:
		return &Ticket{}, nil
	case XChainOwnedClaimIDEntry:
		return &XChainOwnedClaimID{}, nil
	case XChainOwnedCreateAccountClaimIDEntry:
		return &XChainOwnedCreateAccountClaimID{}, nil
	}
	return nil, ErrUnrecognizedLedgerObjectType{
		Type: t,
	}
}

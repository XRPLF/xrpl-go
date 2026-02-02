# transaction

## Overview

As the [`queries`](/docs/xrpl/queries) package contains the types of XRPL queries, this package contains all transaction types available in the XRPL. Contains all transaction structs to build and sign transactions with wallets and clients.

## Transaction types

These are the transaction types available in the XRPL:

- [AccountDelete](https://xrpl.org/docs/references/protocol/transactions/types/accountdelete)
- [AccountSet](https://xrpl.org/docs/references/protocol/transactions/types/accountset)
- [AMMBid](https://xrpl.org/docs/references/protocol/transactions/types/ammbid)
- [AMMClawback](https://xrpl.org/docs/references/protocol/transactions/types/ammclawback)
- [AMMCreate](https://xrpl.org/docs/references/protocol/transactions/types/ammcreate)
- [AMMDelete](https://xrpl.org/docs/references/protocol/transactions/types/ammdelete)
- [AMMDeposit](https://xrpl.org/docs/references/protocol/transactions/types/ammdeposit)
- [AMMVote](https://xrpl.org/docs/references/protocol/transactions/types/ammvote)
- [AMMWithdraw](https://xrpl.org/docs/references/protocol/transactions/types/ammwithdraw)
- [Batch](https://xrpl.org/docs/references/protocol/transactions/types/batch)
- [CheckCancel](https://xrpl.org/docs/references/protocol/transactions/types/checkcancel)
- [CheckCash](https://xrpl.org/docs/references/protocol/transactions/types/checkcash)
- [CheckCreate](https://xrpl.org/docs/references/protocol/transactions/types/checkcreate)
- [Clawback](https://xrpl.org/docs/references/protocol/transactions/types/clawback)
- [CredentialAccept](https://xrpl.org/docs/references/protocol/transactions/types/credentialaccept)
- [CredentialCreate](https://xrpl.org/docs/references/protocol/transactions/types/credentialcreate)
- [CredentialDelete](https://xrpl.org/docs/references/protocol/transactions/types/credentialdelete)
- [DepositPreauth](https://xrpl.org/docs/references/protocol/transactions/types/depositpreauth)
- [DelegateSet](https://xrpl.org/docs/references/protocol/transactions/types/delegateset)
- [DIDDelete](https://xrpl.org/docs/references/protocol/transactions/types/diddelete)
- [DIDSet](https://xrpl.org/docs/references/protocol/transactions/types/didset)
- [LoanBrokerCoverClawback](https://xrpl.org/docs/references/protocol/transactions/types/loanbrokercoverclawback)
- [LoanBrokerCoverDeposit](https://xrpl.org/docs/references/protocol/transactions/types/loanbrokercoverdeposit)
- [LoanBrokerCoverWithdraw](https://xrpl.org/docs/references/protocol/transactions/types/loanbrokercoverwithdraw)
- [LoanBrokerDelete](https://xrpl.org/docs/references/protocol/transactions/types/loanbrokerdelete)
- [LoanBrokerSet](https://xrpl.org/docs/references/protocol/transactions/types/loanbrokerset)
- [LoanDelete](https://xrpl.org/docs/references/protocol/transactions/types/loandelete)
- [LoanManage](https://xrpl.org/docs/references/protocol/transactions/types/loanmanage)
- [LoanPay](https://xrpl.org/docs/references/protocol/transactions/types/loanpay)
- [LoanSet](https://xrpl.org/docs/references/protocol/transactions/types/loanset)
- [EscrowCancel](https://xrpl.org/docs/references/protocol/transactions/types/escrowcancel)
- [EscrowCreate](https://xrpl.org/docs/references/protocol/transactions/types/escrowcreate)
- [EscrowFinish](https://xrpl.org/docs/references/protocol/transactions/types/escrowfinish)
- [NFTokenAcceptOffer](https://xrpl.org/docs/references/protocol/transactions/types/nftokenacceptoffer)
- [NFTokenBurn](https://xrpl.org/docs/references/protocol/transactions/types/nftokenburn)
- [NFTokenCancelOffer](https://xrpl.org/docs/references/protocol/transactions/types/nftokencanceloffer)
- [NFTokenCreateOffer](https://xrpl.org/docs/references/protocol/transactions/types/nftokencreateoffer)
- [NFTokenMint](https://xrpl.org/docs/references/protocol/transactions/types/nftokenmint)
- [NFTokenModify](https://xrpl.org/docs/references/protocol/transactions/types/nftokenmodify)
- [OfferCancel](https://xrpl.org/docs/references/protocol/transactions/types/offercancel)
- [OfferCreate](https://xrpl.org/docs/references/protocol/transactions/types/offercreate)
- [OracleDelete](https://xrpl.org/docs/references/protocol/transactions/types/oracledelete)
- [OracleSet](https://xrpl.org/docs/references/protocol/transactions/types/oracleset)
- [PaymentChannelClaim](https://xrpl.org/docs/references/protocol/transactions/types/paymentchannelclaim)
- [PaymentChannelCreate](https://xrpl.org/docs/references/protocol/transactions/types/paymentchannelcreate)
- [PaymentChannelFund](https://xrpl.org/docs/references/protocol/transactions/types/paymentchannelfund)
- [PermissionedDomainDelete](https://xrpl.org/docs/references/protocol/transactions/types/permissioneddomaindelete)
- [PermissionedDomainSet](https://xrpl.org/docs/references/protocol/transactions/types/permissioneddomainset)
- [Payment](https://xrpl.org/docs/references/protocol/transactions/types/payment)
- [SetRegularKey](https://xrpl.org/docs/references/protocol/transactions/types/setregularkey)
- [SignerListSet](https://xrpl.org/docs/references/protocol/transactions/types/signerlistset)
- [TicketCreate](https://xrpl.org/docs/references/protocol/transactions/types/ticketcreate)
- [TrustSet](https://xrpl.org/docs/references/protocol/transactions/types/trustset)
- [MPTokenAuthorize](https://xrpl.org/docs/references/protocol/transactions/types/mptokenauthorize)
- [MPTokenIssuanceCreate](https://xrpl.org/docs/references/protocol/transactions/types/mptokenissuancecreate)
- [MPTokenIssuanceDestroy](https://xrpl.org/docs/references/protocol/transactions/types/mptokenissuancedestroy)
- [MPTokenIssuanceSet](https://xrpl.org/docs/references/protocol/transactions/types/mptokenissuanceset)
- [XChainAccountCreateCommit](https://xrpl.org/docs/references/protocol/transactions/types/xchainaccountcreatecommit)
- [XChainAddAccountCreateAttestation](https://xrpl.org/docs/references/protocol/transactions/types/xchainaddaccountcreateattestation)
- [XChainAddClaimAttestation](https://xrpl.org/docs/references/protocol/transactions/types/xchainaddclaimattestation)
- [XChainClaim](https://xrpl.org/docs/references/protocol/transactions/types/xchainclaim)
- [XChainCommit](https://xrpl.org/docs/references/protocol/transactions/types/xchaincommit)
- [XChainCreateBridge](https://xrpl.org/docs/references/protocol/transactions/types/xchaincreatebridge)
- [XChainCreateClaimID](https://xrpl.org/docs/references/protocol/transactions/types/xchaincreateclaimid)
- [XChainModifyBridge](https://xrpl.org/docs/references/protocol/transactions/types/xchainmodifybridge)

## MPTokenMetadata

The `MPTokenMetadata` type provides functionality to encode, decode, and validate metadata for Multi-Purpose Tokens (MPTs) as per the [XLS-89 standard](https://xls.xrpl.org/xls/XLS-0089-multi-purpose-token-metadata-schema.html). This metadata includes information about the token such as ticker, name, description, icon, asset classification, and related URIs.

### Overview

MPTokenMetadata is used in MPToken transactions (like `MPTokenIssuanceCreate`) to provide structured metadata about tokens. The metadata is encoded as a hex string and must comply with the XLS-89 standard, which includes:

- Maximum size limit of 1024 bytes
- Support for both long-form and compact-form JSON keys
- Validation of required fields, field types, and allowed values
- Alphabetical ordering of fields for consistent encoding

### Types

#### ParsedMPTokenMetadata

The `ParsedMPTokenMetadata` struct represents the complete metadata structure for an MPToken. Fields are ordered alphabetically by JSON key for consistent encoding.

```go
type ParsedMPTokenMetadata struct {
    // Top-level classification of token purpose (required)
    // Allowed values: "rwa", "memes", "wrapped", "gaming", "defi", "other"
    AssetClass string `json:"ac"`
    
    // Freeform field for key token details (optional)
    // Can be any valid JSON object or UTF-8 string
    AdditionalInfo any `json:"ai,omitempty"`
    
    // Optional subcategory of the asset class (optional, required if AssetClass is "rwa")
    // Allowed values: "stablecoin", "commodity", "real_estate", "private_credit", "equity", "treasury", "other"
    AssetSubclass *string `json:"as,omitempty"`
    
    // Short description of the token (optional)
    Desc *string `json:"d,omitempty"`
    
    // URI to the token icon (required)
    // Can be a hostname/path (HTTPS assumed) or full URI for other protocols (e.g., ipfs://)
    Icon string `json:"i"`
    
    // The name of the issuer account (required)
    IssuerName string `json:"in"`
    
    // Display name of the token (required)
    Name string `json:"n"`
    
    // Ticker symbol used to represent the token (required)
    // Uppercase letters (A-Z) and digits (0-9) only. Max 6 chars.
    Ticker string `json:"t"`
    
    // List of related URIs (optional)
    // Each URI object contains the link, its category, and a human-readable title
    URIs []ParsedMPTokenMetadataURI `json:"us,omitempty"`
}
```

**Field Requirements:**

- **Required fields**: `Ticker`, `Name`, `Icon`, `AssetClass`, `IssuerName`
- **Conditional**: `AssetSubclass` is required when `AssetClass` is `"rwa"`
- **Optional fields**: `Desc`, `URIs`, `AdditionalInfo`

#### ParsedMPTokenMetadataURI

The `ParsedMPTokenMetadataURI` struct represents a URI entry within the metadata. Fields are ordered alphabetically by JSON key for consistent encoding.

```go
type ParsedMPTokenMetadataURI struct {
    // The category of the link (required)
    // Allowed values: "website", "social", "docs", "other"
    Category string `json:"c"`
    
    // A human-readable label for the link (required)
    Title string `json:"t"`
    
    // URI to the related resource (required)
    // Can be a hostname/path (HTTPS assumed) or full URI for other protocols (e.g., ipfs://)
    URI string `json:"u"`
}
```

### Functions

#### EncodeMPTokenMetadata

Encodes a `ParsedMPTokenMetadata` struct into a hex string compliant with XLS-89.

```go
func EncodeMPTokenMetadata(meta ParsedMPTokenMetadata) (string, error)
```

**Returns:**
- `string`: The encoded hex string (uppercase)
- `error`: An error if encoding fails

#### DecodeMPTokenMetadata

Decodes a hex string into a `ParsedMPTokenMetadata` struct. Handles input with either long-form or compact-form keys via custom `UnmarshalJSON` methods.

```go
func DecodeMPTokenMetadata(hexInput string) (ParsedMPTokenMetadata, error)
```

**Returns:**
- `ParsedMPTokenMetadata`: The decoded metadata struct
- `error`: An error if decoding fails (e.g., invalid hex, invalid JSON)

#### ValidateMPTokenMetadata

Validates MPToken metadata according to the XLS-89 standard. Checks for:
- Valid hex string format
- Maximum size limit (1024 bytes)
- Valid JSON structure
- Required fields presence
- Field types and formats
- Allowed values for enums
- Field count limits

```go
func ValidateMPTokenMetadata(input string) error
```

**Returns:**
- `error`: `nil` if valid, or `MPTokenMetadataValidationErrors` containing all validation errors

### Constants

```go
// Maximum byte length for MPToken metadata (1024 bytes)
const MaxMPTokenMetadataByteLength = 1024

// Allowed values for the asset class field
var MPTokenMetadataAssetClasses = [6]string{"rwa", "memes", "wrapped", "gaming", "defi", "other"}

// Allowed values for the asset subclass field
var MPTokenMetadataAssetSubClasses = [7]string{"stablecoin", "commodity", "real_estate", "private_credit", "equity", "treasury", "other"}

// Allowed values for the URI category field
var MPTokenMetadataURICategories = [4]string{"website", "social", "docs", "other"}
```

### Usage Examples

#### Creating and Encoding Metadata

```go
package main

import (
    "fmt"
    "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func main() {
    // Create metadata struct
    assetSubclass := "treasury"
    desc := "A yield-bearing stablecoin backed by short-term U.S. Treasuries."
    
    metadata := types.ParsedMPTokenMetadata{
        Ticker:        "TBILL",
        Name:          "T-Bill Yield Token",
        Desc:          &desc,
        Icon:          "https://example.org/tbill-icon.png",
        AssetClass:    "rwa",
        AssetSubclass: &assetSubclass,
        IssuerName:    "Example Yield Co.",
        URIs: []types.ParsedMPTokenMetadataURI{
            {
                URI:      "https://exampleyield.co/tbill",
                Category: "website",
                Title:    "Product Page",
            },
            {
                URI:      "https://exampleyield.co/docs",
                Category: "docs",
                Title:    "Yield Token Docs",
            },
        },
        AdditionalInfo: map[string]any{
            "interest_rate": "5.00%",
            "maturity_date": "2045-06-30",
        },
    }
    
    // Encode to hex string
    hexStr, err := types.EncodeMPTokenMetadata(metadata)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Encoded metadata: %s\n", hexStr)
}
```

#### Decoding Metadata

```go
package main

import (
    "fmt"
    "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func main() {
    hexStr := "7B226163223A22727761222C226173223A227472656173757279222C..."
    
    // Decode from hex string
    metadata, err := types.DecodeMPTokenMetadata(hexStr)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Ticker: %s\n", metadata.Ticker)
    fmt.Printf("Name: %s\n", metadata.Name)
    fmt.Printf("Asset Class: %s\n", metadata.AssetClass)
}
```

#### Validating Metadata

```go
package main

import (
    "fmt"
    "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func main() {
    hexStr := "7B226163223A22727761222C226173223A227472656173757279222C..."
    
    // Validate metadata
    err := types.ValidateMPTokenMetadata(hexStr)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }
    
    fmt.Println("Metadata is valid!")
}
```

#### Using in MPTokenIssuanceCreate Transaction

```go
package main

import (
    "github.com/Peersyst/xrpl-go/xrpl/transaction"
    "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func main() {
    // Create and encode metadata
    metadata := types.ParsedMPTokenMetadata{
        Ticker:      "TBILL",
        Name:        "T-Bill Yield Token",
        Icon:        "https://example.org/tbill-icon.png",
        AssetClass:  "rwa",
        IssuerName:  "Example Yield Co.",
    }
    
    hexStr, _ := types.EncodeMPTokenMetadata(metadata)
    
    // Use in transaction
    tx := transaction.MPTokenIssuanceCreate{
        BaseTx: transaction.BaseTx{
            TransactionType: "MPTokenIssuanceCreate",
            Account:         "rajgkBmMxmz161r8bWYH7CQAFZP5bA9oSG",
        },
        AssetScale:     2,
        TransferFee:    314,
        MaximumAmount: "50000000",
        MPTokenMetadata: types.MPTokenMetadata(hexStr),
    }
    
    // ... continue with transaction signing and submission
}
```

### JSON Key Formats

The implementation supports both long-form and compact-form JSON keys for backward compatibility and flexibility:

**Long-form keys:**
- `ticker`, `name`, `desc`, `icon`, `asset_class`, `asset_subclass`, `issuer_name`, `uris`, `additional_info`
- `uri`, `category`, `title` (for URI objects)

**Compact-form keys:**
- `t`, `n`, `d`, `i`, `ac`, `as`, `in`, `us`, `ai`
- `u`, `c`, `t` (for URI objects)

Both formats are accepted when decoding, but encoding always uses compact-form keys for consistency and size efficiency.

### Import

To use MPTokenMetadata types and functions, import:

```go
import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
```

## Usage

To use the `transaction` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/transaction"
```

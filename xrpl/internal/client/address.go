package client

import (
	"fmt"
	"reflect"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

type addressChange struct {
	tx           transactionMap
	addressField string
	classic      string
	tagField     string
	tag          uint32
	hasTag       bool
}

var taglessAddressFields = [...]string{
	"Authorize",
	"Unauthorize",
	"Owner",
	"RegularKey",
	"Delegate",
	"NFTokenMinter",
	"Subject",
	"Issuer",
	"Holder",
}

// SetValidAddresses converts every in-scope X-address in an outer transaction
// and its Batch inner transactions to a classic address. Embedded Account and
// Destination tags are applied only after all conflicts have been validated.
func SetValidAddresses(tx transactionMap) error {
	changes, err := collectTransactionAddressChanges(tx)
	if err != nil {
		return err
	}

	for _, change := range changes {
		change.tx[change.addressField] = change.classic
		if change.hasTag {
			change.tx[change.tagField] = change.tag
		}
	}
	return nil
}

func collectTransactionAddressChanges(tx transactionMap) ([]addressChange, error) {
	// Reserve space for four common-case changes while allowing the slice to grow for larger transactions.
	changes := make([]addressChange, 0, 4)
	if err := appendTransactionAddressChanges(tx, &changes); err != nil {
		return nil, err
	}
	return changes, nil
}

func appendTransactionAddressChanges(tx transactionMap, changes *[]addressChange) error {
	if err := collectAddressChange(tx, "Account", "SourceTag", changes); err != nil {
		return err
	}
	if err := collectAddressChange(tx, "Destination", "DestinationTag", changes); err != nil {
		return err
	}

	for _, field := range taglessAddressFields {
		if err := collectAddressChange(tx, field, "", changes); err != nil {
			return err
		}
	}

	inners, err := batchInnerTransactions(tx)
	if err != nil {
		return err
	}
	for _, inner := range inners {
		if err := appendTransactionAddressChanges(inner, changes); err != nil {
			return err
		}
	}
	return nil
}

func collectAddressChange(tx transactionMap, addressField, tagField string, changes *[]addressChange) error {
	value, present := tx[addressField]
	if !present || value == nil {
		return nil
	}
	address, ok := TransactionString(value)
	if !ok {
		return fmt.Errorf("%w: %s", ErrAddressFieldIsNotAString, addressField)
	}
	if addresscodec.IsValidClassicAddress(address) {
		return nil
	}
	if !addresscodec.IsValidXAddress(address) {
		return fmt.Errorf("%w: %s", ErrInvalidAddress, addressField)
	}

	classic, tag, hasTag, _, err := addresscodec.XAddressToClassicAddress(address)
	if err != nil {
		return fmt.Errorf("decode %s X-address: %w", addressField, err)
	}
	if hasTag && tagField == "" {
		return fmt.Errorf("%w: %s", ErrAccountIDTagNotAllowed, addressField)
	}
	change := addressChange{
		tx:           tx,
		addressField: addressField,
		classic:      classic,
		tagField:     tagField,
		tag:          tag,
		hasTag:       hasTag,
	}
	if change.hasTag {
		explicit, explicitPresent := tx[tagField]
		if explicitPresent && explicit != nil {
			explicitTag, ok := explicit.(uint32)
			if !ok {
				return fmt.Errorf("%w: %s", ErrTagFieldIsNotAUint32, tagField)
			}
			if explicitTag != tag {
				return fmt.Errorf("%w: %q must equal the tag embedded in %q", ErrMismatchedTag, tagField, addressField)
			}
		}
	}
	*changes = append(*changes, change)
	return nil
}

// TransactionString returns value as a string when it is a string or a named string type.
func TransactionString(value any) (string, bool) {
	if address, ok := value.(string); ok {
		return address, true
	}
	// Use reflection to accept named string types for compatibility with typed transaction values.
	reflected := reflect.ValueOf(value)
	if reflected.IsValid() && reflected.Kind() == reflect.String {
		return reflected.String(), true
	}
	return "", false
}

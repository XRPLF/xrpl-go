package client

import (
	"fmt"
	"reflect"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

type addressChange struct {
	tx           map[string]any
	addressField string
	classic      string
	tagField     string
	tag          uint32
	hasTag       bool
}

// SetValidAddresses converts every in-scope X-address in an outer transaction
// and its Batch inner transactions to a classic address. Embedded Account and
// Destination tags are applied only after all conflicts have been validated.
func SetValidAddresses(tx map[string]any) error {
	// Reserve space for four common-case changes while allowing the slice to grow for larger transactions.
	changes := make([]addressChange, 0, 4)
	if err := collectTransactionAddressChanges(tx, &changes); err != nil {
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

func collectTransactionAddressChanges(tx map[string]any, changes *[]addressChange) error {
	if err := collectAddressChange(tx, "Account", "SourceTag", changes); err != nil {
		return err
	}
	if err := collectAddressChange(tx, "Destination", "DestinationTag", changes); err != nil {
		return err
	}

	for _, field := range [...]string{"Authorize", "Unauthorize", "Owner", "RegularKey"} {
		if err := collectAddressChange(tx, field, "", changes); err != nil {
			return err
		}
	}

	inners, err := batchInnerTransactions(tx)
	if err != nil {
		return err
	}
	for _, inner := range inners {
		if err := collectTransactionAddressChanges(inner, changes); err != nil {
			return err
		}
	}
	return nil
}

func collectAddressChange(tx map[string]any, addressField, tagField string, changes *[]addressChange) error {
	value, present := tx[addressField]
	if !present || value == nil {
		return nil
	}
	address, ok := transactionString(value)
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
	change := addressChange{
		tx:           tx,
		addressField: addressField,
		classic:      classic,
		tagField:     tagField,
		tag:          tag,
		hasTag:       hasTag && tagField != "",
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

func transactionString(value any) (string, bool) {
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

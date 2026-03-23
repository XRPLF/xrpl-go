package ledger

import (
	"errors"
	"fmt"
)

var (
	// ledger object

	// ErrUnsupportedLedgerObjectType is returned when an unsupported ledger object type is encountered.
	ErrUnsupportedLedgerObjectType = errors.New("unsupported ledger object type")

	// oracle

	// ErrPriceDataAssetPriceAndScale is returned when the asset price and scale are not set together.
	ErrPriceDataAssetPriceAndScale = errors.New("asset price and scale must be set together")
	// ErrPriceDataBaseAsset is returned when the base asset is required but not set.
	ErrPriceDataBaseAsset = errors.New("base asset is required")
	// ErrPriceDataQuoteAsset is returned when the quote asset is required but not set.
	ErrPriceDataQuoteAsset = errors.New("quote asset is required")
)

// ErrPriceDataScale is returned when the scale is greater than the maximum allowed.
type ErrPriceDataScale struct {
	Value uint8
	Limit uint8
}

// Error implements the error interface for ErrPriceDataScale
func (e ErrPriceDataScale) Error() string {
	return fmt.Sprintf("invalid price data scale: got %d, must be less than %d", e.Value, e.Limit)
}

// ErrUnrecognizedLedgerObjectType is returned when an unrecognized ledger object type is encountered.
type ErrUnrecognizedLedgerObjectType struct {
	Type any
}

// Error implements the error interface for ErrUnrecognizedLedgerObjectType
func (e ErrUnrecognizedLedgerObjectType) Error() string {
	return fmt.Sprintf("unrecognized Ledger Object type: %v", e.Type)
}

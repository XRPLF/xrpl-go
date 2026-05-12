package currency

import "errors"

var (
	// ErrXrpToDropsInvalidValue indicates an invalid XRP amount.
	ErrXrpToDropsInvalidValue = errors.New("xrp to drops: invalid value")
	// ErrXrpToDropsNegativeValue indicates a negative XRP amount.
	ErrXrpToDropsNegativeValue = errors.New("xrp to drops: value cannot be negative")
	// ErrXrpToDropsTooManyDecimals indicates an XRP amount with more than six decimal places.
	ErrXrpToDropsTooManyDecimals = errors.New("xrp to drops: value has too many decimals")
	// ErrXrpToDropsExceedsMax indicates an XRP amount above the maximum native supply.
	ErrXrpToDropsExceedsMax = errors.New("xrp to drops: value exceeds maximum XRP supply")

	// ErrDropsToXrpInvalidValue indicates an invalid drops amount.
	ErrDropsToXrpInvalidValue = errors.New("drops to xrp: invalid value")
	// ErrDropsToXrpNegativeValue indicates a negative drops amount.
	ErrDropsToXrpNegativeValue = errors.New("drops to xrp: value cannot be negative")
	// ErrDropsToXrpFractionalDrops indicates a drops amount with a fractional value.
	ErrDropsToXrpFractionalDrops = errors.New("drops to xrp: value cannot contain fractional drops")
	// ErrDropsToXrpExceedsMax indicates a drops amount above the maximum native supply.
	ErrDropsToXrpExceedsMax = errors.New("drops to xrp: value exceeds maximum XRP supply")
)

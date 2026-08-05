// Package client contains transport-independent client autofill policy.
package client

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// RestrictedNetworks is the largest network ID for which transactions must
	// omit NetworkID. Deliberately untyped so the client re-exports keep the
	// untyped-constant shape of the original public constants.
	RestrictedNetworks = 1024
	// RequiredNetworkIDVersion is the first rippled version that enforces the
	// NetworkID transaction field.
	RequiredNetworkIDVersion = "1.11.0"
)

// NetworkIdentity is the server identity needed to apply transaction NetworkID
// policy. A nil NetworkID means the server identity is unknown. A non-nil zero
// value is the explicitly discovered mainnet ID.
type NetworkIdentity struct {
	NetworkID    *uint32
	BuildVersion string
}

// CloneNetworkID returns a copy of networkID so callers never alias the
// original pointer. Nil stays nil.
func CloneNetworkID(networkID *uint32) *uint32 {
	if networkID == nil {
		return nil
	}
	value := *networkID
	return &value
}

// ResolveNetworkIdentity validates a server_info identity against an optional
// trusted override. A matching override pointer is preserved instead of being
// replaced by the discovered pointer.
func ResolveNetworkIdentity(override *uint32, discovered NetworkIdentity) (NetworkIdentity, error) {
	if discovered.NetworkID == nil {
		return NetworkIdentity{}, ErrNetworkIDUnavailable
	}
	if override != nil && *override != *discovered.NetworkID {
		return NetworkIdentity{}, fmt.Errorf(
			"%w: configured %d, discovered %d",
			ErrNetworkIDOverrideMismatch,
			*override,
			*discovered.NetworkID,
		)
	}

	resolved := discovered
	if override != nil {
		resolved.NetworkID = override
	}
	return ValidateNetworkIdentity(resolved)
}

// ValidateNetworkIdentity returns identity unchanged when it is complete
// enough to apply NetworkID policy, and the zero identity with the policy
// error otherwise.
func ValidateNetworkIdentity(identity NetworkIdentity) (NetworkIdentity, error) {
	if _, err := NetworkIDRequired(identity); err != nil {
		return NetworkIdentity{}, err
	}
	return identity, nil
}

// ApplyNetworkIDPolicy validates and applies NetworkID to an outer transaction
// and, for Batch transactions, every inner transaction using the same policy.
// Validation completes for every target before any map is mutated.
func ApplyNetworkIDPolicy(tx map[string]any, identity NetworkIdentity) error {
	required, err := NetworkIDRequired(identity)
	if err != nil {
		return err
	}

	inners, err := batchInnerTransactions(tx)
	if err != nil {
		return err
	}
	targets := append([]map[string]any{tx}, inners...)

	for _, target := range targets {
		if err := validateNetworkID(target, identity, required); err != nil {
			return err
		}
	}
	for _, target := range targets {
		if value, present := target["NetworkID"]; present && value != nil {
			continue
		}
		delete(target, "NetworkID")
		if required {
			target["NetworkID"] = *identity.NetworkID
		}
	}
	return nil
}

// NetworkIDRequired reports whether transactions for identity must include a
// NetworkID. Networks 0 through 1024 always omit it. Restricted networks add it
// only when the server is rippled 1.11.0 or newer.
func NetworkIDRequired(identity NetworkIdentity) (bool, error) {
	if identity.NetworkID == nil {
		return false, ErrNetworkIDUnavailable
	}
	if *identity.NetworkID <= RestrictedNetworks {
		return false, nil
	}
	if identity.BuildVersion == "" {
		return false, ErrBuildVersionUnavailable
	}

	comparison, err := compareRippledVersions(identity.BuildVersion, RequiredNetworkIDVersion)
	if err != nil {
		return false, fmt.Errorf("%w %q: %w", ErrInvalidBuildVersion, identity.BuildVersion, err)
	}
	return comparison >= 0, nil
}

// batchInnerTransactions returns the inner transaction objects of a Batch
// transaction, or nil when tx is not a Batch transaction.
func batchInnerTransactions(tx map[string]any) ([]map[string]any, error) {
	if txType, _ := tx["TransactionType"].(string); txType != "Batch" {
		return nil, nil
	}
	rawTransactions, ok := tx["RawTransactions"].([]map[string]any)
	if !ok {
		return nil, ErrRawTransactionsFieldIsNotAnArray
	}
	inners := make([]map[string]any, 0, len(rawTransactions))
	for _, wrapper := range rawTransactions {
		inner, ok := wrapper["RawTransaction"].(map[string]any)
		if !ok {
			return nil, ErrRawTransactionFieldIsNotAnObject
		}
		inners = append(inners, inner)
	}
	return inners, nil
}

func validateNetworkID(tx map[string]any, identity NetworkIdentity, required bool) error {
	value, present := tx["NetworkID"]
	if !present || value == nil {
		return nil
	}

	networkID, ok := value.(uint32)
	if !ok {
		return ErrNetworkIDFieldIsNotAUint32
	}
	if networkID != *identity.NetworkID {
		return ErrNetworkIDFieldMismatch
	}
	if !required {
		return ErrNetworkIDFieldUnexpected
	}
	return nil
}

type rippledVersion struct {
	core       [3]uint64 // major, minor, patch
	prerelease string
}

func compareRippledVersions(left, right string) (int, error) {
	leftVersion, err := parseRippledVersion(left)
	if err != nil {
		return 0, err
	}
	rightVersion, err := parseRippledVersion(right)
	if err != nil {
		return 0, err
	}

	for i := range leftVersion.core {
		switch {
		case leftVersion.core[i] < rightVersion.core[i]:
			return -1, nil
		case leftVersion.core[i] > rightVersion.core[i]:
			return 1, nil
		}
	}

	switch {
	case leftVersion.prerelease == rightVersion.prerelease:
		return 0, nil
	case leftVersion.prerelease == "":
		return 1, nil
	case rightVersion.prerelease == "":
		return -1, nil
	default:
		return strings.Compare(leftVersion.prerelease, rightVersion.prerelease), nil
	}
}

func parseRippledVersion(version string) (rippledVersion, error) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	if buildIndex := strings.IndexByte(version, '+'); buildIndex >= 0 {
		version = version[:buildIndex]
	}

	core, prerelease, _ := strings.Cut(version, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return rippledVersion{}, errInvalidRippledVersionFormat
	}

	parsed := rippledVersion{prerelease: prerelease}
	for i, part := range parts {
		value, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return rippledVersion{}, fmt.Errorf("component %q is not an unsigned integer", part)
		}
		parsed.core[i] = value
	}
	return parsed, nil
}

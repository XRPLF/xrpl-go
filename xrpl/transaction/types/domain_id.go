//revive:disable:var-naming
package types

// DomainID returns the domain ID as a string.
func DomainID(id string) *string {
	return new(id)
}

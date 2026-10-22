package server

import "errors"

var (
	// ErrInvalidDefinitionsHash is returned when a server_definitions hash is not a 256-bit hexadecimal string.
	ErrInvalidDefinitionsHash = errors.New("server_definitions: hash must be a 256-bit hexadecimal string")
	// ErrInvalidDefinitionsResponse is returned when a server_definitions response is neither hash-only nor a complete legacy definitions payload.
	ErrInvalidDefinitionsResponse = errors.New("server_definitions: response must be hash-only or contain all core definition sections")
	// ErrInvalidDefinitionField is returned when a FIELDS entry is not a valid [name, properties] tuple.
	ErrInvalidDefinitionField = errors.New("server_definitions: invalid FIELDS entry")
	// ErrInvalidDefinitionFormatField is returned when a TRANSACTION_FORMATS or LEDGER_ENTRY_FORMATS field has an empty name or an out-of-range optionality.
	ErrInvalidDefinitionFormatField = errors.New("server_definitions: invalid format field")
)

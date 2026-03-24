package rpc

// Request represents a JSON-RPC request with method and parameters.
type Request struct {
	Method string `json:"method"`
	Params [1]any `json:"params,omitempty"`
}

// APIVersionRequest defines the interface for requests that support API versioning.
type APIVersionRequest interface {
	APIVersion() int
	SetAPIVersion(apiVersion int)
}

// XRPLRequest defines the interface for XRPL-specific requests that include API versioning, method, and validation.
type XRPLRequest interface {
	APIVersionRequest
	Method() string
	Validate() error
}

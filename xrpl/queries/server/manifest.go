package server

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ManifestDetails represents validator manifest information, including the domain,
// ephemeral public key, master public key, and sequence number of the manifest.
type ManifestDetails struct {
	Domain       string `json:"domain"`
	EphemeralKey string `json:"ephemeral_key"`
	MasterKey    string `json:"master_key"`
	Seq          uint   `json:"seq"`
}

// ############################################################################
// Request
// ############################################################################

// ManifestRequest is the request type for the manifest command.
// It reports the current manifest information for a given validator public key.
type ManifestRequest struct {
	common.BaseRequest
	PublicKey string `json:"public_key"`
}

// Method returns the JSON-RPC method name for the ManifestRequest.
func (*ManifestRequest) Method() string {
	return "manifest"
}

// APIVersion returns the API version required by the ManifestRequest.
func (*ManifestRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate verifies the ManifestRequest parameters.
// TODO: implement V2.
func (*ManifestRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// ManifestResponse is the response type returned by the manifest command.
// It includes the parsed manifest details, the raw manifest string, and the requested key.
type ManifestResponse struct {
	Details   ManifestDetails `json:"details,omitzero"`
	Manifest  string          `json:"manifest,omitempty"`
	Requested string          `json:"requested"`
}

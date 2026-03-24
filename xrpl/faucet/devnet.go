package faucet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// DevnetFaucetHost is the hostname for the XRPL Devnet faucet service.
	DevnetFaucetHost = "faucet.devnet.rippletest.net"
	// DevnetFaucetPath is the API path for account operations on the Devnet faucet.
	DevnetFaucetPath = "/accounts"
)

// DevnetFaucetProvider implements the FaucetProvider interface for the XRPL Devnet.
// It provides functionality to interact with the Devnet faucet for funding wallets.
type DevnetFaucetProvider struct {
	host        string // The hostname of the Devnet faucet
	accountPath string // The API path for account-related operations
}

// NewDevnetFaucetProvider creates and returns a new instance of DevnetFaucetProvider
// with predefined Devnet faucet host and account path.
func NewDevnetFaucetProvider() *DevnetFaucetProvider {
	return &DevnetFaucetProvider{
		host:        DevnetFaucetHost,
		accountPath: DevnetFaucetPath,
	}
}

// FundWallet sends a request to the Devnet faucet to fund the specified wallet address.
// It returns an error if the funding request fails.
func (fp *DevnetFaucetProvider) FundWallet(address types.Address) error {
	url := fmt.Sprintf("https://%s%s", fp.host, fp.accountPath)
	payload := map[string]string{"destination": address.String(), "userAgent": UserAgent}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return ErrMarshalPayload{
			Err: err,
		}
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return ErrCreateRequest{Err: err}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrSendRequest{Err: err}
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return ErrUnexpectedStatusCode{
			Code: resp.StatusCode,
		}
	}

	return nil
}

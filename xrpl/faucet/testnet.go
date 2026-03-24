package faucet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TestnetFaucetHost is the hostname for the XRPL Testnet faucet service.
	TestnetFaucetHost = "faucet.altnet.rippletest.net"
	// TestnetFaucetPath is the API path for account operations on the Testnet faucet.
	TestnetFaucetPath = "/accounts"
)

// TestnetFaucetProvider implements the FaucetProvider interface for the XRPL Testnet.
// It provides functionality to interact with the Testnet faucet for funding wallets.
type TestnetFaucetProvider struct {
	host        string
	accountPath string
}

// NewTestnetFaucetProvider creates and returns a new instance of TestnetFaucetProvider
// with predefined Testnet faucet host and account path.
func NewTestnetFaucetProvider() *TestnetFaucetProvider {
	return &TestnetFaucetProvider{
		host:        TestnetFaucetHost,
		accountPath: TestnetFaucetPath,
	}
}

// FundWallet sends a request to the Testnet faucet to fund the specified wallet address.
// It returns an error if the funding request fails.
func (fp *TestnetFaucetProvider) FundWallet(address types.Address) error {
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
		return ErrCreateRequest{
			Err: err,
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrSendRequest{
			Err: err,
		}
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

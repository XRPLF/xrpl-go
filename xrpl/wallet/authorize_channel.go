package wallet

import (
	"encoding/hex"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/keypairs"
)

// AuthorizeChannel returns a signature authorizing the redemption of a specific
// amount of XRP from a payment channel.
//
// channelID identifies the payment channel.
// amount is the amount to redeem, expressed in drops.
//
// Returns the signature or an error if the signature cannot be created.
func AuthorizeChannel(channelID, amount string, wallet Wallet) (string, error) {
	encodedData, err := binarycodec.EncodeForSigningClaim(map[string]any{
		"Channel": channelID,
		"Amount":  amount,
	})
	if err != nil {
		return "", err
	}
	hexData, err := hex.DecodeString(encodedData)
	if err != nil {
		return "", err
	}
	signedData, err := keypairs.Sign(string(hexData), wallet.PrivateKey)
	if err != nil {
		return "", err
	}
	return signedData, nil
}

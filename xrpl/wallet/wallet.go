// Package wallet provides utilities for deriving and managing XRPL wallets,
// including keypair generation, address derivation, and offline transaction signing.
package wallet

import (
	"encoding/hex"
	"fmt"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/keypairs"
	"github.com/Peersyst/xrpl-go/pkg/random"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/interfaces"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	bip32 "github.com/bsv-blockchain/go-sdk/compat/bip32"
	"github.com/bsv-blockchain/go-sdk/compat/bip39"
	chaincfg "github.com/bsv-blockchain/go-sdk/transaction/chaincfg"
)

var (
	nilHDPrivateKeyID = [4]byte{0x00, 0x00, 0x00, 0x00}
)

// Wallet is a utility for deriving a wallet composed of a keypair (publicKey/privateKey).
// It can be derived from a seed, mnemonic, or entropy, and supports offline signing and verification.
type Wallet struct {
	PublicKey      string
	PrivateKey     string
	ClassicAddress types.Address
	Seed           string
}

// New creates a new random Wallet. In order to make this a valid account on ledger, you must send XRP to it.
func New(alg interfaces.CryptoImplementation) (Wallet, error) {
	seed, err := keypairs.GenerateSeed("", alg, random.NewRandomizer())
	if err != nil {
		return Wallet{}, err
	}
	return FromSeed(seed, "")
}

// FromSeed derives a Wallet from a seed.
func FromSeed(seed string, masterAddress string) (Wallet, error) {
	privKey, pubKey, err := keypairs.DeriveKeypair(seed, false)
	if err != nil {
		return Wallet{}, err
	}

	var classicAddr types.Address

	if masterAddress != "" {
		classicAddr, err = ensureClassicAddress(masterAddress)
		if err != nil {
			return Wallet{}, err
		}
	} else {
		addr, err := keypairs.DeriveClassicAddress(pubKey)
		if err != nil {
			return Wallet{}, err
		}
		classicAddr = types.Address(addr)
	}

	return Wallet{
		PublicKey:      pubKey,
		PrivateKey:     privKey,
		Seed:           seed,
		ClassicAddress: classicAddr,
	}, nil

}

// FromSecret derives a Wallet from a secret (AKA a seed).
func FromSecret(seed string) (Wallet, error) {
	return FromSeed(seed, "")
}

// FromMnemonic derives a Wallet from a bip39 or RFC1751 mnemonic (defaults to bip39).
func FromMnemonic(mnemonic string) (*Wallet, error) {
	// Validate the mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, bip39.ErrInvalidMnemonic
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Derive the master key

	params := &chaincfg.Params{
		HDPrivateKeyID: nilHDPrivateKeyID,
	}
	masterKey, err := bip32.NewMaster(seed, params)
	if err != nil {
		return nil, err
	}

	// Derive the key using the path m/44'/144'/0'/0/0
	path := []uint32{
		44 + bip32.HardenedKeyStart,
		144 + bip32.HardenedKeyStart,
		bip32.HardenedKeyStart,
		0,
		0,
	}

	key := masterKey
	for _, childNum := range path {
		key, err = key.Child(childNum)
		if err != nil {
			return nil, err
		}
	}

	// Convert the private key to the format expected by the XRPL library
	ecPriv, err := key.ECPrivKey()
	if err != nil {
		return nil, err
	}

	privKey := strings.ToUpper(ecPriv.Hex())
	pubKey := strings.ToUpper(hex.EncodeToString(ecPriv.PubKey().Compressed()))

	// Derive classic address
	classicAddr, err := keypairs.DeriveClassicAddress(pubKey)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PublicKey:      pubKey,
		PrivateKey:     fmt.Sprintf("00%s", privKey),
		ClassicAddress: types.Address(classicAddr),
		Seed:           "", // We don't have the seed in this case
	}, nil
}

// Sign signs a transaction offline, returning the transaction blob and its signature.
// TODO: Refactor to accept a `Transaction` object instead of a map.
func (w *Wallet) Sign(tx map[string]interface{}) (string, string, error) {
	tx["SigningPubKey"] = w.PublicKey

	// Copy the transaction to avoid modifying the original transaction
	signTx := make(map[string]interface{}, len(tx))
	for k, v := range tx {
		signTx[k] = v
	}

	encodedTx, err := binarycodec.EncodeForSigning(signTx)
	if err != nil {
		return "", "", err
	}

	txHash, err := w.computeSignature(encodedTx)
	if err != nil {
		return "", "", err
	}

	tx["TxnSignature"] = txHash

	txBlob, err := binarycodec.Encode(tx)
	if err != nil {
		return "", "", err
	}

	txHash, err = hash.SignTxBlob(txBlob)
	if err != nil {
		return "", "", err
	}

	return txBlob, txHash, nil
}

// GetAddress returns the classic address of the wallet.
func (w *Wallet) GetAddress() types.Address {
	return types.Address(w.ClassicAddress)
}

// Multisign signs a multisigned transaction offline, returning the signed transaction blob and its transaction hash.
func (w *Wallet) Multisign(tx map[string]interface{}) (string, string, error) {
	encodedTx, err := binarycodec.EncodeForMultisigning(tx, w.ClassicAddress.String())
	if err != nil {
		return "", "", err
	}

	txHash, err := w.computeSignature(encodedTx)
	if err != nil {
		return "", "", err
	}

	signer := types.Signer{
		SignerData: types.SignerData{
			Account:       w.ClassicAddress,
			TxnSignature:  txHash,
			SigningPubKey: w.PublicKey,
		},
	}

	tx["Signers"] = []any{signer.Flatten()}
	blob, err := binarycodec.Encode(tx)
	if err != nil {
		return "", "", err
	}
	blobHash, err := hash.SignTxBlob(blob)
	if err != nil {
		return "", "", err
	}

	return blob, blobHash, nil
}

// Computes the signature of a transaction.
// Returns the signature of the transaction. If an error occurs, it will return an error.
func (w *Wallet) computeSignature(encodedTx string) (string, error) {
	hexTx, err := hex.DecodeString(encodedTx)
	if err != nil {
		return "", err
	}

	txHash, err := keypairs.Sign(string(hexTx), w.PrivateKey)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

// ComputeSignature is the public wrapper for computeSignature, exposed for
// external dual-signing use cases (e.g., XLS-66 LoanSet transactions).
func (w *Wallet) ComputeSignature(encodedTx string) (string, error) {
	return w.computeSignature(encodedTx)
}

// Ensures that the address is a classic address.
// If the address is an x-address with a tag of 0 (no tag), it will be converted to a classic address.
// If the address is not a classic address, it will be returned as is.
func ensureClassicAddress(account string) (types.Address, error) {
	if ok := addresscodec.IsValidXAddress(account); ok {
		classicAddr, tag, _, err := addresscodec.XAddressToClassicAddress(account)
		if err != nil {
			return "", err
		}

		if tag != 0 {
			return "", ErrAddressTagNotZero
		}

		return types.Address(classicAddr), nil
	}

	return types.Address(account), nil
}

// Verifies a signed transaction offline.
// Returns a boolean indicating if the transaction is valid and an error if it is not.
// If the transaction is signed with a public key, the public key must match the one in the transaction.
// func (w *Wallet) VerifyTransaction(tx map[string]any) (bool, error) {
// 	return false, errors.New("not implemented")
// }

// // Gets an X-address in Testnet/Mainnet format.
// func (w *Wallet) GetXAddress() (string, error) {
// 	return "", errors.New("not implemented")
// }

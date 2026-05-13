package keypairs

import (
	"errors"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/keypairs/interfaces"
)

const (
	// verificationMessage is the message that is used to verify the signature of the derived keypair.
	// Only used for testing purposes.
	verificationMessage = "This test message should verify."
)

// GenerateSeed generates a seed from raw entropy, a crypto algorithm implementation and a randomizer.
// If entropy is nil or empty, it generates a random seed using r, r must be non-nil in that case
// (otherwise ErrRandomizerRequired is returned). When entropy is provided it must be exactly
// addresscodec.FamilySeedLength bytes and r is not consulted, so callers may pass nil for r.
// Do not pass passphrases directly. For deterministic passphrase-based seeds, hash or KDF the
// passphrase outside this function and pass exactly 16 derived bytes.
// The seed is encoded using the addresscodec package.
func GenerateSeed(entropy []byte, alg interfaces.KeypairCryptoAlg, r interfaces.Randomizer) (string, error) {
	if len(entropy) == 0 {
		if r == nil {
			return "", ErrRandomizerRequired
		}
		pe, err := r.GenerateBytes(addresscodec.FamilySeedLength)
		if err != nil {
			return "", err
		}
		return addresscodec.EncodeSeed(pe, alg)
	}
	// EncodeSeed validates that caller-supplied entropy is exactly FamilySeedLength bytes.
	encoded, err := addresscodec.EncodeSeed(entropy, alg)
	if err != nil {
		var lengthErr *addresscodec.EncodeLengthError
		if errors.As(err, &lengthErr) {
			return "", fmt.Errorf("%w: %w", ErrInvalidEntropyLength, err)
		}
		return "", err
	}
	return encoded, nil
}

// DeriveKeypair derives a key pair from a given seed. Returns a tuple of private key and public key.
// The seed has to be encoded using the addresscodec package. Otherwise, it returns an error.
func DeriveKeypair(seed string, validator bool) (private, public string, err error) {
	ds, alg, err := addresscodec.DecodeSeed(seed)
	if err != nil {
		return "", "", err
	}
	private, public, err = alg.DeriveKeypair(ds, validator)
	if err != nil {
		return "", "", err
	}
	signature, err := alg.Sign(verificationMessage, private)
	if err != nil {
		return "", "", err
	}
	if !alg.Validate(verificationMessage, public, signature) {
		return "", "", ErrInvalidSignature
	}
	return private, public, nil
}

// DeriveClassicAddress derives a classic address from a given public key.
// The public key has to be encoded using the addresscodec package. Otherwise, it returns an error.
func DeriveClassicAddress(pubKey string) (string, error) {
	return addresscodec.EncodeClassicAddressFromPublicKeyHex(pubKey)
}

// DeriveNodeAddress derives a node address from a given public key.
// The public key has to be encoded using the addresscodec package. Otherwise, it returns an error.
func DeriveNodeAddress(pubKey string, alg interfaces.NodeDerivationCryptoAlg) (string, error) {
	decoded, err := addresscodec.DecodeNodePublicKey(pubKey)
	if err != nil {
		return "", err
	}
	accountPubKey, err := alg.DerivePublicKeyFromPublicGenerator(decoded)
	if err != nil {
		return "", err
	}

	accountID := addresscodec.Sha256RipeMD160(accountPubKey)

	return addresscodec.EncodeAccountIDToClassicAddress(accountID)
}

// Sign signs a message with a given private key.
// The private key needs to satisfy a crypto algorithm implementation. Otherwise, it returns an error.
// Currently, only ED25519 and SECP256K1 are supported.
// If the message is empty, it returns an error.
func Sign(msg, privKey string) (string, error) {
	alg := getCryptoImplementationFromKey(privKey)
	if alg == nil {
		return "", ErrInvalidCryptoImplementation
	}
	return alg.Sign(msg, privKey)
}

// Validate validates a signature of a message with a given public key.
// The public key needs to satisfy a crypto algorithm implementation. Otherwise, it returns an error.
// Currently, only ED25519 and SECP256K1 are supported.
// If the message is empty, it returns an error.
func Validate(msg, pubKey, sig string) (bool, error) {
	alg := getCryptoImplementationFromKey(pubKey)
	if alg == nil {
		return false, ErrInvalidCryptoImplementation
	}
	return alg.Validate(msg, pubKey, sig), nil
}

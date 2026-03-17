package crypto

import (
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"

	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

const (
	// SECP256K1 prefix - value is 0
	secp256K1Prefix byte = 0x00
)

var (
	_ Algorithm = SECP256K1CryptoAlgorithm{}
	// SECP256K1 family seed prefix - value is 33
	secp256K1FamilySeedPrefix = []byte{0x21}
)

// SECP256K1CryptoAlgorithm is the implementation of the SECP256K1 algorithm.
type SECP256K1CryptoAlgorithm struct {
	prefix byte
}

// SECP256K1 returns a new SECP256K1CryptoAlgorithm instance.
func SECP256K1() SECP256K1CryptoAlgorithm {
	return SECP256K1CryptoAlgorithm{
		prefix: secp256K1Prefix,
	}
}

// Prefix returns the prefix for the SECP256K1 algorithm.
func (c SECP256K1CryptoAlgorithm) Prefix() byte {
	return c.prefix
}

// FamilySeedPrefix returns the family seed prefix for the SECP256K1 algorithm.
func (c SECP256K1CryptoAlgorithm) FamilySeedPrefix() []byte {
	return secp256K1FamilySeedPrefix
}

// deriveScalar derives a scalar from a seed.
func (c SECP256K1CryptoAlgorithm) deriveScalar(bytes []byte, discrim *uint32) secp256k1.ModNScalar {
	for i := uint32(0); ; i++ {
		hash := sha512.New()
		hash.Write(bytes)

		if discrim != nil {
			var discrimBytes [4]byte
			binary.BigEndian.PutUint32(discrimBytes[:], *discrim)
			hash.Write(discrimBytes[:])
		}

		var shiftBytes [4]byte
		binary.BigEndian.PutUint32(shiftBytes[:], i)
		hash.Write(shiftBytes[:])

		// Convert hash slice to fixed 32 byte
		var hashBytes [32]byte
		copy(hashBytes[:], hash.Sum(nil)[:32])

		var scalar secp256k1.ModNScalar
		// overflow is non-zero if the hash value exceeds the secp256k1 curve order.
		// A valid scalar must be in range (0, order), so we retry on overflow or zero.
		overflow := scalar.SetBytes(&hashBytes)
		if overflow == 0 && !scalar.IsZero() {
			return scalar
		}

		if i == 0xffffffff {
			break
		}
	}
	// This error is practically impossible to reach.
	// The order of the curve describes the (finite) amount of points on the curve.
	panic("impossible unicorn ;)")
}

// DeriveKeypair derives a keypair from a seed.
func (c SECP256K1CryptoAlgorithm) DeriveKeypair(seed []byte, validator bool) (string, string, error) {
	if validator {
		return "", "", ErrValidatorKeypairDerivation
	}

	privateGen := c.deriveScalar(seed, nil)

	rootPubKey := secp256k1.NewPrivateKey(&privateGen).PubKey().SerializeCompressed()
	discrim := uint32(0)
	derivedScalar := c.deriveScalar(rootPubKey, &discrim)

	var finalScalar secp256k1.ModNScalar
	finalScalar.Add2(&derivedScalar, &privateGen)
	if finalScalar.IsZero() {
		return "", "", ErrDerivedKeyIsZero
	}

	finalPrivKey := secp256k1.NewPrivateKey(&finalScalar)
	privKeyHex := "00" + hexutil.EncodeToUpperHex(finalPrivKey.Serialize())
	pubKeyHex := hexutil.EncodeToUpperHex(finalPrivKey.PubKey().SerializeCompressed())

	return privKeyHex, pubKeyHex, nil
}

// Sign signs a message with a private key.
func (c SECP256K1CryptoAlgorithm) Sign(msg, privKey string) (string, error) {
	if len(privKey) != 64 && len(privKey) != 66 {
		return "", ErrInvalidPrivateKey
	}
	if len(msg) == 0 {
		return "", ErrInvalidMessage
	}

	if len(privKey) == 66 {
		privKey = privKey[2:]
	}
	key, err := hex.DecodeString(privKey)
	if err != nil {
		return "", ErrInvalidPrivateKey
	}

	secpPrivKey := secp256k1.PrivKeyFromBytes(key)
	sig := ecdsa.Sign(secpPrivKey, Sha512Half([]byte(msg)))

	return hexutil.EncodeToUpperHex(sig.Serialize()), nil
}

// Validate validates a signature for a message with a public key.
func (c SECP256K1CryptoAlgorithm) Validate(msg, pubkey, sig string) bool {
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	parsedSig, err := ecdsa.ParseDERSignature(sigBytes)
	if err != nil {
		return false
	}

	hash := Sha512Half([]byte(msg))

	pubkeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return false
	}

	pubKey, err := secp256k1.ParsePubKey(pubkeyBytes)
	if err != nil {
		return false
	}

	return parsedSig.Verify(hash, pubKey)
}

// DerivePublicKeyFromPublicGenerator derives a public key from a public generator.
func (c SECP256K1CryptoAlgorithm) DerivePublicKeyFromPublicGenerator(pubKey []byte) ([]byte, error) {
	rootPubKey, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return nil, err
	}

	discrim := uint32(0)
	scalar := c.deriveScalar(pubKey, &discrim)

	var scalarPoint secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(&scalar, &scalarPoint)

	var rootPoint secp256k1.JacobianPoint
	rootPubKey.AsJacobian(&rootPoint)

	var result secp256k1.JacobianPoint
	secp256k1.AddNonConst(&rootPoint, &scalarPoint, &result)
	result.ToAffine()

	finalPubKey := secp256k1.NewPublicKey(&result.X, &result.Y)
	return finalPubKey.SerializeCompressed(), nil
}

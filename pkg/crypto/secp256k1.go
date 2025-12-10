package crypto

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

const (
	// SECP256K1 prefix - value is 0
	secp256K1Prefix byte = 0x00
	// SECP256K1 family seed prefix - value is 33
	secp256K1FamilySeedPrefix byte = 0x21
)

var (
	_ Algorithm = SECP256K1CryptoAlgorithm{}

	// ErrValidatorKeypairDerivation is returned when a validator keypair is attempted to be derived
	ErrValidatorKeypairDerivation = errors.New("validator keypair derivation not supported")
	// ErrInvalidPrivateKey is returned when a private key is invalid
	ErrInvalidPrivateKey = errors.New("invalid private key")
	// ErrInvalidMessage is returned when a message is required but not provided
	ErrInvalidMessage = errors.New("message is required")
)

// SECP256K1CryptoAlgorithm is the implementation of the SECP256K1 algorithm.
type SECP256K1CryptoAlgorithm struct {
	prefix           byte
	familySeedPrefix byte
}

// SECP256K1 returns a new SECP256K1CryptoAlgorithm instance.
func SECP256K1() SECP256K1CryptoAlgorithm {
	return SECP256K1CryptoAlgorithm{
		prefix:           secp256K1Prefix,
		familySeedPrefix: secp256K1FamilySeedPrefix,
	}
}

// Prefix returns the prefix for the SECP256K1 algorithm.
func (c SECP256K1CryptoAlgorithm) Prefix() byte {
	return c.prefix
}

// FamilySeedPrefix returns the family seed prefix for the SECP256K1 algorithm.
func (c SECP256K1CryptoAlgorithm) FamilySeedPrefix() byte {
	return c.familySeedPrefix
}

// deriveScalar derives a scalar from a seed.
func (c SECP256K1CryptoAlgorithm) deriveScalar(bytes []byte, discrim *big.Int) *big.Int {

	order := btcec.S256().N
	for i := int64(0); i <= int64(0xffffffff); i++ {
		hash := sha512.New()

		hash.Write(bytes)

		if discrim != nil {
			discrimBytes := make([]byte, 4)
			bytes[0] = byte(discrim.Uint64())
			bytes[1] = byte(discrim.Uint64() >> 8)
			bytes[2] = byte(discrim.Uint64() >> 16)
			bytes[3] = byte(discrim.Uint64() >> 24)

			hash.Write(discrimBytes)
		}

		shiftBytes := make([]byte, 4)
		bytes[0] = byte(i)
		bytes[1] = byte(i >> 8)
		bytes[2] = byte(i >> 16)
		bytes[3] = byte(i >> 24)

		hash.Write(shiftBytes)

		key := new(big.Int).SetBytes(hash.Sum(nil)[:32])

		if key.Cmp(big.NewInt(0)) > 0 && key.Cmp(order) < 0 {
			return key
		}
	}
	// This error is practically impossible to reach.
	// The order of the curve describes the (finite) amount of points on the curve.
	panic("impossible unicorn ;)")
}

// DeriveKeypair derives a keypair from a seed.
func (c SECP256K1CryptoAlgorithm) DeriveKeypair(seed []byte, validator bool) (string, string, error) {
	curve := btcec.S256()
	order := curve.N

	privateGen := c.deriveScalar(seed, nil)

	if validator {
		return "", "", ErrValidatorKeypairDerivation
	}

	rootPrivateKey, _ := btcec.PrivKeyFromBytes(privateGen.Bytes())

	derivatedScalar := c.deriveScalar(rootPrivateKey.PubKey().SerializeCompressed(), big.NewInt(0))
	scalarWithPrivateGen := derivatedScalar.Add(derivatedScalar, privateGen)
	privateKey := scalarWithPrivateGen.Mod(scalarWithPrivateGen, order)

	privKeyBytes := privateKey.Bytes()
	private := strings.ToUpper(hex.EncodeToString(privKeyBytes))

	_, pubKey := btcec.PrivKeyFromBytes(privKeyBytes)

	pubKeyBytes := pubKey.SerializeCompressed()

	return "00" + private, strings.ToUpper(hex.EncodeToString(pubKeyBytes)), nil
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

	parsedSig, err := DERHexFromSig(sig.R().String(), sig.S().String())
	if err != nil {
		return "", err
	}
	return strings.ToUpper(parsedSig), nil
}

// Validate validates a signature for a message with a public key.
func (c SECP256K1CryptoAlgorithm) Validate(msg, pubkey, sig string) bool {
	// Decode the signature from DERHex to a hex string
	r, s, err := DERHexToSig(sig)
	if err != nil {
		return false
	}

	// Convert r and s slices to [32]byte arrays
	var rBytes, sBytes [32]byte

	copy(rBytes[32-len(r):], r)
	copy(sBytes[32-len(s):], s)

	ecdsaR := &secp256k1.ModNScalar{}
	ecdsaS := &secp256k1.ModNScalar{}

	ecdsaR.SetBytes(&rBytes)
	ecdsaS.SetBytes(&sBytes)

	parsedSig := ecdsa.NewSignature(ecdsaR, ecdsaS)
	// Hash the message
	hash := Sha512Half([]byte(msg))

	// Decode the pubkey from hex to a byte slice
	pubkeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return false
	}

	// Verify the signature
	pubKey, err := secp256k1.ParsePubKey(pubkeyBytes)
	if err != nil {
		return false
	}
	return parsedSig.Verify(hash, pubKey)
}

// DerivePublicKeyFromPublicGenerator derives a public key from a public generator.
func (c SECP256K1CryptoAlgorithm) DerivePublicKeyFromPublicGenerator(pubKey []byte) ([]byte, error) {
	// Get the curve
	curve := btcec.S256()

	// Parse the input public key as a point
	rootPubKey, err := btcec.ParsePubKey(pubKey)
	if err != nil {
		return nil, err
	}

	// Derive scalar using existing function
	scalar := c.deriveScalar(pubKey, big.NewInt(0))

	// Multiply base point with scalar
	x, y := curve.ScalarBaseMult(scalar.Bytes())
	xField, yField := secp256k1.FieldVal{}, secp256k1.FieldVal{}

	xField.SetByteSlice(x.Bytes())
	yField.SetByteSlice(y.Bytes())

	scalarPoint := secp256k1.NewPublicKey(&xField, &yField)

	// Add the points
	resultX, resultY := curve.Add(
		rootPubKey.X(), rootPubKey.Y(),
		scalarPoint.X(), scalarPoint.Y(),
	)

	resultXField, resultYField := secp256k1.FieldVal{}, secp256k1.FieldVal{}
	resultXField.SetByteSlice(resultX.Bytes())
	resultYField.SetByteSlice(resultY.Bytes())

	// Create the final public key
	finalPubKey := secp256k1.NewPublicKey(&resultXField, &resultYField)

	// Return compressed format
	return finalPubKey.SerializeCompressed(), nil
}

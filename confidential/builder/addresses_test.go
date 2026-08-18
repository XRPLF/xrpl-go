//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// The rejection cases for these rules live in each builder's own validation table. The
// two tests here cover what those failure-only tables cannot express: that an X-address
// is accepted at all, and that it reaches the ledger and keylet layers as a classic
// address, since both decode classic addresses only.

// TestBuilderAcceptsXAddresses pins the acceptance half of the address rules.
func TestBuilderAcceptsXAddresses(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	t.Run("convert", func(t *testing.T) {
		require.NoError(t, validateConvertBase(BuildConvertParams{
			Account: xAddressOf(t, testAccount), IssuanceID: testIssuanceID, HolderPubKey: kp.PubKeyHex,
		}))
	})
	t.Run("merge inbox", func(t *testing.T) {
		require.NoError(t, validateMergeInboxBase(BuildMergeInboxParams{
			Account: xAddressOf(t, testAccount), IssuanceID: testIssuanceID,
		}))
	})
	t.Run("send", func(t *testing.T) {
		require.NoError(t, validateSendBase(BuildSendParams{
			Account: xAddressOf(t, testAccount), Destination: xAddressOf(t, testDestination),
			IssuanceID: testIssuanceID, Amount: 1,
			SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex,
		}))
	})
	t.Run("convert with a tagged account", func(t *testing.T) {
		// Account has a SourceTag companion field, so the embedded tag has somewhere to
		// land and the builder keeps it.
		require.NoError(t, validateConvertBase(BuildConvertParams{
			Account: taggedXAddressOf(t, testAccount, 42), IssuanceID: testIssuanceID, HolderPubKey: kp.PubKeyHex,
		}))
	})
	t.Run("send with a tagged destination and no explicit tag", func(t *testing.T) {
		require.NoError(t, validateSendBase(BuildSendParams{
			Account: testAccount, Destination: taggedXAddressOf(t, testDestination, 42),
			IssuanceID: testIssuanceID, Amount: 1,
			SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex,
		}))
	})
	t.Run("clawback", func(t *testing.T) {
		require.NoError(t, validateClawbackBase(BuildClawbackParams{
			Account: xAddressOf(t, testAccount), Holder: xAddressOf(t, testDestination),
			IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex,
		}))
	})
}

// TestQueryAddressNormalization pins the conversion to the classic form. xrplhash.MPToken
// and the account query decode classic addresses only, so an X-address that reached them
// unconverted would surface much later as an unrelated error.
func TestQueryAddressNormalization(t *testing.T) {
	t.Run("both address forms resolve to the same keylet", func(t *testing.T) {
		fromClassic, err := mpTokenIndex(testIssuanceID, testAccount)
		require.NoError(t, err)
		fromTaggedXAddress, err := mpTokenIndex(testIssuanceID, taggedXAddressOf(t, testAccount, 42))
		require.NoError(t, err)
		require.Equal(t, fromClassic, fromTaggedXAddress)
	})

	t.Run("the account query receives the classic form", func(t *testing.T) {
		q := &mockQuerier{accountSeq: 7}
		seq, err := getSequence(q, taggedXAddressOf(t, testAccount, 42))
		require.NoError(t, err)
		require.Equal(t, uint32(7), seq)
		require.Equal(t, types.Address(testAccount), q.lastAccountReq.Account)
	})

	t.Run("a malformed address fails before any ledger query", func(t *testing.T) {
		q := &mockQuerier{}
		_, err := getSequence(q, "notanaddress")
		require.ErrorIs(t, err, ErrInvalidAccount)
		require.NotErrorIs(t, err, ErrLedgerQuery)
		require.Zero(t, q.queryCalls)

		// The keylet helper serves Account, Destination, and Holder alike, so it names no
		// field. The field-specific report is each builder's own validation.
		_, err = mpTokenIndex(testIssuanceID, "notanaddress")
		require.ErrorIs(t, err, ErrInvalidAddress)
		require.NotErrorIs(t, err, ErrInvalidHolder)
	})
}

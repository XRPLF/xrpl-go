package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestSponsorshipTransfer_TxType(t *testing.T) {
	var tx Tx = &SponsorshipTransfer{}
	require.Equal(t, TxType("SponsorshipTransfer"), tx.TxType())
}

func TestSponsorshipTransfer_Validate(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const sponsor = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
	object := types.Hash256(strings.Repeat("AB", 32))
	emptyObject, shortObject := types.Hash256(""), types.Hash256("ABC")
	longObject, nonHexObject := types.Hash256(strings.Repeat("A", 66)), types.Hash256(strings.Repeat("G", 64))
	zeroObject, lowerObject := types.Hash256(strings.Repeat("0", 64)), types.Hash256(strings.Repeat("ab", 32))
	key, sig := "AB", "CD"
	single := &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &sig}
	multi := &types.SponsorSignature{Signers: []types.Signer{{SignerData: types.SignerData{
		Account: account, SigningPubKey: key, TxnSignature: sig,
	}}}}
	tests := []struct {
		name string
		tx   SponsorshipTransfer
		err  error
	}{
		{name: "end own account", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}}},
		{name: "end another account", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, Sponsee: sponsor}},
		{name: "end own object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &object}},
		{name: "end another account's object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &object, Sponsee: sponsor}},
		{name: "create prefunded object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}, ObjectID: &object}},
		{name: "reassign prefunded object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipReassign, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}, ObjectID: &object}},
		{name: "create account with single signature", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve, SponsorSignature: single}}},
		{name: "reassign account with multi signature", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipReassign, Sponsor: sponsor, SponsorFlags: types.SpfSponsorFee | types.SpfSponsorReserve, SponsorSignature: multi}}},
		{name: "canonical flag", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd | types.TfFullyCanonicalSig}}},
		{name: "inner prefunded object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate | types.TfInnerBatchTxn, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}, ObjectID: &object}},
		{name: "inner account authorization placeholder", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate | types.TfInnerBatchTxn, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve, SponsorSignature: &types.SponsorSignature{}}}},
		{name: "inner account missing authorization field", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate | types.TfInnerBatchTxn, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}}, err: ErrSponsorshipTransferSignatureRequired},
		{name: "outer empty authorization invalid", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve, SponsorSignature: &types.SponsorSignature{}}}, err: ErrInvalidSponsorSignature},
		{name: "missing scenario", err: ErrInvalidFlags},
		{name: "universal flag without scenario", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: types.TfUniversal}}, err: ErrInvalidFlags},
		{name: "create and reassign", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate | TfSponsorshipReassign}}, err: ErrInvalidFlags},
		{name: "create and end", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate | TfSponsorshipEnd}}, err: ErrInvalidFlags},
		{name: "reassign and end", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipReassign | TfSponsorshipEnd}}, err: ErrInvalidFlags},
		{name: "all scenarios", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate | TfSponsorshipReassign | TfSponsorshipEnd}}, err: ErrInvalidFlags},
		{name: "unknown flag", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd | 1}}, err: ErrInvalidFlags},
		{name: "empty object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &emptyObject}, err: ErrSponsorshipTransferObjectID},
		{name: "short object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &shortObject}, err: ErrSponsorshipTransferObjectID},
		{name: "long object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &longObject}, err: ErrSponsorshipTransferObjectID},
		{name: "nonhex object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &nonHexObject}, err: ErrSponsorshipTransferObjectID},
		{name: "zero hash is still object-level", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}, ObjectID: &zeroObject}},
		{name: "lowercase object", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &lowerObject}},
		{name: "end with reserve sponsor", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}}, err: ErrSponsorshipTransferSponsorNotAllowed},
		{name: "end with fee sponsor", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd, Sponsor: sponsor, SponsorFlags: types.SpfSponsorFee}}, err: ErrSponsorshipTransferSponsorNotAllowed},
		{name: "end with orphan sponsor flags", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd, SponsorFlags: types.SpfSponsorReserve}}, err: ErrSponsorFieldsMissing},
		{name: "end with orphan signature", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd, SponsorSignature: single}}, err: ErrSponsorFieldsMissing},
		{name: "create without sponsor", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate}, ObjectID: &object}, err: ErrSponsorshipTransferSponsorRequired},
		{name: "reassign without sponsor", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipReassign}, ObjectID: &object}, err: ErrSponsorshipTransferSponsorRequired},
		{name: "sponsor without flags", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor}, ObjectID: &object}, err: ErrSponsorFieldsMissing},
		{name: "create without reserve flag", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorFee}, ObjectID: &object}, err: ErrSponsorshipTransferReserveRequired},
		{name: "reassign without reserve flag", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipReassign, Sponsor: sponsor, SponsorFlags: types.SpfSponsorFee}, ObjectID: &object}, err: ErrSponsorshipTransferReserveRequired},
		{name: "create with sponsee", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}, ObjectID: &object, Sponsee: sponsor}, err: ErrSponsorshipTransferSponseeNotAllowed},
		{name: "reassign with sponsee", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipReassign, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}, ObjectID: &object, Sponsee: sponsor}, err: ErrSponsorshipTransferSponseeNotAllowed},
		{name: "unsigned account create", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}}, err: ErrSponsorshipTransferSignatureRequired},
		{name: "unsigned account reassign", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipReassign, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}}, err: ErrSponsorshipTransferSignatureRequired},
		{name: "reserve delegation forbidden", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve, Delegate: sponsor}, ObjectID: &object}, err: ErrSponsorDelegateConflict},
		{name: "signed inner account forbidden", tx: SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate | types.TfInnerBatchTxn, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve, SponsorSignature: single}}, err: ErrInnerBatchSponsorSignature},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.tx
			tx.Account = account
			tx.TransactionType = SponsorshipTransferTx
			before, err := json.Marshal(tx)
			require.NoError(t, err)
			valid, err := tx.Validate()
			require.Equal(t, tt.err == nil, valid)
			require.ErrorIs(t, err, tt.err)
			after, err := json.Marshal(tx)
			require.NoError(t, err)
			require.Equal(t, before, after)
		})
	}
}

func TestSponsorshipTransfer_Sponsee(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	selfX, err := addresscodec.ClassicAddressToXAddress(account, 0, false, true)
	require.NoError(t, err)
	tests := []struct {
		name    string
		sponsee types.Address
		err     error
	}{
		{"self classic", account, ErrSponsorshipAccountConflict},
		{"self X-address", types.Address(selfX), ErrSponsorshipAccountConflict},
		{"invalid", "invalid", ErrInvalidSponsee},
		{"zero", "rrrrrrrrrrrrrrrrrrrrrhoLvTp", ErrZeroAccountID},
		{"tagged", "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGZMhc9YTE92ehJ2Fu", ErrAccountIDTagNotAllowed},
		{"untagged", "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SponsorshipTransfer{
				BaseTx:  BaseTx{Account: account, TransactionType: SponsorshipTransferTx, Flags: TfSponsorshipEnd},
				Sponsee: tt.sponsee,
			}
			valid, err := tx.Validate()
			require.Equal(t, tt.err == nil, valid)
			require.ErrorIs(t, err, tt.err)
			if tt.err != nil {
				require.ErrorIs(t, err, ErrInvalidSponsee)
			}
		})
	}
}

func TestSponsorshipTransfer_BaseValidation(t *testing.T) {
	tx := SponsorshipTransfer{BaseTx: BaseTx{Account: "invalid", TransactionType: SponsorshipTransferTx, Flags: TfSponsorshipEnd}}
	valid, err := tx.Validate()
	require.False(t, valid)
	require.ErrorIs(t, err, ErrInvalidAccount)
}

// Validate must reject a type mismatch even when both types allow reserve sponsorship.
func TestSponsorshipTransfer_ValidateTransactionTypeMismatch(t *testing.T) {
	object := types.Hash256(strings.Repeat("AB", 32))
	tx := SponsorshipTransfer{
		BaseTx: BaseTx{
			Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: PaymentTx,
			Flags:   TfSponsorshipCreate,
			Sponsor: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", SponsorFlags: types.SpfSponsorReserve,
		},
		ObjectID: &object,
	}
	valid, err := tx.Validate()
	require.False(t, valid)
	require.ErrorIs(t, err, ErrInvalidTransactionType)
	require.Equal(t, PaymentTx, tx.TransactionType)
}

func TestSponsorshipTransfer_Flatten(t *testing.T) {
	object, empty := types.Hash256(strings.Repeat("AB", 32)), types.Hash256("")
	tests := []struct {
		name string
		tx   SponsorshipTransfer
		want FlatTransaction
	}{
		{"omitted fields", SponsorshipTransfer{}, FlatTransaction{"TransactionType": "SponsorshipTransfer"}},
		{"present empty object is not omitted", SponsorshipTransfer{ObjectID: &empty}, FlatTransaction{"TransactionType": "SponsorshipTransfer", "ObjectID": ""}},
		{"end object's sponsorship", SponsorshipTransfer{
			BaseTx:   BaseTx{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", Fee: 12, Sequence: 1, Flags: TfSponsorshipEnd},
			ObjectID: &object, Sponsee: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		}, FlatTransaction{
			"TransactionType": "SponsorshipTransfer", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Fee": "12", "Sequence": uint32(1),
			"Flags": uint32(0x00010000), "ObjectID": strings.Repeat("AB", 32), "Sponsee": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, err := json.Marshal(tt.tx)
			require.NoError(t, err)
			require.Equal(t, tt.want, tt.tx.Flatten())
			after, err := json.Marshal(tt.tx)
			require.NoError(t, err)
			require.Equal(t, before, after)
		})
	}
}

func TestSponsorshipTransfer_CodecRoundTrip(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const sponsor = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
	object := types.Hash256(strings.Repeat("AB", 32))
	key, signature := "AB", "CD"
	tests := []struct {
		name string
		tx   SponsorshipTransfer
		want map[string]any
	}{
		{"end own account", SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}}, map[string]any{
			"TransactionType": "SponsorshipTransfer", "Account": account, "Flags": uint32(0x00010000),
		}},
		{"end another account's object", SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipEnd}, ObjectID: &object, Sponsee: sponsor}, map[string]any{
			"TransactionType": "SponsorshipTransfer", "Account": account, "Flags": uint32(0x00010000), "ObjectID": strings.Repeat("AB", 32), "Sponsee": sponsor,
		}},
		{"create prefunded object", SponsorshipTransfer{BaseTx: BaseTx{Flags: TfSponsorshipCreate, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve}, ObjectID: &object}, map[string]any{
			"TransactionType": "SponsorshipTransfer", "Account": account, "Flags": uint32(0x00020000), "ObjectID": strings.Repeat("AB", 32), "Sponsor": sponsor, "SponsorFlags": uint32(2),
		}},
		{"reassign co-signed account", SponsorshipTransfer{BaseTx: BaseTx{
			Flags: TfSponsorshipReassign, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve,
			SponsorSignature: &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &signature},
		}}, map[string]any{
			"TransactionType": "SponsorshipTransfer", "Account": account, "Flags": uint32(0x00040000), "Sponsor": sponsor, "SponsorFlags": uint32(2),
			"SponsorSignature": map[string]any{"SigningPubKey": "AB", "TxnSignature": "CD"},
		}},
		{"inner account placeholder", SponsorshipTransfer{BaseTx: BaseTx{
			Flags: TfSponsorshipCreate | types.TfInnerBatchTxn, Sponsor: sponsor, SponsorFlags: types.SpfSponsorReserve,
			SponsorSignature: &types.SponsorSignature{},
		}}, map[string]any{
			"TransactionType": "SponsorshipTransfer", "Account": account, "Flags": uint32(0x40020000), "Sponsor": sponsor, "SponsorFlags": uint32(2),
			"SigningPubKey": "", "SponsorSignature": map[string]any{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.tx
			tx.Account = account
			tx.TransactionType = SponsorshipTransferTx
			valid, err := tx.Validate()
			require.True(t, valid)
			require.NoError(t, err)
			blob, err := binarycodec.Encode(tx.Flatten())
			require.NoError(t, err)
			got, err := binarycodec.Decode(blob)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSponsorshipTransfer_JSONPresence(t *testing.T) {
	object := types.Hash256("13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8")
	tests := []struct {
		name string
		tx   SponsorshipTransfer
		json string
	}{
		{"account", SponsorshipTransfer{}, `{"Account":"","TransactionType":""}`},
		{
			"object",
			SponsorshipTransfer{ObjectID: &object, Sponsee: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"},
			`{"Account":"","TransactionType":"","ObjectID":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8","Sponsee":"r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.tx)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
			var got SponsorshipTransfer
			require.NoError(t, json.Unmarshal([]byte(tt.json), &got))
			require.Equal(t, tt.tx, got)
		})
	}
}

func TestSponsorshipTransfer_Flags(t *testing.T) {
	tests := []struct {
		name string
		set  func(*SponsorshipTransfer)
		want uint32
	}{
		{"end", (*SponsorshipTransfer).SetSponsorshipEndFlag, 0x00010000},
		{"create", (*SponsorshipTransfer).SetSponsorshipCreateFlag, 0x00020000},
		{"reassign", (*SponsorshipTransfer).SetSponsorshipReassignFlag, 0x00040000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SponsorshipTransfer{BaseTx: BaseTx{Flags: types.TfFullyCanonicalSig}}
			tt.set(&tx)
			tt.set(&tx)
			require.Equal(t, types.TfFullyCanonicalSig|tt.want, tx.Flags)
		})
	}
}

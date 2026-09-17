package transaction

import (
	"encoding/json"
	"math"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestSponsorshipSet_TxType(t *testing.T) {
	var tx Tx = &SponsorshipSet{}
	require.Equal(t, TxType("SponsorshipSet"), tx.TxType())
}

func TestSponsorshipSet_Validate(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const sponsee = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
	positive, negative, zero := "10", "-10", "0"
	reserve, zeroReserve, minReserve, maxReserve := int32(-1), int32(0), int32(math.MinInt32), int32(math.MaxInt32)
	zeroFee, maxFee := types.XRPCurrencyAmount(0), types.XRPCurrencyAmount(100000000000000000)
	tests := []struct {
		name string
		tx   SponsorshipSet
		err  error
	}{
		{name: "top up", tx: SponsorshipSet{Sponsee: sponsee, FeeAmountDelta: &positive}},
		{name: "reduce existing budgets", tx: SponsorshipSet{Sponsee: sponsee, FeeAmountDelta: &negative, RemainingOwnerCountDelta: &reserve}},
		{name: "remove fee limit", tx: SponsorshipSet{Sponsee: sponsee, MaxFee: &zeroFee}},
		{name: "limit can exceed delta", tx: SponsorshipSet{Sponsee: sponsee, FeeAmountDelta: &positive, MaxFee: &maxFee}},
		{name: "minimum reserve delta", tx: SponsorshipSet{Sponsee: sponsee, RemainingOwnerCountDelta: &minReserve}},
		{name: "maximum reserve delta", tx: SponsorshipSet{Sponsee: sponsee, RemainingOwnerCountDelta: &maxReserve}},
		{name: "set fee signing", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipSetRequireSignForFee}, Sponsee: sponsee}},
		{name: "clear fee signing", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipClearRequireSignForFee}, Sponsee: sponsee}},
		{name: "set reserve signing", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipSetRequireSignForReserve}, Sponsee: sponsee}},
		{name: "clear reserve signing", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipClearRequireSignForReserve}, Sponsee: sponsee}},
		{name: "independent fee and reserve flags", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipSetRequireSignForFee | TfSponsorshipClearRequireSignForReserve}, Sponsee: sponsee}},
		{name: "universal flags with update", tx: SponsorshipSet{BaseTx: BaseTx{Flags: types.TfUniversal}, Sponsee: sponsee, FeeAmountDelta: &positive}},
		{name: "delete by sponsor", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject}, Sponsee: sponsee}},
		{name: "delete by sponsee", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject | types.TfFullyCanonicalSig}, CounterpartySponsor: sponsee}},
		{name: "missing counterparty", tx: SponsorshipSet{FeeAmountDelta: &positive}, err: ErrSponsorshipSetCounterpartyConflict},
		{name: "both counterparties", tx: SponsorshipSet{Sponsee: sponsee, CounterpartySponsor: sponsee, FeeAmountDelta: &positive}, err: ErrSponsorshipSetCounterpartyConflict},
		{name: "sponsee cannot update", tx: SponsorshipSet{CounterpartySponsor: sponsee, FeeAmountDelta: &positive}, err: ErrSponsorshipSetCounterpartyCannotModify},
		{name: "sponsee cannot modify flags", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipSetRequireSignForFee}, CounterpartySponsor: sponsee}, err: ErrSponsorshipSetCounterpartyCannotModify},
		{name: "empty update", tx: SponsorshipSet{Sponsee: sponsee}, err: ErrSponsorshipSetEmptyUpdate},
		{name: "universal flags alone", tx: SponsorshipSet{BaseTx: BaseTx{Flags: types.TfUniversal}, Sponsee: sponsee}, err: ErrSponsorshipSetEmptyUpdate},
		{name: "unknown flag", tx: SponsorshipSet{BaseTx: BaseTx{Flags: 1}, Sponsee: sponsee, FeeAmountDelta: &positive}, err: ErrInvalidFlags},
		{name: "fee signing conflict", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipSetRequireSignForFee | TfSponsorshipClearRequireSignForFee}, Sponsee: sponsee}, err: ErrInvalidFlags},
		{name: "reserve signing conflict", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfSponsorshipSetRequireSignForReserve | TfSponsorshipClearRequireSignForReserve}, Sponsee: sponsee}, err: ErrInvalidFlags},
		{name: "delete with set fee flag", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject | TfSponsorshipSetRequireSignForFee}, Sponsee: sponsee}, err: ErrInvalidFlags},
		{name: "delete with clear fee flag", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject | TfSponsorshipClearRequireSignForFee}, Sponsee: sponsee}, err: ErrInvalidFlags},
		{name: "delete with set reserve flag", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject | TfSponsorshipSetRequireSignForReserve}, Sponsee: sponsee}, err: ErrInvalidFlags},
		{name: "delete with clear reserve flag", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject | TfSponsorshipClearRequireSignForReserve}, Sponsee: sponsee}, err: ErrInvalidFlags},
		{name: "delete with fee delta", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject}, Sponsee: sponsee, FeeAmountDelta: &positive}, err: ErrSponsorshipSetDeleteConflict},
		{name: "delete with explicit zero delta", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject}, Sponsee: sponsee, FeeAmountDelta: &zero}, err: ErrSponsorshipSetDeleteConflict},
		{name: "delete with reserve delta", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject}, Sponsee: sponsee, RemainingOwnerCountDelta: &reserve}, err: ErrSponsorshipSetDeleteConflict},
		{name: "delete with explicit zero reserve", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject}, Sponsee: sponsee, RemainingOwnerCountDelta: &zeroReserve}, err: ErrSponsorshipSetDeleteConflict},
		{name: "delete with explicit zero limit", tx: SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject}, Sponsee: sponsee, MaxFee: &zeroFee}, err: ErrSponsorshipSetDeleteConflict},
		{name: "zero reserve update", tx: SponsorshipSet{Sponsee: sponsee, RemainingOwnerCountDelta: &zeroReserve}, err: ErrSponsorshipSetRemainingOwnerCountDelta},
		{name: "fee-sponsored update", tx: SponsorshipSet{BaseTx: BaseTx{Sponsor: sponsee, SponsorFlags: types.SpfSponsorFee}, Sponsee: sponsee, FeeAmountDelta: &positive}},
		{name: "reserve-sponsored update forbidden", tx: SponsorshipSet{BaseTx: BaseTx{Sponsor: sponsee, SponsorFlags: types.SpfSponsorReserve}, Sponsee: sponsee, FeeAmountDelta: &positive}, err: ErrReserveSponsorshipNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.tx
			tx.Account = account
			tx.TransactionType = SponsorshipSetTx
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

func TestSponsorshipSet_FeeAmountDelta(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{"1", true},
		{"-1", true},
		{"9007199254740993", true},
		{"-9007199254740993", true},
		{"100000000000000000", true},
		{"-100000000000000000", true},
		{"100000000000000001", false},
		{"-100000000000000001", false},
		{"9223372036854775808", false},
		{"-9223372036854775809", false},
		{"", false},
		{"0", false},
		{"-0", false},
		{"+1", false},
		{"01", false},
		{"-01", false},
		{" 1", false},
		{"1 ", false},
		{"1\n", false},
		{"1.0", false},
		{"1e2", false},
		{"abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			tx := SponsorshipSet{
				BaseTx:  BaseTx{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: SponsorshipSetTx},
				Sponsee: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", FeeAmountDelta: &tt.value,
			}
			valid, err := tx.Validate()
			require.Equal(t, tt.valid, valid)
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrSponsorshipSetFeeAmountDelta)
			}
		})
	}
}

func TestSponsorshipSet_MaxFee(t *testing.T) {
	tests := []struct {
		name  string
		value types.XRPCurrencyAmount
		err   error
	}{
		{"zero", 0, nil},
		{"one drop", 1, nil},
		{"above float precision", 9007199254740993, nil},
		{"native limit", 100000000000000000, nil},
		{"above native limit", 100000000000000001, ErrSponsorshipSetMaxFee},
		{"uint64 limit", math.MaxUint64, ErrSponsorshipSetMaxFee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SponsorshipSet{
				BaseTx:  BaseTx{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: SponsorshipSetTx},
				Sponsee: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", MaxFee: &tt.value,
			}
			valid, err := tx.Validate()
			require.Equal(t, tt.err == nil, valid)
			require.ErrorIs(t, err, tt.err)
		})
	}
}

func TestSponsorshipSet_Counterparty(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const other = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
	selfX, err := addresscodec.ClassicAddressToXAddress(account, 0, false, false)
	require.NoError(t, err)
	otherX, err := addresscodec.ClassicAddressToXAddress(other, 0, false, true)
	require.NoError(t, err)
	taggedX, err := addresscodec.ClassicAddressToXAddress(other, 0, true, false)
	require.NoError(t, err)
	tests := []struct {
		name         string
		sponsee      types.Address
		counterparty types.Address
		err          error
		fieldErr     error
	}{
		{"sponsee X-address", types.Address(otherX), "", nil, nil},
		{"counterparty X-address", "", types.Address(otherX), nil, nil},
		{"self sponsee", account, "", ErrSponsorshipAccountConflict, ErrInvalidSponsee},
		{"self counterparty", "", account, ErrSponsorshipAccountConflict, ErrInvalidCounterpartySponsor},
		{"equivalent sponsee", types.Address(selfX), "", ErrSponsorshipAccountConflict, ErrInvalidSponsee},
		{"equivalent counterparty", "", types.Address(selfX), ErrSponsorshipAccountConflict, ErrInvalidCounterpartySponsor},
		{"invalid sponsee", "invalid", "", ErrInvalidSponsee, ErrInvalidSponsee},
		{"invalid counterparty", "", "invalid", ErrInvalidCounterpartySponsor, ErrInvalidCounterpartySponsor},
		{"zero sponsee", "rrrrrrrrrrrrrrrrrrrrrhoLvTp", "", ErrZeroAccountID, ErrInvalidSponsee},
		{"zero counterparty", "", "rrrrrrrrrrrrrrrrrrrrrhoLvTp", ErrZeroAccountID, ErrInvalidCounterpartySponsor},
		{"tagged sponsee", types.Address(taggedX), "", ErrAccountIDTagNotAllowed, ErrInvalidSponsee},
		{"tagged counterparty", "", types.Address(taggedX), ErrAccountIDTagNotAllowed, ErrInvalidCounterpartySponsor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SponsorshipSet{
				BaseTx:  BaseTx{Account: account, TransactionType: SponsorshipSetTx, Flags: TfDeleteObject},
				Sponsee: tt.sponsee, CounterpartySponsor: tt.counterparty,
			}
			valid, err := tx.Validate()
			require.Equal(t, tt.err == nil, valid)
			require.ErrorIs(t, err, tt.err)
			require.ErrorIs(t, err, tt.fieldErr)
		})
	}
}

func TestSponsorshipSet_BaseValidation(t *testing.T) {
	tx := SponsorshipSet{BaseTx: BaseTx{Account: "invalid", TransactionType: SponsorshipSetTx}}
	valid, err := tx.Validate()
	require.False(t, valid)
	require.ErrorIs(t, err, ErrInvalidAccount)
}

// Validate must reject a type mismatch rather than apply Payment's reserve policy.
func TestSponsorshipSet_ValidateTransactionTypeMismatch(t *testing.T) {
	delta := "1"
	tx := SponsorshipSet{
		BaseTx: BaseTx{
			Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: PaymentTx,
			Sponsor: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", SponsorFlags: types.SpfSponsorReserve,
		},
		Sponsee: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", FeeAmountDelta: &delta,
	}
	valid, err := tx.Validate()
	require.False(t, valid)
	require.ErrorIs(t, err, ErrInvalidTransactionType)
	require.Equal(t, PaymentTx, tx.TransactionType)
}

func TestSponsorshipSet_Flatten(t *testing.T) {
	delta, reserve, zero := "-9007199254740993", int32(-2), types.XRPCurrencyAmount(0)
	tests := []struct {
		name string
		tx   SponsorshipSet
		want FlatTransaction
	}{
		{"omitted fields", SponsorshipSet{}, FlatTransaction{"TransactionType": "SponsorshipSet"}},
		{"negative deltas and explicit zero limit", SponsorshipSet{
			BaseTx:  BaseTx{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", Fee: 12, Sequence: 1},
			Sponsee: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", FeeAmountDelta: &delta, MaxFee: &zero, RemainingOwnerCountDelta: &reserve,
		}, FlatTransaction{
			"TransactionType": "SponsorshipSet", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Fee": "12", "Sequence": uint32(1),
			"Sponsee": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", "FeeAmountDelta": "-9007199254740993", "MaxFee": "0", "RemainingOwnerCountDelta": int32(-2),
		}},
		{"sponsee deletion", SponsorshipSet{BaseTx: BaseTx{Flags: TfDeleteObject}, CounterpartySponsor: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"}, FlatTransaction{
			"TransactionType": "SponsorshipSet", "Flags": uint32(0x00100000), "CounterpartySponsor": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
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

func TestSponsorshipSet_CodecRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		delta   string
		reserve int32
	}{
		{"positive bounds", "100000000000000000", math.MaxInt32},
		{"negative bounds", "-100000000000000000", math.MinInt32},
		{"exact negative above float precision", "-9007199254740993", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zero := types.XRPCurrencyAmount(0)
			tx := SponsorshipSet{
				BaseTx:  BaseTx{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: SponsorshipSetTx, Fee: 12, Sequence: 1},
				Sponsee: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", FeeAmountDelta: &tt.delta, MaxFee: &zero, RemainingOwnerCountDelta: &tt.reserve,
			}
			valid, err := tx.Validate()
			require.True(t, valid)
			require.NoError(t, err)
			blob, err := binarycodec.Encode(tx.Flatten())
			require.NoError(t, err)
			got, err := binarycodec.Decode(blob)
			require.NoError(t, err)
			require.Equal(t, map[string]any{
				"TransactionType": "SponsorshipSet", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Fee": "12", "Sequence": uint32(1),
				"Sponsee": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", "FeeAmountDelta": tt.delta, "MaxFee": "0", "RemainingOwnerCountDelta": tt.reserve,
			}, got)
		})
	}
}

func TestSponsorshipSet_JSONPresence(t *testing.T) {
	delta, reserve, zero := "-10", int32(-1), types.XRPCurrencyAmount(0)
	tests := []struct {
		name string
		tx   SponsorshipSet
		json string
	}{
		{"absent", SponsorshipSet{}, `{"Account":"","TransactionType":""}`},
		{
			"present",
			SponsorshipSet{FeeAmountDelta: &delta, MaxFee: &zero, RemainingOwnerCountDelta: &reserve},
			`{"Account":"","TransactionType":"","FeeAmountDelta":"-10","MaxFee":"0","RemainingOwnerCountDelta":-1}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.tx)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
			var got SponsorshipSet
			require.NoError(t, json.Unmarshal([]byte(tt.json), &got))
			require.Equal(t, tt.tx, got)
		})
	}
}

func TestSponsorshipSet_JSONNumericTypes(t *testing.T) {
	for _, payload := range []string{
		`{"RemainingOwnerCountDelta":2147483648}`, `{"RemainingOwnerCountDelta":-2147483649}`,
		`{"RemainingOwnerCountDelta":1.5}`, `{"RemainingOwnerCountDelta":"1"}`,
		`{"FeeAmountDelta":-1}`, `{"MaxFee":"-1"}`, `{"MaxFee":1}`, `{"MaxFee":"1.5"}`,
	} {
		t.Run(payload, func(t *testing.T) {
			var tx SponsorshipSet
			require.Error(t, json.Unmarshal([]byte(payload), &tx))
		})
	}
}

func TestSponsorshipSet_Flags(t *testing.T) {
	tests := []struct {
		name string
		set  func(*SponsorshipSet)
		want uint32
	}{
		{"set fee signing", (*SponsorshipSet).SetSponsorshipSetRequireSignForFeeFlag, 0x00010000},
		{"clear fee signing", (*SponsorshipSet).SetSponsorshipClearRequireSignForFeeFlag, 0x00020000},
		{"set reserve signing", (*SponsorshipSet).SetSponsorshipSetRequireSignForReserveFlag, 0x00040000},
		{"clear reserve signing", (*SponsorshipSet).SetSponsorshipClearRequireSignForReserveFlag, 0x00080000},
		{"delete", (*SponsorshipSet).SetDeleteObjectFlag, 0x00100000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := SponsorshipSet{BaseTx: BaseTx{Flags: types.TfFullyCanonicalSig}}
			tt.set(&tx)
			tt.set(&tx)
			require.Equal(t, types.TfFullyCanonicalSig|tt.want, tx.Flags)
		})
	}
}

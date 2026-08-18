package transaction

import (
	"encoding/hex"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAddrAccount  = "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU"
	testAddrDelegate = "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf"
	// testAddrZero is ACCOUNT_ZERO, which decodes cleanly in either address form but can
	// never sign, so Account and Delegate both reject it.
	testAddrZero = "rrrrrrrrrrrrrrrrrrrrrhoLvTp"
)

// xAddr encodes classic as an X-address so a case can name the form it exercises.
func xAddr(t *testing.T, classic string, tag uint32, tagged bool) types.Address {
	t.Helper()
	encoded, err := addresscodec.ClassicAddressToXAddress(classic, tag, tagged, false)
	require.NoError(t, err)
	return types.Address(encoded)
}

// signersOf builds a single-entry Signers list so a case can name the address form it
// exercises without repeating the two signature fields that are not under test.
func signersOf(account types.Address) []types.Signer {
	return []types.Signer{{SignerData: types.SignerData{
		Account:       account,
		TxnSignature:  "0123456789abcdef",
		SigningPubKey: "abcdef0123456789",
	}}}
}

func TestTx_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		tx      *BaseTx
		wantErr error
	}{
		{
			name: "Valid transaction",
			tx: &BaseTx{
				Account:            "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType:    PaymentTx,
				Fee:                types.XRPCurrencyAmount(10),
				Sequence:           1,
				AccountTxnID:       "abcdef123456",
				LastLedgerSequence: 100,
				SourceTag:          123,
				SigningPubKey:      "abcdefg",
				TicketSequence:     2,
				TxnSignature:       "xyz123",
				NetworkID:          1,
				Memos: []types.MemoWrapper{
					{
						Memo: types.Memo{
							MemoType:   hex.EncodeToString([]byte("text")),
							MemoData:   hex.EncodeToString([]byte("Hello, world!")),
							MemoFormat: hex.EncodeToString([]byte("plain")),
						},
					},
					{
						Memo: types.Memo{
							MemoType:   hex.EncodeToString([]byte("text")),
							MemoData:   hex.EncodeToString([]byte("Hello, world 2!")),
							MemoFormat: hex.EncodeToString([]byte("plain")),
						},
					},
				},
				Signers: []types.Signer{
					{
						SignerData: types.SignerData{
							Account:       "rDqbKhee18wUCnvjPjZA5Kgpe4zeubLQUC",
							TxnSignature:  "abc123",
							SigningPubKey: "def456",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Missing required Account field",
			tx: &BaseTx{
				TransactionType: PaymentTx,
			},
			wantErr: ErrInvalidAccount,
		},
		{
			name: "Missing required TransactionType field",
			tx: &BaseTx{
				Account: "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
			},
			wantErr: ErrInvalidTransactionType,
		},
		{
			name: "Invalid memos",
			tx: &BaseTx{
				Account:            "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType:    PaymentTx,
				Fee:                types.XRPCurrencyAmount(10),
				Sequence:           1,
				AccountTxnID:       "abcdef123456",
				LastLedgerSequence: 100,
				SourceTag:          123,
				SigningPubKey:      "abcdefg",
				TicketSequence:     2,
				TxnSignature:       "xyz123",
				Memos: []types.MemoWrapper{
					{
						Memo: types.Memo{
							MemoType:   "invalid",
							MemoData:   "Hello, world!",
							MemoFormat: "plain",
						},
					},
				},
			},
			wantErr: ErrMemoDataShouldBeHex,
		},
		{
			name: "Invalid signers",
			tx: &BaseTx{
				Account:            "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType:    PaymentTx,
				Fee:                types.XRPCurrencyAmount(10),
				Sequence:           1,
				AccountTxnID:       "abcdef123456",
				LastLedgerSequence: 100,
				SourceTag:          123,
				SigningPubKey:      "abcdefg",
				TicketSequence:     2,
				TxnSignature:       "xyz123",
				Signers: []types.Signer{
					{
						SignerData: types.SignerData{
							Account: "rDqbKhee18wUCnvjPjZA5Kgpe4zeubLQUC",
						},
					},
				},
			},
			wantErr: ErrSignerShouldHaveThreeFields,
		},
		{
			name: "Valid transaction with Delegate",
			tx: &BaseTx{
				Account:         "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType: PaymentTx,
				Fee:             types.XRPCurrencyAmount(10),
				Sequence:        1,
				Delegate:        "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
			wantErr: nil,
		},
		{
			name: "Invalid Delegate address",
			tx: &BaseTx{
				Account:         "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType: PaymentTx,
				Fee:             types.XRPCurrencyAmount(10),
				Sequence:        1,
				Delegate:        "invalid_address",
			},
			wantErr: ErrInvalidDelegate,
		},
		{
			name: "Delegate same as Account",
			tx: &BaseTx{
				Account:         "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType: PaymentTx,
				Fee:             types.XRPCurrencyAmount(10),
				Sequence:        1,
				Delegate:        "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
			},
			wantErr: ErrDelegateAccountConflict,
		},
		// A classic address and its X-address form name the same account, so they must
		// validate alike. An embedded tag is accepted only where the transaction has a
		// companion field to carry it.
		{
			name:    "Untagged X-address Account",
			tx:      &BaseTx{Account: xAddr(t, testAddrAccount, 0, false), TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: nil,
		},
		{
			name:    "Tagged X-address Account without SourceTag",
			tx:      &BaseTx{Account: xAddr(t, testAddrAccount, 42, true), TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: nil,
		},
		{
			// BaseTx does not reject the pairing. Client autofill rewrites Account to its
			// classic form and carries the tag across, so the transaction reaches the
			// ledger intact and rejecting it here would refuse input that works today.
			// The transaction types that resolve the tag themselves check it: see
			// Clawback and TestConfidentialMPTRejectsDuplicateAccountTag.
			name:    "Tagged X-address Account with SourceTag is left to the transaction type",
			tx:      &BaseTx{Account: xAddr(t, testAddrAccount, 42, true), SourceTag: 42, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: nil,
		},
		{
			name:    "ACCOUNT_ZERO Account",
			tx:      &BaseTx{Account: testAddrZero, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: ErrInvalidAccount,
		},
		{
			name:    "ACCOUNT_ZERO Account as X-address",
			tx:      &BaseTx{Account: xAddr(t, testAddrZero, 0, false), TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: ErrInvalidAccount,
		},
		// A pseudo-transaction is generated by consensus, and the binary codec requires
		// its Account to be ACCOUNT_ZERO, so the signer rules must not reject it.
		{
			name:    "ACCOUNT_ZERO Account on a UNLModify pseudo-transaction",
			tx:      &BaseTx{Account: testAddrZero, TransactionType: UNLModifyTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: nil,
		},
		{
			name:    "ACCOUNT_ZERO Account on a SetFee pseudo-transaction",
			tx:      &BaseTx{Account: testAddrZero, TransactionType: SetFeeTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: nil,
		},
		{
			name:    "ACCOUNT_ZERO Account on an EnableAmendment pseudo-transaction",
			tx:      &BaseTx{Account: testAddrZero, TransactionType: EnableAmendmentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: nil,
		},
		{
			name:    "Untagged X-address Delegate",
			tx:      &BaseTx{Account: testAddrAccount, Delegate: xAddr(t, testAddrDelegate, 0, false), TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: nil,
		},
		{
			name:    "Tagged X-address Delegate",
			tx:      &BaseTx{Account: testAddrAccount, Delegate: xAddr(t, testAddrDelegate, 42, true), TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: ErrAccountIDTagNotAllowed,
		},
		{
			name:    "Delegate is the Account in X-address form",
			tx:      &BaseTx{Account: testAddrAccount, Delegate: xAddr(t, testAddrAccount, 0, false), TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: ErrDelegateAccountConflict,
		},
		{
			name:    "ACCOUNT_ZERO Delegate",
			tx:      &BaseTx{Account: testAddrAccount, Delegate: testAddrZero, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			wantErr: ErrInvalidDelegate,
		},
		// A Signers entry names an account that must sign, so it carries the same rules as
		// Account and Delegate. A tagged X-address is the dangerous case: Signer has no tag
		// field, and the binary codec routes an embedded tag by field name, so an unrejected
		// one is written as a SourceTag inside the Signer object instead of failing the encode.
		{
			name:    "Untagged X-address Signers account",
			tx:      &BaseTx{Account: testAddrAccount, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10), Signers: signersOf(xAddr(t, testAddrDelegate, 0, false))},
			wantErr: nil,
		},
		{
			name:    "Tagged X-address Signers account",
			tx:      &BaseTx{Account: testAddrAccount, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10), Signers: signersOf(xAddr(t, testAddrDelegate, 42, true))},
			wantErr: ErrAccountIDTagNotAllowed,
		},
		{
			name:    "ACCOUNT_ZERO Signers account",
			tx:      &BaseTx{Account: testAddrAccount, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10), Signers: signersOf(testAddrZero)},
			wantErr: ErrZeroAccountID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := tc.tx.Validate()
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.False(t, valid)
				return
			}
			require.NoError(t, err)
			require.True(t, valid)
		})
	}
}

// TestBaseTx_ValidateAddressSentinels pins that every address failure returns a fixed
// sentinel value rather than a per-call wrap, so a caller matching with == keeps working.
// Where a sentinel names both a field and a condition, both stay matchable with errors.Is
// so a caller can tell a malformed address from ACCOUNT_ZERO or an embedded tag.
func TestBaseTx_ValidateAddressSentinels(t *testing.T) {
	testCases := []struct {
		name string
		tx   *BaseTx
		// want is the exact sentinel value the call must return, compared by identity.
		want      error
		wantField error
		wantCause error
	}{
		{
			name: "malformed Account",
			tx:   &BaseTx{Account: "notanaddress", TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			want: ErrInvalidAccount,
		},
		{
			name:      "ACCOUNT_ZERO Account",
			tx:        &BaseTx{Account: testAddrZero, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			want:      ErrAccountZero,
			wantField: ErrInvalidAccount,
			wantCause: ErrZeroAccountID,
		},
		{
			name: "malformed Delegate",
			tx:   &BaseTx{Account: testAddrAccount, Delegate: "notanaddress", TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			want: ErrInvalidDelegate,
		},
		{
			name:      "ACCOUNT_ZERO Delegate",
			tx:        &BaseTx{Account: testAddrAccount, Delegate: testAddrZero, TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			want:      ErrDelegateZero,
			wantField: ErrInvalidDelegate,
			wantCause: ErrZeroAccountID,
		},
		{
			name:      "tagged X-address Delegate",
			tx:        &BaseTx{Account: testAddrAccount, Delegate: xAddr(t, testAddrDelegate, 42, true), TransactionType: PaymentTx, Fee: types.XRPCurrencyAmount(10)},
			want:      ErrDelegateTagNotAllowed,
			wantField: ErrInvalidDelegate,
			// The encoder reports this same value, so preflight and encoding stay
			// matchable as one condition.
			wantCause: ErrAccountIDTagNotAllowed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.tx.Validate()
			// Same compares identity, not structure. require.Equal would use
			// reflect.DeepEqual, which a freshly built but identical wrap satisfies, so it
			// would miss exactly the change that breaks a caller matching with ==.
			require.Same(t, tc.want, err,
				"want the sentinel value itself, not an equal but distinct error")
			if tc.wantField != nil {
				require.ErrorIs(t, err, tc.wantField)
			}
			if tc.wantCause != nil {
				require.ErrorIs(t, err, tc.wantCause)
			}
		})
	}
}

func TestBinary_TxType(t *testing.T) {
	tx := &Binary{}
	assert.Equal(t, BinaryTx, tx.TxType())
}

func TestTxHash_TxType(t *testing.T) {
	var tx TxHash = "abcdef123456"
	assert.Equal(t, HashedTx, tx.TxType())
}

func TestBaseTx_TxType(t *testing.T) {
	tx := &BaseTx{
		TransactionType: PaymentTx,
	}
	assert.Equal(t, PaymentTx, tx.TxType())
}

func TestBaseTx_Flatten(t *testing.T) {
	testCases := []struct {
		name     string
		tx       *BaseTx
		expected string
	}{
		{
			name: "All fields populated",
			tx: &BaseTx{
				Account:            "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType:    PaymentTx,
				Fee:                types.XRPCurrencyAmount(10),
				Sequence:           1,
				AccountTxnID:       "abcdef123456",
				Delegate:           "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Flags:              2147483648,
				LastLedgerSequence: 100,
				Memos: []types.MemoWrapper{
					{
						Memo: types.Memo{
							MemoType:   hex.EncodeToString([]byte("text")),
							MemoData:   hex.EncodeToString([]byte("Hello, world!")),
							MemoFormat: hex.EncodeToString([]byte("plain")),
						},
					},
				},
				NetworkID:      1,
				Signers:        []types.Signer{{SignerData: types.SignerData{Account: "rDqbKhee18wUCnvjPjZA5Kgpe4zeubLQUC", TxnSignature: "abc123", SigningPubKey: "def456"}}},
				SourceTag:      123,
				SigningPubKey:  "abcdefg",
				TicketSequence: 2,
				TxnSignature:   "xyz123",
			},
			expected: `{
				"Account": "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				"TransactionType": "Payment",
				"Fee": "10",
				"Sequence": 1,
				"AccountTxnID": "abcdef123456",
				"Delegate": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Flags": 2147483648,
				"LastLedgerSequence": 100,
				"Memos": [
					{
						"Memo": {
							"MemoType": "74657874",
							"MemoData": "48656c6c6f2c20776f726c6421",
							"MemoFormat": "706c61696e"
						}
					}
				],
				"NetworkID": 1,
				"Signers": [
					{
						"Signer": {
							"Account": "rDqbKhee18wUCnvjPjZA5Kgpe4zeubLQUC",
							"TxnSignature": "abc123",
							"SigningPubKey": "def456"
						}
					}
				],
				"SourceTag": 123,
				"SigningPubKey": "abcdefg",
				"TicketSequence": 2,
				"TxnSignature": "xyz123"
			}`,
		},
		{
			name: "Zero Sequence preserved when TicketSequence is set",
			tx: &BaseTx{
				Account:         "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType: PaymentTx,
				Fee:             types.XRPCurrencyAmount(10),
				Sequence:        0,
				TicketSequence:  2,
			},
			expected: `{
				"Account": "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				"TransactionType": "Payment",
				"Fee": "10",
				"Sequence": 0,
				"TicketSequence": 2
			}`,
		},
		{
			name: "Sequence absent when both Sequence and TicketSequence are zero",
			tx: &BaseTx{
				Account:         "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				TransactionType: PaymentTx,
				Fee:             types.XRPCurrencyAmount(10),
				Sequence:        0,
				TicketSequence:  0,
			},
			expected: `{
				"Account": "rhbi7TGHknHCsRrVYmW57tQHmHjmFgjEpU",
				"TransactionType": "Payment",
				"Fee": "10"
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tc.tx.Flatten(), []byte(tc.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

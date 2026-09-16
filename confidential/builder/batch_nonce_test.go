package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// The tests in this file cover how the assembler resolves the sequence or Ticket of every inner.

func TestBuildBatchTicketNonces(t *testing.T) {
	const (
		accountSequence = 12
		outerTicket     = 900
		innerTicket     = 901
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(500)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, accountSequence)

	ticketed := sendOp(fixture, testAccount, testDestination, testIssuanceID, 10)
	ticketed.TicketSequence = innerTicket

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		TxOptions: TxOptions{TicketSequence: outerTicket},
		Account:   testAccount,
		Operations: []BatchOperation{
			ticketed,
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 20),
		},
	})
	require.NoError(t, err)

	require.Equal(t, uint32(outerTicket), batch.TicketSequence)
	require.Zero(t, batch.Sequence)

	first := innerOf(t, batch, 0)
	require.Equal(t, uint32(innerTicket), first["TicketSequence"])
	require.Equal(t, uint32(0), first["Sequence"])
	require.Equal(t, uint32(accountSequence), innerOf(t, batch, 1)["Sequence"])

	keys := sendKeys{
		sender:   fixture.holderKey(testAccount, testIssuanceID).PubKeyHex,
		receiver: fixture.holderKey(testDestination, testIssuanceID).PubKeyHex,
		issuer:   fixture.issuerKey(testIssuanceID).PubKeyHex,
	}
	requireSendBinding(t, firstSend(t, batch, 0), keys, spendingCiphertext(t, fixture, testAccount, testIssuanceID), innerTicket, 0)
}

func TestBuildBatchPlainInnerMatchingSequence(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, 40)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			plainAccountSet(testAccount, 41),
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 10),
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(41), innerOf(t, batch, 0)["Sequence"])
	require.Equal(t, uint32(42), innerOf(t, batch, 1)["Sequence"])
}

func TestBuildBatchPlainInner(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, 40)

	domain := "6578616D706C652E636F6D"
	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			TransactionOp{Tx: &transaction.AccountSet{
				BaseTx: transaction.BaseTx{Account: types.Address(testAccount), TransactionType: transaction.AccountSetTx},
				Domain: &domain,
			}},
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 10),
		},
	})
	require.NoError(t, err)

	plain := innerOf(t, batch, 0)
	requireInnerShape(t, plain)
	require.Equal(t, transaction.AccountSetTx.String(), plain["TransactionType"])
	require.Equal(t, domain, plain["Domain"])
	require.Equal(t, uint32(41), plain["Sequence"])
	require.Equal(t, uint32(42), innerOf(t, batch, 1)["Sequence"])
}

func TestBuildBatchEightOperations(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(1000)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, 1)

	operations := make([]BatchOperation, 0, maxBatchOperations)
	for range maxBatchOperations {
		operations = append(operations, sendOp(fixture, testAccount, testDestination, testIssuanceID, 10))
	}

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{Account: testAccount, Operations: operations})
	require.NoError(t, err)
	require.Len(t, batch.RawTransactions, maxBatchOperations)
	for index := range batch.RawTransactions {
		require.Equal(t, uint32(2+index), innerOf(t, batch, index)["Sequence"])
	}
}

func TestBuildBatchExplicitOuterSequence(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(200)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, 99)

	querier := fixture.querier()
	batch, err := BuildBatch(querier, BuildBatchParams{
		TxOptions: TxOptions{Sequence: 50},
		Account:   testAccount,
		Operations: []BatchOperation{
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 10),
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 20),
		},
	})
	require.NoError(t, err)

	require.Equal(t, uint32(50), batch.Sequence)
	require.Equal(t, uint32(51), innerOf(t, batch, 0)["Sequence"])
	require.Equal(t, uint32(52), innerOf(t, batch, 1)["Sequence"])
	require.Empty(t, querier.accounts, "an explicit outer sequence answers for the batch account too")
}

func TestBuildBatchExplicitSequencesOfAnotherAccount(t *testing.T) {
	fixture := newBatchFixture(t).
		withSequence(testDestination, 5)

	querier := fixture.querier()
	batch, err := BuildBatch(querier, BuildBatchParams{
		TxOptions: TxOptions{Sequence: 10},
		Account:   testAccount,
		Operations: []BatchOperation{
			plainAccountSet(testDestination, 5),
			plainAccountSet(testDestination, 6),
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(5), innerOf(t, batch, 0)["Sequence"])
	require.Equal(t, uint32(6), innerOf(t, batch, 1)["Sequence"])
	require.Len(t, querier.accounts, 1, "caller-set sequences are checked against the account's current sequence")
}

func TestBuildBatchTicketCreateAdvancesSequence(t *testing.T) {
	const (
		outerSequence = 10
		ticketCount   = 3
	)

	tests := []struct {
		name  string
		nonce TxOptions
		// wantNext is the sequence the AccountSet after the TicketCreate must spend.
		wantNext uint32
	}{
		{
			name:     "sequence-funded",
			wantNext: outerSequence + 1 + 1 + ticketCount,
		},
		{
			name:     "ticket-funded",
			nonce:    TxOptions{TicketSequence: 700},
			wantNext: outerSequence + 1 + ticketCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newBatchFixture(t)
			batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
				TxOptions: TxOptions{Sequence: outerSequence},
				Account:   testAccount,
				Operations: []BatchOperation{
					plainTicketCreate(testAccount, ticketCount, tt.nonce),
					plainAccountSet(testAccount, 0),
				},
			})
			require.NoError(t, err)
			require.Equal(t, tt.wantNext, innerOf(t, batch, 1)["Sequence"])
		})
	}
}

func TestBuildBatchRejectsInvalidNonces(t *testing.T) {
	tests := []struct {
		name       string
		outer      TxOptions
		operations func(f *batchFixture) []BatchOperation
		wantErr    error
	}{
		{
			name: "plain inner reuses the outer sequence",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					plainAccountSet(testAccount, 5),
					sendOp(f, testAccount, testDestination, testIssuanceID, 10),
				}
			},
			wantErr: ErrBatchInnerSequenceMismatch,
		},
		{
			name: "plain inner skips an allocated sequence",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					sendOp(f, testAccount, testDestination, testIssuanceID, 10),
					plainAccountSet(testAccount, 8),
				}
			},
			wantErr: ErrBatchInnerSequenceMismatch,
		},
		{
			name: "another account's caller-set sequences skip one",
			operations: func(_ *batchFixture) []BatchOperation {
				return []BatchOperation{
					plainAccountSet(testDestination, 5),
					plainAccountSet(testDestination, 7),
				}
			},
			wantErr: ErrBatchInnerSequenceMismatch,
		},
		{
			name: "another account's caller-set sequences repeat",
			operations: func(_ *batchFixture) []BatchOperation {
				return []BatchOperation{
					plainAccountSet(testDestination, 5),
					plainAccountSet(testDestination, 5),
				}
			},
			wantErr: ErrBatchInnerSequenceMismatch,
		},
		{
			name: "another account's caller-set sequence is not its current one",
			operations: func(_ *batchFixture) []BatchOperation {
				return []BatchOperation{
					plainAccountSet(testDestination, 9),
					plainAccountSet(testAccount, 0),
				}
			},
			wantErr: ErrBatchInnerSequenceMismatch,
		},
		{
			name: "two inners spend one Ticket",
			operations: func(f *batchFixture) []BatchOperation {
				first := sendOp(f, testAccount, testDestination, testIssuanceID, 10)
				first.TicketSequence = 900
				second := sendOp(f, testAccount, testDestination, testIssuanceID, 20)
				second.TicketSequence = 900
				return []BatchOperation{first, second}
			},
			wantErr: ErrBatchDuplicateNonce,
		},
		{
			name:  "an inner spends the outer Batch's Ticket",
			outer: TxOptions{TicketSequence: 900},
			operations: func(f *batchFixture) []BatchOperation {
				ticketed := sendOp(f, testAccount, testDestination, testIssuanceID, 10)
				ticketed.TicketSequence = 900
				return []BatchOperation{ticketed, sendOp(f, testAccount, testDestination, testIssuanceID, 20)}
			},
			wantErr: ErrBatchDuplicateNonce,
		},
		{
			name: "a Ticket repeats a sequence another inner spends",
			operations: func(f *batchFixture) []BatchOperation {
				ticketed := plainTicketCreate(testAccount, 1, TxOptions{TicketSequence: 6})
				return []BatchOperation{sendOp(f, testAccount, testDestination, testIssuanceID, 10), ticketed}
			},
			wantErr: ErrBatchDuplicateNonce,
		},
		{
			name: "an inner spends a sequence a TicketCreate turned into a Ticket",
			operations: func(_ *batchFixture) []BatchOperation {
				return []BatchOperation{
					plainTicketCreate(testAccount, 2, TxOptions{}),
					plainAccountSet(testAccount, 7),
				}
			},
			wantErr: ErrBatchInnerSequenceMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newBatchFixture(t).
				withIssuance(testIssuanceID).
				withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100)}).
				withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
				withSequence(testAccount, 5).
				withSequence(testDestination, 5)

			outer := tt.outer
			if outer.TicketSequence == 0 {
				outer.Sequence = 5
			}
			_, err := BuildBatch(fixture.querier(), BuildBatchParams{
				TxOptions:  outer,
				Account:    testAccount,
				Operations: tt.operations(fixture),
			})
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

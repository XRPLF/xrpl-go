package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type finalityTestResponse struct {
	result string
}

type finalityLookupStep struct {
	status TransactionStatus[finalityTestResponse]
	err    error
}

type finalityLedgerStep struct {
	index uint32
	err   error
}

func TestValidatePreliminaryResult(t *testing.T) {
	tests := []struct {
		name         string
		engineResult string
		wantFamily   EngineResultFamily
		wantMonitor  bool
	}{
		{name: "tes preliminary success", engineResult: "tesSUCCESS", wantFamily: EngineResultTES, wantMonitor: true},
		{name: "ter retryable", engineResult: "terQUEUED", wantFamily: EngineResultTER, wantMonitor: true},
		{name: "tec fee claiming", engineResult: "tecPATH_DRY", wantFamily: EngineResultTEC, wantMonitor: true},
		{name: "tef local failure", engineResult: "tefPAST_SEQ", wantFamily: EngineResultTEF},
		{name: "tem malformed", engineResult: "temBAD_AMOUNT", wantFamily: EngineResultTEM},
		{name: "tel is outside monitored policy", engineResult: "telINSUF_FEE_P", wantFamily: EngineResultUnknown},
		{name: "empty result", engineResult: "", wantFamily: EngineResultUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantFamily, ClassifyEngineResult(tt.engineResult))

			err := ValidatePreliminaryResult(tt.engineResult)
			if tt.wantMonitor {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, ErrPreliminaryResult)
			var preliminaryErr *PreliminaryResultError
			require.ErrorAs(t, err, &preliminaryErr)
			require.Equal(t, tt.engineResult, preliminaryErr.EngineResult)
		})
	}
}

func TestWaitForFinalityMatrix(t *testing.T) {
	transportTimeout := errors.New("transport timeout")
	lastLedger20 := uint32(20)
	success := &finalityTestResponse{result: "tesSUCCESS"}
	tecFailure := &finalityTestResponse{result: "tecPATH_DRY"}

	notFound := TransactionStatus[finalityTestResponse]{}
	unvalidated := TransactionStatus[finalityTestResponse]{
		Response: &finalityTestResponse{result: "provisional"},
		Found:    true,
	}
	validatedSuccessAt20 := TransactionStatus[finalityTestResponse]{
		Response:     success,
		Found:        true,
		Validated:    true,
		LedgerIndex:  20,
		EngineResult: "tesSUCCESS",
	}
	validatedTECAt20 := TransactionStatus[finalityTestResponse]{
		Response:     tecFailure,
		Found:        true,
		Validated:    true,
		LedgerIndex:  20,
		EngineResult: "tecPATH_DRY",
	}

	tests := []struct {
		name                 string
		lastLedgerSequence   *uint32
		maxAttempts          int
		lookupSteps          []finalityLookupStep
		ledgerSteps          []finalityLedgerStep
		wantResponse         *finalityTestResponse
		wantError            error
		wantValidatedFailure bool
		wantTransportCause   error
		wantLookupCalls      int
		wantLedgerCalls      int
	}{
		{
			name:               "validated success exactly at LastLedgerSequence",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        1,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
				{status: validatedSuccessAt20},
			},
			ledgerSteps:     []finalityLedgerStep{{index: 20}},
			wantResponse:    success,
			wantLookupCalls: 2,
			wantLedgerCalls: 1,
		},
		{
			name:               "fixed attempt budget does not override ledger finality",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        1,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
				{status: unvalidated},
				{status: notFound},
				{status: validatedSuccessAt20},
			},
			ledgerSteps: []finalityLedgerStep{
				{index: 17},
				{index: 18},
				{index: 20},
			},
			wantResponse:    success,
			wantLookupCalls: 4,
			wantLedgerCalls: 3,
		},
		{
			name:               "validation racing expiry is rechecked",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        1,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
				{status: validatedSuccessAt20},
			},
			ledgerSteps:     []finalityLedgerStep{{index: 21}},
			wantResponse:    success,
			wantLookupCalls: 2,
			wantLedgerCalls: 1,
		},
		{
			name:               "expiry only after LastLedgerSequence passes",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        1,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
				{status: notFound},
				{status: notFound},
			},
			ledgerSteps: []finalityLedgerStep{
				{index: 20},
				{index: 21},
			},
			wantError:       ErrTransactionExpired,
			wantLookupCalls: 3,
			wantLedgerCalls: 2,
		},
		{
			name:               "validated tec is authoritative failure",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        1,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
				{status: validatedTECAt20},
			},
			ledgerSteps:          []finalityLedgerStep{{index: 20}},
			wantResponse:         tecFailure,
			wantError:            ErrValidatedTransaction,
			wantValidatedFailure: true,
			wantLookupCalls:      2,
			wantLedgerCalls:      1,
		},
		{
			name:               "transient transaction transport error is retried",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        2,
			lookupSteps: []finalityLookupStep{
				{err: transportTimeout},
				{status: notFound},
				{status: validatedSuccessAt20},
			},
			ledgerSteps:     []finalityLedgerStep{{index: 20}},
			wantResponse:    success,
			wantLookupCalls: 3,
			wantLedgerCalls: 1,
		},
		{
			name:               "transient ledger transport error is retried",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        2,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
				{status: validatedSuccessAt20},
			},
			ledgerSteps: []finalityLedgerStep{
				{err: transportTimeout},
			},
			wantResponse:    success,
			wantLookupCalls: 2,
			wantLedgerCalls: 1,
		},
		{
			name:               "repeated transport errors remain transport outcome",
			lastLedgerSequence: &lastLedger20,
			maxAttempts:        2,
			lookupSteps: []finalityLookupStep{
				{err: transportTimeout},
				{err: transportTimeout},
			},
			wantError:          ErrFinalityTransport,
			wantTransportCause: transportTimeout,
			wantLookupCalls:    2,
		},
		{
			name:            "missing LastLedgerSequence uses bounded unknown fallback",
			maxAttempts:     2,
			lookupSteps:     []finalityLookupStep{{status: notFound}, {status: unvalidated}},
			wantError:       ErrFinalityNotDetermined,
			wantLookupCalls: 2,
		},
		{
			name:               "nonpositive attempt budget still performs one lookup",
			lastLedgerSequence: nil,
			maxAttempts:        0,
			lookupSteps:        []finalityLookupStep{{status: notFound}},
			wantError:          ErrFinalityNotDetermined,
			wantLookupCalls:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookupCalls := 0
			ledgerCalls := 0
			hooks := FinalityHooks[finalityTestResponse]{
				LookupTransaction: func(context.Context) (TransactionStatus[finalityTestResponse], error) {
					require.Less(t, lookupCalls, len(tt.lookupSteps), "unexpected transaction lookup")
					step := tt.lookupSteps[lookupCalls]
					lookupCalls++
					return step.status, step.err
				},
				GetValidatedLedger: func(context.Context) (uint32, error) {
					require.Less(t, ledgerCalls, len(tt.ledgerSteps), "unexpected validated ledger lookup")
					step := tt.ledgerSteps[ledgerCalls]
					ledgerCalls++
					return step.index, step.err
				},
			}

			response, err := WaitForFinality(
				context.Background(),
				FinalityConfig{
					LastLedgerSequence: tt.lastLedgerSequence,
					MaxAttempts:        tt.maxAttempts,
				},
				hooks,
			)

			require.Same(t, tt.wantResponse, response)
			if tt.wantError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantError)
			}
			if tt.wantValidatedFailure {
				var validatedErr *ValidatedTransactionError
				require.ErrorAs(t, err, &validatedErr)
				require.Equal(t, "tecPATH_DRY", validatedErr.EngineResult)
				require.Equal(t, uint32(20), validatedErr.LedgerIndex)
			}
			if tt.wantTransportCause != nil {
				require.ErrorIs(t, err, tt.wantTransportCause)
				var transportErr *FinalityTransportError
				require.ErrorAs(t, err, &transportErr)
				require.Equal(t, tt.maxAttempts, transportErr.Attempts)
			}
			require.Equal(t, tt.wantLookupCalls, lookupCalls)
			require.Equal(t, tt.wantLedgerCalls, ledgerCalls)
		})
	}
}

func TestWaitForFinalityReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	lastLedger := uint32(20)

	response, err := WaitForFinality(
		ctx,
		FinalityConfig{LastLedgerSequence: &lastLedger, MaxAttempts: 2},
		FinalityHooks[finalityTestResponse]{
			LookupTransaction: func(context.Context) (TransactionStatus[finalityTestResponse], error) {
				return TransactionStatus[finalityTestResponse]{}, nil
			},
			GetValidatedLedger: func(context.Context) (uint32, error) {
				cancel()
				return lastLedger, nil
			},
		},
	)

	require.Nil(t, response)
	require.ErrorIs(t, err, context.Canceled)
	require.NotErrorIs(t, err, ErrTransactionExpired)
}

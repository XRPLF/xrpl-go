package client

import (
	"context"
	"errors"
	"testing"
	"time"

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
	const resultMessage = "preliminary result message"
	tests := []struct {
		name         string
		engineResult string
		wantFamily   EngineResultFamily
		wantError    bool
	}{
		{name: "tes preliminary success", engineResult: "tesSUCCESS", wantFamily: EngineResultTES},
		{name: "ter retryable", engineResult: "terQUEUED", wantFamily: EngineResultTER},
		{name: "tec fee claiming", engineResult: "tecPATH_DRY", wantFamily: EngineResultTEC},
		{name: "tef failure", engineResult: "tefPAST_SEQ", wantFamily: EngineResultTEF},
		{name: "tel local error", engineResult: "telINSUF_FEE_P", wantFamily: EngineResultTEL},
		{name: "unknown result", engineResult: "customResult", wantFamily: EngineResultUnknown},
		{name: "empty result", engineResult: "", wantFamily: EngineResultUnknown},
		{name: "tem malformed", engineResult: "temBAD_AMOUNT", wantFamily: EngineResultTEM, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantFamily, ClassifyEngineResult(tt.engineResult))

			err := ValidatePreliminaryResult(tt.engineResult, resultMessage)
			if !tt.wantError {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, ErrPreliminaryResult)
			var preliminaryErr *PreliminaryResultError
			require.ErrorAs(t, err, &preliminaryErr)
			require.Equal(t, tt.engineResult, preliminaryErr.EngineResult)
			require.Equal(t, resultMessage, preliminaryErr.EngineResultMessage)
			require.ErrorContains(t, err, resultMessage)
		})
	}
}

func TestWaitForFinalityMatrix(t *testing.T) {
	transportTimeout := errors.New("transport timeout")
	const (
		lastLedger20      = uint32(20)
		preliminaryResult = "terQUEUED"
	)
	success := &finalityTestResponse{result: "tesSUCCESS"}
	tecResult := &finalityTestResponse{result: "tecPATH_DRY"}
	unknownResult := &finalityTestResponse{result: "customResult"}

	notFound := TransactionStatus[finalityTestResponse]{}
	unvalidated := TransactionStatus[finalityTestResponse]{
		Response: &finalityTestResponse{result: "provisional"},
		Found:    true,
	}
	validatedSuccess := TransactionStatus[finalityTestResponse]{Response: success, Found: true, Validated: true}
	validatedTEC := TransactionStatus[finalityTestResponse]{Response: tecResult, Found: true, Validated: true}
	validatedUnknown := TransactionStatus[finalityTestResponse]{Response: unknownResult, Found: true, Validated: true}

	tests := []struct {
		name               string
		maxAttempts        int
		lookupSteps        []finalityLookupStep
		ledgerSteps        []finalityLedgerStep
		wantResponse       *finalityTestResponse
		wantError          error
		wantTransportCause error
		wantLookupCalls    int
		wantLedgerCalls    int
		wantExpiryLedger   uint32
	}{
		{
			name:            "validated success exactly at LastLedgerSequence",
			maxAttempts:     1,
			lookupSteps:     []finalityLookupStep{{status: validatedSuccess}},
			ledgerSteps:     []finalityLedgerStep{{index: 20}},
			wantResponse:    success,
			wantLookupCalls: 1,
			wantLedgerCalls: 1,
		},
		{
			name:        "fixed attempt budget does not override ledger finality",
			maxAttempts: 1,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
				{status: unvalidated},
				{status: notFound},
				{status: validatedSuccess},
			},
			ledgerSteps: []finalityLedgerStep{
				{index: 17},
				{index: 18},
				{index: 20},
				{index: 20},
			},
			wantResponse:    success,
			wantLookupCalls: 4,
			wantLedgerCalls: 4,
		},
		{
			name:             "passed LastLedgerSequence expires before transaction lookup",
			maxAttempts:      1,
			ledgerSteps:      []finalityLedgerStep{{index: 21}},
			wantError:        ErrTransactionExpired,
			wantLookupCalls:  0,
			wantLedgerCalls:  1,
			wantExpiryLedger: 21,
		},
		{
			name:        "expiry only after LastLedgerSequence passes",
			maxAttempts: 1,
			lookupSteps: []finalityLookupStep{
				{status: notFound},
			},
			ledgerSteps: []finalityLedgerStep{
				{index: 20},
				{index: 21},
			},
			wantError:        ErrTransactionExpired,
			wantLookupCalls:  1,
			wantLedgerCalls:  2,
			wantExpiryLedger: 21,
		},
		{
			name:            "validated tec returns response without error",
			maxAttempts:     1,
			lookupSteps:     []finalityLookupStep{{status: validatedTEC}},
			ledgerSteps:     []finalityLedgerStep{{index: 20}},
			wantResponse:    tecResult,
			wantLookupCalls: 1,
			wantLedgerCalls: 1,
		},
		{
			name:            "unknown validated result returns response without error",
			maxAttempts:     1,
			lookupSteps:     []finalityLookupStep{{status: validatedUnknown}},
			ledgerSteps:     []finalityLedgerStep{{index: 20}},
			wantResponse:    unknownResult,
			wantLookupCalls: 1,
			wantLedgerCalls: 1,
		},
		{
			name:        "transient transaction transport error is retried",
			maxAttempts: 2,
			lookupSteps: []finalityLookupStep{
				{err: transportTimeout},
				{status: notFound},
				{status: validatedSuccess},
			},
			ledgerSteps: []finalityLedgerStep{
				{index: 19},
				{index: 20},
				{index: 20},
			},
			wantResponse:    success,
			wantLookupCalls: 3,
			wantLedgerCalls: 3,
		},
		{
			name:            "transient ledger transport error is retried",
			maxAttempts:     2,
			lookupSteps:     []finalityLookupStep{{status: validatedSuccess}},
			ledgerSteps:     []finalityLedgerStep{{err: transportTimeout}, {index: 20}},
			wantResponse:    success,
			wantLookupCalls: 1,
			wantLedgerCalls: 2,
		},
		{
			name:        "complete round resets incomplete round count",
			maxAttempts: 2,
			lookupSteps: []finalityLookupStep{
				{err: transportTimeout},
				{status: notFound},
				{err: transportTimeout},
				{err: transportTimeout},
			},
			ledgerSteps: []finalityLedgerStep{
				{index: 19},
				{index: 19},
				{index: 19},
				{index: 19},
			},
			wantError:          ErrFinalityTransport,
			wantTransportCause: transportTimeout,
			wantLookupCalls:    4,
			wantLedgerCalls:    4,
		},
		{
			name:        "repeated transport errors remain transport outcome",
			maxAttempts: 2,
			lookupSteps: []finalityLookupStep{
				{err: transportTimeout},
				{err: transportTimeout},
			},
			ledgerSteps: []finalityLedgerStep{
				{index: 19},
				{index: 19},
			},
			wantError:          ErrFinalityTransport,
			wantTransportCause: transportTimeout,
			wantLookupCalls:    2,
			wantLedgerCalls:    2,
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
					LastLedgerSequence: lastLedger20,
					PreliminaryResult:  preliminaryResult,
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
			if tt.wantTransportCause != nil {
				require.ErrorIs(t, err, tt.wantTransportCause)
				var transportErr *FinalityTransportError
				require.ErrorAs(t, err, &transportErr)
				require.Equal(t, tt.maxAttempts, transportErr.Attempts)
			}
			if tt.wantExpiryLedger != 0 {
				var expiryErr *TransactionExpiredError
				require.ErrorAs(t, err, &expiryErr)
				require.Equal(t, lastLedger20, expiryErr.LastLedgerSequence)
				require.Equal(t, tt.wantExpiryLedger, expiryErr.ValidatedLedger)
				require.Equal(t, preliminaryResult, expiryErr.PreliminaryResult)
			}
			require.Equal(t, tt.wantLookupCalls, lookupCalls)
			require.Equal(t, tt.wantLedgerCalls, ledgerCalls)
		})
	}
}

func TestWaitForFinalityReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	const lastLedger = uint32(20)

	response, err := WaitForFinality(
		ctx,
		FinalityConfig{LastLedgerSequence: lastLedger, MaxAttempts: 2},
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

func TestWaitForFinalityRejectsNegativePollInterval(t *testing.T) {
	lookupCalls := 0
	ledgerCalls := 0

	response, err := WaitForFinality(
		context.Background(),
		FinalityConfig{PollInterval: -time.Nanosecond},
		FinalityHooks[finalityTestResponse]{
			LookupTransaction: func(context.Context) (TransactionStatus[finalityTestResponse], error) {
				lookupCalls++
				return TransactionStatus[finalityTestResponse]{}, nil
			},
			GetValidatedLedger: func(context.Context) (uint32, error) {
				ledgerCalls++
				return 0, nil
			},
		},
	)

	require.Nil(t, response)
	require.ErrorIs(t, err, ErrInvalidPollInterval)
	var intervalErr *InvalidPollIntervalError
	require.ErrorAs(t, err, &intervalErr)
	require.Equal(t, -time.Nanosecond, intervalErr.PollInterval)
	require.Zero(t, lookupCalls)
	require.Zero(t, ledgerCalls)
}

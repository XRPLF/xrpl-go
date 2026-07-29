package client

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrPreliminaryResult indicates that reliable submission stopped on a
	// non-monitorable provisional engine result.
	ErrPreliminaryResult = errors.New("non-monitorable preliminary transaction result")
	// ErrValidatedTransaction indicates that a transaction was included in a
	// validated ledger with a non-success result.
	ErrValidatedTransaction = errors.New("transaction validated with a failure result")
	// ErrTransactionExpired indicates that the validated ledger passed the
	// transaction's LastLedgerSequence without a validated transaction result.
	ErrTransactionExpired = errors.New("transaction expired before validation")
	// ErrFinalityNotDetermined indicates that bounded monitoring of a transaction
	// without LastLedgerSequence ended without an authoritative result.
	ErrFinalityNotDetermined = errors.New("transaction finality was not determined")
	// ErrFinalityTransport indicates that repeated transport or response failures
	// prevented reliable-submission monitoring from making progress.
	ErrFinalityTransport = errors.New("transaction finality monitoring transport failure")
)

// EngineResultFamily identifies an XRPL transaction engine-result token family.
type EngineResultFamily string

const (
	// EngineResultTES is the provisional-success family.
	EngineResultTES EngineResultFamily = "tes"
	// EngineResultTER is the retryable family.
	EngineResultTER EngineResultFamily = "ter"
	// EngineResultTEC is the fee-claiming family.
	EngineResultTEC EngineResultFamily = "tec"
	// EngineResultTEF is the local failure family.
	EngineResultTEF EngineResultFamily = "tef"
	// EngineResultTEM is the malformed family.
	EngineResultTEM EngineResultFamily = "tem"
	// EngineResultUnknown is an unrecognized result family.
	EngineResultUnknown EngineResultFamily = ""
)

// engineResultValidatedSuccess is the exact engine result of a validated
// success. Validated success is an exact-token match, not a tes-family check.
const engineResultValidatedSuccess = "tesSUCCESS"

// PreliminaryResultError reports a provisional submit result that the SDK does
// not monitor. It does not claim that the transaction has a validated-ledger
// outcome.
type PreliminaryResultError struct {
	EngineResult string
}

// Error implements error.
func (e *PreliminaryResultError) Error() string {
	return fmt.Sprintf("transaction failed to submit with engine result: %s", e.EngineResult)
}

// Is supports errors.Is with ErrPreliminaryResult.
func (e *PreliminaryResultError) Is(target error) bool {
	return target == ErrPreliminaryResult
}

// ValidatedTransactionError reports an authoritative non-success result from a
// validated ledger. The caller also receives the transaction response.
type ValidatedTransactionError struct {
	EngineResult string
	LedgerIndex  uint32
}

// Error implements error.
func (e *ValidatedTransactionError) Error() string {
	return fmt.Sprintf("transaction validated with engine result %s in ledger %d", e.EngineResult, e.LedgerIndex)
}

// Is supports errors.Is with ErrValidatedTransaction.
func (e *ValidatedTransactionError) Is(target error) bool {
	return target == ErrValidatedTransaction
}

// TransactionExpiredError reports ledger-driven expiry after the validated
// ledger has advanced strictly beyond LastLedgerSequence.
type TransactionExpiredError struct {
	LastLedgerSequence uint32
	ValidatedLedger    uint32
}

// Error implements error.
func (e *TransactionExpiredError) Error() string {
	return fmt.Sprintf(
		"transaction expired: validated ledger %d passed LastLedgerSequence %d",
		e.ValidatedLedger,
		e.LastLedgerSequence,
	)
}

// Is supports errors.Is with ErrTransactionExpired.
func (e *TransactionExpiredError) Is(target error) bool {
	return target == ErrTransactionExpired
}

// FinalityNotDeterminedError reports exhaustion of the bounded fallback used
// when a transaction has no LastLedgerSequence.
type FinalityNotDeterminedError struct {
	Attempts int
}

// Error implements error.
func (e *FinalityNotDeterminedError) Error() string {
	return fmt.Sprintf("transaction finality was not determined after %d polling attempts", e.Attempts)
}

// Is supports errors.Is with ErrFinalityNotDetermined.
func (e *FinalityNotDeterminedError) Is(target error) bool {
	return target == ErrFinalityNotDetermined
}

// FinalityTransportError reports repeated query failures while monitoring. Err
// retains the transport-specific cause, including request timeout sentinels.
type FinalityTransportError struct {
	Operation string
	Attempts  int
	Err       error
}

// Error implements error.
func (e *FinalityTransportError) Error() string {
	return fmt.Sprintf("%s failed %d consecutive times while monitoring transaction finality: %v", e.Operation, e.Attempts, e.Err)
}

// Is supports errors.Is with ErrFinalityTransport.
func (e *FinalityTransportError) Is(target error) bool {
	return target == ErrFinalityTransport
}

// Unwrap retains the transport-specific failure for errors.Is/errors.As.
func (e *FinalityTransportError) Unwrap() error {
	return e.Err
}

// ClassifyEngineResult returns the textual family of an engine-result token.
// Token strings, rather than numeric codes, are stable protocol identifiers.
func ClassifyEngineResult(engineResult string) EngineResultFamily {
	if len(engineResult) < 3 {
		return EngineResultUnknown
	}

	switch family := EngineResultFamily(engineResult[:3]); family {
	case EngineResultTES, EngineResultTER, EngineResultTEC, EngineResultTEF, EngineResultTEM:
		return family
	case EngineResultUnknown:
		return EngineResultUnknown
	default:
		return EngineResultUnknown
	}
}

// ValidatePreliminaryResult accepts provisional families that require ledger
// monitoring and returns a typed error for fail-fast or unknown families.
func ValidatePreliminaryResult(engineResult string) error {
	switch ClassifyEngineResult(engineResult) {
	case EngineResultTES, EngineResultTER, EngineResultTEC:
		return nil
	case EngineResultTEF, EngineResultTEM, EngineResultUnknown:
		return &PreliminaryResultError{EngineResult: engineResult}
	default:
		return &PreliminaryResultError{EngineResult: engineResult}
	}
}

// TransactionStatus is the transport-neutral result of looking up a submitted
// transaction by hash.
type TransactionStatus[T any] struct {
	Response     *T
	Found        bool
	Validated    bool
	LedgerIndex  uint32
	EngineResult string
}

// FinalityConfig configures ledger-driven transaction monitoring.
type FinalityConfig struct {
	LastLedgerSequence *uint32
	PollInterval       time.Duration
	// MaxAttempts limits consecutive failed polling rounds. When
	// LastLedgerSequence is absent, it also limits successful inconclusive
	// polling rounds so monitoring cannot continue forever.
	MaxAttempts int
}

// FinalityHooks provide transport-specific transaction and validated-ledger
// queries to the shared state machine.
type FinalityHooks[T any] struct {
	LookupTransaction  func(context.Context) (TransactionStatus[T], error)
	GetValidatedLedger func(context.Context) (uint32, error)
}

// WaitForFinality monitors a transaction until it has an authoritative
// validated-ledger result, expires, is cancelled, or monitoring itself can no
// longer make bounded progress.
func WaitForFinality[T any](
	ctx context.Context,
	cfg FinalityConfig,
	hooks FinalityHooks[T],
) (*T, error) {
	maxAttempts := max(cfg.MaxAttempts, 1)

	transportFailures := 0
	inconclusiveAttempts := 0

	transportFailure := func(operation string, cause error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		transportFailures++
		if transportFailures >= maxAttempts {
			return &FinalityTransportError{
				Operation: operation,
				Attempts:  transportFailures,
				Err:       cause,
			}
		}
		return Wait(ctx, cfg.PollInterval)
	}
	authoritativeResult := func(status TransactionStatus[T]) (*T, error, bool) {
		if !status.Found || !status.Validated {
			return nil, nil, false
		}
		if status.Response == nil {
			return nil, &FinalityTransportError{
				Operation: "validated transaction response",
				Attempts:  1,
				Err:       errors.New("validated transaction response is nil"),
			}, true
		}
		if status.EngineResult == engineResultValidatedSuccess {
			return status.Response, nil, true
		}
		return status.Response, &ValidatedTransactionError{
			EngineResult: status.EngineResult,
			LedgerIndex:  status.LedgerIndex,
		}, true
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		status, err := hooks.LookupTransaction(ctx)
		if err != nil {
			if failureErr := transportFailure("transaction lookup", err); failureErr != nil {
				return nil, failureErr
			}
			continue
		}

		if response, resultErr, done := authoritativeResult(status); done {
			return response, resultErr
		}

		if cfg.LastLedgerSequence == nil {
			transportFailures = 0
			inconclusiveAttempts++
			if inconclusiveAttempts >= maxAttempts {
				return nil, &FinalityNotDeterminedError{Attempts: inconclusiveAttempts}
			}
			if err := Wait(ctx, cfg.PollInterval); err != nil {
				return nil, err
			}
			continue
		}

		validatedLedger, err := hooks.GetValidatedLedger(ctx)
		if err != nil {
			if failureErr := transportFailure("validated ledger lookup", err); failureErr != nil {
				return nil, failureErr
			}
			continue
		}
		transportFailures = 0

		if validatedLedger > *cfg.LastLedgerSequence {
			// The lookup that preceded the ledger query may have raced with
			// validation at LastLedgerSequence. Recheck after observing the
			// passed ledger before declaring expiry.
			status, err := hooks.LookupTransaction(ctx)
			if err != nil {
				if failureErr := transportFailure("transaction expiry recheck", err); failureErr != nil {
					return nil, failureErr
				}
				continue
			}
			if response, resultErr, done := authoritativeResult(status); done {
				return response, resultErr
			}
			return nil, &TransactionExpiredError{
				LastLedgerSequence: *cfg.LastLedgerSequence,
				ValidatedLedger:    validatedLedger,
			}
		}

		if err := Wait(ctx, cfg.PollInterval); err != nil {
			return nil, err
		}
	}
}

// Wait blocks until delay elapses or ctx is done, returning ctx.Err() on
// cancellation. A non-positive delay returns immediately after a context check.
func Wait(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

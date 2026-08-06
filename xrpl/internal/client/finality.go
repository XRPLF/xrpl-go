package client

import (
	"context"
	"errors"
	"fmt"
	"time"
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
	// EngineResultTEF is the failure family.
	EngineResultTEF EngineResultFamily = "tef"
	// EngineResultTEL is the local-error family.
	EngineResultTEL EngineResultFamily = "tel"
	// EngineResultTEM is the malformed family.
	EngineResultTEM EngineResultFamily = "tem"
	// EngineResultUnknown is an unrecognized result family.
	EngineResultUnknown EngineResultFamily = ""
)

// PreliminaryResultError reports a malformed preliminary submit result. It
// does not claim that the transaction has a validated-ledger outcome.
type PreliminaryResultError struct {
	EngineResult        string
	EngineResultMessage string
}

// Error implements error.
func (e *PreliminaryResultError) Error() string {
	return fmt.Sprintf(
		"transaction failed to submit with engine result %s: %s",
		e.EngineResult,
		e.EngineResultMessage,
	)
}

// Is supports errors.Is with ErrPreliminaryResult.
func (e *PreliminaryResultError) Is(target error) bool {
	return target == ErrPreliminaryResult
}

// TransactionExpiredError reports ledger-driven expiry after the validated
// ledger has advanced strictly beyond LastLedgerSequence. PreliminaryResult
// retains the engine result returned by the submit request.
type TransactionExpiredError struct {
	LastLedgerSequence uint32
	ValidatedLedger    uint32
	PreliminaryResult  string
}

// Error implements error.
func (e *TransactionExpiredError) Error() string {
	return fmt.Sprintf(
		"transaction expired: validated ledger %d passed LastLedgerSequence %d, preliminary result %s",
		e.ValidatedLedger,
		e.LastLedgerSequence,
		e.PreliminaryResult,
	)
}

// Is supports errors.Is with ErrTransactionExpired.
func (e *TransactionExpiredError) Is(target error) bool {
	return target == ErrTransactionExpired
}

// FinalityTransportError reports consecutive incomplete polling rounds. Err
// retains the transport-specific cause, including request timeout sentinels.
type FinalityTransportError struct {
	Operation string
	Attempts  int
	Err       error
}

// Error implements error.
func (e *FinalityTransportError) Error() string {
	return fmt.Sprintf(
		"transaction finality monitoring had %d consecutive incomplete rounds, last failed operation %s: %v",
		e.Attempts,
		e.Operation,
		e.Err,
	)
}

// Is supports errors.Is with ErrFinalityTransport.
func (e *FinalityTransportError) Is(target error) bool {
	return target == ErrFinalityTransport
}

// Unwrap retains the transport-specific failure for errors.Is and errors.As.
func (e *FinalityTransportError) Unwrap() error {
	return e.Err
}

// InvalidPollIntervalError reports a negative finality polling interval.
type InvalidPollIntervalError struct {
	PollInterval time.Duration
}

// Error implements error.
func (e *InvalidPollIntervalError) Error() string {
	return fmt.Sprintf("transaction finality poll interval must not be negative: %s", e.PollInterval)
}

// Is supports errors.Is with ErrInvalidPollInterval.
func (e *InvalidPollIntervalError) Is(target error) bool {
	return target == ErrInvalidPollInterval
}

// ValidatePollInterval rejects negative finality polling intervals. A zero
// interval remains valid for callers that intentionally request no delay.
func ValidatePollInterval(pollInterval time.Duration) error {
	if pollInterval < 0 {
		return &InvalidPollIntervalError{PollInterval: pollInterval}
	}
	return nil
}

// ClassifyEngineResult returns the textual family of an engine-result token.
// Token strings, rather than numeric codes, are stable protocol identifiers.
func ClassifyEngineResult(engineResult string) EngineResultFamily {
	if len(engineResult) < 3 {
		return EngineResultUnknown
	}

	switch family := EngineResultFamily(engineResult[:3]); family {
	case EngineResultTES, EngineResultTER, EngineResultTEC, EngineResultTEF, EngineResultTEL, EngineResultTEM:
		return family
	case EngineResultUnknown:
		return EngineResultUnknown
	default:
		return EngineResultUnknown
	}
}

// ValidatePreliminaryResult rejects malformed results and monitors all other
// preliminary results until a validated result or ledger expiry is available.
func ValidatePreliminaryResult(engineResult, engineResultMessage string) error {
	if ClassifyEngineResult(engineResult) != EngineResultTEM {
		return nil
	}
	return &PreliminaryResultError{
		EngineResult:        engineResult,
		EngineResultMessage: engineResultMessage,
	}
}

// TransactionStatus is the transport-neutral result of looking up a submitted
// transaction by hash.
type TransactionStatus[T any] struct {
	Response  *T
	Found     bool
	Validated bool
}

// FinalityConfig configures ledger-driven transaction monitoring.
type FinalityConfig struct {
	LastLedgerSequence uint32
	PreliminaryResult  string
	PollInterval       time.Duration
	// MaxAttempts limits consecutive incomplete polling rounds caused by query
	// or transport errors. It does not limit successful finality polling.
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
// longer make bounded progress. Each polling round follows the xrpl.js order:
// wait, read the latest validated ledger, check expiry, and then look up the
// transaction.
func WaitForFinality[T any](
	ctx context.Context,
	cfg FinalityConfig,
	hooks FinalityHooks[T],
) (*T, error) {
	if err := ValidatePollInterval(cfg.PollInterval); err != nil {
		return nil, err
	}

	maxAttempts := max(cfg.MaxAttempts, 1)
	incompleteRounds := 0

	incompleteRound := func(operation string, cause error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		incompleteRounds++
		if incompleteRounds >= maxAttempts {
			return &FinalityTransportError{
				Operation: operation,
				Attempts:  incompleteRounds,
				Err:       cause,
			}
		}
		return nil
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
		return status.Response, nil, true
	}

	for {
		if err := Wait(ctx, cfg.PollInterval); err != nil {
			return nil, err
		}

		validatedLedger, err := hooks.GetValidatedLedger(ctx)
		if err != nil {
			if failureErr := incompleteRound("validated ledger lookup", err); failureErr != nil {
				return nil, failureErr
			}
			continue
		}

		if validatedLedger > cfg.LastLedgerSequence {
			return nil, &TransactionExpiredError{
				LastLedgerSequence: cfg.LastLedgerSequence,
				ValidatedLedger:    validatedLedger,
				PreliminaryResult:  cfg.PreliminaryResult,
			}
		}

		status, err := hooks.LookupTransaction(ctx)
		if err != nil {
			if failureErr := incompleteRound("transaction lookup", err); failureErr != nil {
				return nil, failureErr
			}
			continue
		}

		if response, resultErr, done := authoritativeResult(status); done {
			return response, resultErr
		}

		incompleteRounds = 0
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

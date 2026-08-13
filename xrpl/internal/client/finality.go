package client

import (
	"context"
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

// ValidatePollInterval rejects negative finality polling intervals. A zero
// interval remains valid for callers that intentionally request no delay.
func ValidatePollInterval(pollInterval time.Duration) error {
	if pollInterval < 0 {
		return fmt.Errorf("%w: %s", ErrInvalidPollInterval, pollInterval)
	}
	return nil
}

// ValidateMaxRetries rejects non-positive finality retry limits.
func ValidateMaxRetries(maxRetries int) error {
	if maxRetries <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidMaxRetries, maxRetries)
	}
	return nil
}

// ValidateLastLedgerSequence rejects a zero finality ledger boundary.
func ValidateLastLedgerSequence(lastLedgerSequence uint32) error {
	if lastLedgerSequence == 0 {
		return ErrInvalidLastLedgerSequence
	}
	return nil
}

// ValidateFinalityMonitoring validates settings used before submission and by
// the finality state machine.
func ValidateFinalityMonitoring(pollInterval time.Duration, maxRetries int) error {
	if err := ValidatePollInterval(pollInterval); err != nil {
		return err
	}
	return ValidateMaxRetries(maxRetries)
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

// ValidatePreliminaryResult rejects missing or malformed results and monitors
// all other preliminary results until validation or ledger expiry.
func ValidatePreliminaryResult(engineResult, engineResultMessage string) error {
	if engineResult != "" && ClassifyEngineResult(engineResult) != EngineResultTEM {
		return nil
	}
	return fmt.Errorf(
		"%w: engine result %q: %s",
		ErrPreliminaryResult,
		engineResult,
		engineResultMessage,
	)
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
	// or transport errors. It does not limit successful finality polling. The
	// value must be positive.
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
// longer make bounded progress. Each polling round waits, reads the latest
// validated ledger, looks up the transaction, and then reports expiry when the
// final lookup remains inconclusive after LastLedgerSequence.
func WaitForFinality[T any](
	ctx context.Context,
	cfg FinalityConfig,
	hooks FinalityHooks[T],
) (*T, error) {
	if err := ValidateFinalityMonitoring(cfg.PollInterval, cfg.MaxAttempts); err != nil {
		return nil, err
	}
	if err := ValidateLastLedgerSequence(cfg.LastLedgerSequence); err != nil {
		return nil, err
	}

	maxAttempts := cfg.MaxAttempts
	incompleteRounds := 0

	incompleteRound := func(operation string, cause error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		incompleteRounds++
		if incompleteRounds >= maxAttempts {
			return fmt.Errorf(
				"%w after %d consecutive incomplete rounds, last failed operation %s: %w",
				ErrFinalityTransport,
				incompleteRounds,
				operation,
				cause,
			)
		}
		return nil
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

		status, err := hooks.LookupTransaction(ctx)
		if err != nil {
			if failureErr := incompleteRound("transaction lookup", err); failureErr != nil {
				return nil, failureErr
			}
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if status.Found && status.Validated {
			if status.Response == nil {
				if failureErr := incompleteRound(
					"validated transaction response",
					errNilValidatedTransactionResponse,
				); failureErr != nil {
					return nil, failureErr
				}
				continue
			}
			return status.Response, nil
		}

		if validatedLedger > cfg.LastLedgerSequence {
			return nil, fmt.Errorf(
				"%w: validated ledger %d passed LastLedgerSequence %d, preliminary result %s",
				ErrTransactionExpired,
				validatedLedger,
				cfg.LastLedgerSequence,
				cfg.PreliminaryResult,
			)
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

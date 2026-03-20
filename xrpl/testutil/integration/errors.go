package integration

import "errors"

// ErrFailedToFundWallet is returned when funding a wallet fails after exceeding retry limit.
var ErrFailedToFundWallet = errors.New("failed to fund wallet")

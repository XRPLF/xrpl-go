package amm

import "errors"

// ErrInvalidInfoRequest is returned when neither amm_account nor both asset and asset2 are specified.
var ErrInvalidInfoRequest = errors.New("amm_info: must specify either amm_account or both asset and asset2")

package amm

import "errors"

// ErrInvalidInfoRequest is returned unless exactly one AMM lookup is specified:
// either amm_account, or both asset and asset2.
var ErrInvalidInfoRequest = errors.New("amm_info: must specify exactly one of amm_account or both asset and asset2")

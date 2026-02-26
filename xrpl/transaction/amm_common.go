package transaction

// Common flags for AMM transactions (Deposit and Withdraw).
const (
	// TfLPToken performs a double-asset withdrawal/deposit and receive the specified amount of LP Tokens.
	TfLPToken uint32 = 65536
	// TfSingleAsset performs a single-asset withdrawal/deposit with a specified amount of the asset to deposit.
	TfSingleAsset uint32 = 524288
	// TfTwoAsset performs a double-asset withdrawal/deposit with specified amounts of both assets.
	TfTwoAsset uint32 = 1048576
	// TfOneAssetLPToken performs a single-asset withdrawal/deposit and receive the specified amount of LP Tokens.
	TfOneAssetLPToken uint32 = 2097152
	// TfLimitLPToken performs a single-asset withdrawal/deposit with a specified effective price.
	TfLimitLPToken uint32 = 4194304

	// AmmMaxTradingFee is the maximum trading fee; a value of 1000 corresponds to a 1% fee.
	AmmMaxTradingFee = 1000
)

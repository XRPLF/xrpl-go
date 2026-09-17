package proof

// Participant represents a participant with hex-encoded fields.
type Participant struct {
	PubKeyHex     string // 66 hex chars (33 bytes)
	CiphertextHex string // 132 hex chars (66 bytes)
}

// Params holds hex-encoded Pedersen linkage proof parameters.
type Params struct {
	CommitmentHex     string // 66 hex chars (33 bytes)
	Amount            uint64
	CiphertextHex     string // 132 hex chars (66 bytes)
	BlindingFactorHex string // 64 hex chars (32 bytes)
}

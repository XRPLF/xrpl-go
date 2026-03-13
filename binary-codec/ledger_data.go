package binarycodec

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// LedgerData represents the data of a ledger.
type LedgerData struct {
	LedgerIndex         uint32
	TotalCoins          string
	ParentHash          string
	TransactionHash     string
	AccountHash         string
	ParentCloseTime     uint32
	CloseTime           uint32
	CloseTimeResolution uint8
	CloseFlags          uint8
}

// DecodeLedgerData decodes a hex string in the canonical binary format into a LedgerData object.
// The hex string should represent a ledger data object.
func DecodeLedgerData(data string) (LedgerData, error) {
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return LedgerData{}, err
	}

	parser := serdes.NewBinaryParser(decoded, definitions.Get())
	var ledgerData LedgerData

	ledgerIndex, err := parser.ReadBytes(4)
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.LedgerIndex = binary.BigEndian.Uint32(ledgerIndex)

	totalCoins, err := parser.ReadBytes(8)
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.TotalCoins = strconv.FormatUint(binary.BigEndian.Uint64(totalCoins), 10)

	parentHash, err := parser.ReadBytes(32)
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.ParentHash = hexutil.EncodeToUpperHex(parentHash)

	transactionHash, err := parser.ReadBytes(32)
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.TransactionHash = hexutil.EncodeToUpperHex(transactionHash)

	accountHash, err := parser.ReadBytes(32)
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.AccountHash = hexutil.EncodeToUpperHex(accountHash)

	parentCloseTime, err := parser.ReadBytes(4)
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.ParentCloseTime = binary.BigEndian.Uint32(parentCloseTime)

	closeTime, err := parser.ReadBytes(4)
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.CloseTime = binary.BigEndian.Uint32(closeTime)

	closeTimeResolution, err := parser.ReadByte()
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.CloseTimeResolution = uint8(closeTimeResolution)

	closeFlags, err := parser.ReadByte()
	if err != nil {
		return LedgerData{}, err
	}

	ledgerData.CloseFlags = uint8(closeFlags)

	return ledgerData, nil
}

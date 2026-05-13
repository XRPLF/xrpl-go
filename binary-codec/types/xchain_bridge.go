//revive:disable:var-naming
package types

import (
	"errors"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// Errors
var (
	errNotValidXChainBridge = errors.New("not a valid xchain bridge")
)

const xChainBridgeLength = 80

// XChainBridge is a struct that represents an xchain bridge.
type XChainBridge struct{}

// FromJSON converts a json XChainBridge object to its byte slice representation.
// It returns an error if the json is not valid or if the classic addresses are not valid.
func (x *XChainBridge) FromJSON(json any) ([]byte, error) {
	v, ok := json.(map[string]any)
	if !ok {
		return nil, errNotValidJSON
	}

	lockingChainDoorStr, ok := v["LockingChainDoor"].(string)
	if !ok {
		return nil, errNotValidXChainBridge
	}

	lockingChainIssueStr, ok := v["LockingChainIssue"].(string)
	if !ok {
		return nil, errNotValidXChainBridge
	}

	issuingChainDoorStr, ok := v["IssuingChainDoor"].(string)
	if !ok {
		return nil, errNotValidXChainBridge
	}

	issuingChainIssueStr, ok := v["IssuingChainIssue"].(string)
	if !ok {
		return nil, errNotValidXChainBridge
	}

	_, lockingChainDoor, err := addresscodec.DecodeClassicAddressToAccountID(lockingChainDoorStr)
	if err != nil {
		return nil, errDecodeClassicAddress
	}

	_, lockingChainIssue, err := addresscodec.DecodeClassicAddressToAccountID(lockingChainIssueStr)
	if err != nil {
		return nil, errDecodeClassicAddress
	}

	_, issuingChainDoor, err := addresscodec.DecodeClassicAddressToAccountID(issuingChainDoorStr)
	if err != nil {
		return nil, errDecodeClassicAddress
	}

	_, issuingChainIssue, err := addresscodec.DecodeClassicAddressToAccountID(issuingChainIssueStr)
	if err != nil {
		return nil, errDecodeClassicAddress
	}

	bytes := make([]byte, 0, 80)

	bytes = append(bytes, lockingChainDoor...)
	bytes = append(bytes, lockingChainIssue...)
	bytes = append(bytes, issuingChainDoor...)
	bytes = append(bytes, issuingChainIssue...)

	return bytes, nil
}

// ToJSON converts a byte slice representation of an XChainBridge object to its json representation.
// It returns an error if the bytes are not valid or if the classic addresses are not valid.
func (x *XChainBridge) ToJSON(p interfaces.BinaryParser, opts ...int) (any, error) {
	if len(opts) == 0 {
		return nil, ErrNoLengthPrefix
	}

	bytes, err := p.ReadBytes(opts[0])
	if err != nil {
		return nil, errReadBytes
	}

	if len(bytes) != xChainBridgeLength {
		return nil, errNotValidXChainBridge
	}

	json := make(map[string]string)

	json["LockingChainDoor"], err = addresscodec.Encode(bytes[:20], []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
	if err != nil {
		return nil, err
	}
	json["LockingChainIssue"], err = addresscodec.Encode(bytes[20:40], []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
	if err != nil {
		return nil, err
	}
	json["IssuingChainDoor"], err = addresscodec.Encode(bytes[40:60], []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
	if err != nil {
		return nil, err
	}
	json["IssuingChainIssue"], err = addresscodec.Encode(bytes[60:80], []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
	if err != nil {
		return nil, err
	}

	return json, nil
}

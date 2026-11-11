//revive:disable:var-naming
package types

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// xChainBridgeFields lists the bridge fields in serialization order.
// As in rippled's STXChainBridge, a door is a length-prefixed AccountID and an issue is an Issue.
var xChainBridgeFields = [...]struct {
	name string
	door bool
}{
	{name: "LockingChainDoor", door: true},
	{name: "LockingChainIssue"},
	{name: "IssuingChainDoor", door: true},
	{name: "IssuingChainIssue"},
}

// XChainBridge is a struct that represents an xchain bridge.
type XChainBridge struct{}

// FromJSON converts a json XChainBridge object to its byte slice representation.
// The object must have exactly the fields LockingChainDoor, LockingChainIssue,
// IssuingChainDoor and IssuingChainIssue. A door is an address and an issue is an Issue object.
func (x *XChainBridge) FromJSON(json any) ([]byte, error) {
	v, ok := json.(map[string]any)
	if !ok {
		return nil, errNotValidJSON
	}
	if len(v) != len(xChainBridgeFields) {
		return nil, errNotValidXChainBridge
	}

	var sink []byte
	for _, field := range xChainBridgeFields {
		value, ok := v[field.name]
		if !ok {
			return nil, fmt.Errorf("%w: missing %s", errNotValidXChainBridge, field.name)
		}

		var b []byte
		var err error
		if field.door {
			sink = append(sink, addresscodec.AccountAddressLength)
			b, err = (&AccountID{}).FromJSON(value)
		} else {
			b, err = (&Issue{}).FromJSON(value)
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", errNotValidXChainBridge, field.name, err)
		}
		sink = append(sink, b...)
	}

	return sink, nil
}

// ToJSON reads an XChainBridge from the parser and returns its json representation.
// The value has no length prefix of its own, so the opts parameter is ignored.
func (x *XChainBridge) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	json := make(map[string]any, len(xChainBridgeFields))
	for _, field := range xChainBridgeFields {
		value, err := readXChainBridgeField(p, field.door)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", errNotValidXChainBridge, field.name, err)
		}
		json[field.name] = value
	}

	return json, nil
}

func readXChainBridgeField(p interfaces.BinaryParser, door bool) (any, error) {
	if !door {
		return (&Issue{}).ToJSON(p)
	}
	length, err := p.ReadVariableLength()
	if err != nil {
		return nil, err
	}
	return (&AccountID{}).ToJSON(p, length)
}

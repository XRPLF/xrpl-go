// Package definitions contains XRPL binary codec field and type definitions.
package definitions

import (
	_ "embed"

	"github.com/ugorji/go/codec"
)

var (
	//go:embed definitions.json
	docBytes []byte

	// definitions is the singleton instance of the Definitions struct.
	definitions *Definitions
)

// Definitions holds the binary serialization definitions for the XRP Ledger, loaded from the RFC JSON document.
type Definitions struct {
	Types                  map[string]int32
	LedgerEntryTypes       map[string]int32
	Fields                 fieldInstanceMap
	TransactionResults     map[string]int32
	TransactionTypes       map[string]int32
	FieldIDNameMap         map[FieldHeader]string
	GranularPermissions    map[string]int32
	DelegatablePermissions map[string]int32
}

// Get returns the singleton instance of Definitions.
func Get() *Definitions {
	return definitions
}

type definitionsDoc struct {
	Types              map[string]int32 `json:"TYPES"`
	LedgerEntryTypes   map[string]int32 `json:"LEDGER_ENTRY_TYPES"`
	Fields             fieldInstanceMap `json:"FIELDS"`
	TransactionResults map[string]int32 `json:"TRANSACTION_RESULTS"`
	TransactionTypes   map[string]int32 `json:"TRANSACTION_TYPES"`
}

// Loads JSON from the definitions file and converts it to a preferred format.
// The definitions file contains information required for the XRP Ledger's
// canonical binary serialization format:
// `Serialization <https://xrpl.org/serialization.html>`_
func loadDefinitions() {

	var jh codec.JsonHandle

	jh.MapKeyAsString = true
	jh.SignedInteger = true

	dec := codec.NewDecoderBytes(docBytes, &jh)
	var data definitionsDoc
	dec.MustDecode(&data)

	definitions = &Definitions{
		Types:              data.Types,
		Fields:             data.Fields,
		LedgerEntryTypes:   data.LedgerEntryTypes,
		TransactionResults: data.TransactionResults,
		TransactionTypes:   data.TransactionTypes,
	}

	addFieldHeadersAndOrdinals()
	createFieldIDNameMap()
	initializePermissions()
}

func convertToFieldInstanceMap(m [][]interface{}) map[string]*FieldInstance {
	nm := make(map[string]*FieldInstance, len(m))

	for _, j := range m {
		k := j[0].(string)
		fi, _ := castFieldInfo(j[1])
		nm[k] = &FieldInstance{
			FieldName: k,
			FieldInfo: &fi,
			Ordinal:   fi.Nth,
		}
	}
	return nm
}

func castFieldInfo(v interface{}) (FieldInfo, error) {
	if fi, ok := v.(map[string]interface{}); ok {
		return FieldInfo{
			// TODO: Check if this is still needed
			//nolint:gosec // G115: integer overflow conversion int64 -> int32, nth is a small field ordinal
			Nth:            int32(fi["nth"].(int64)),
			IsVLEncoded:    fi["isVLEncoded"].(bool),
			IsSerialized:   fi["isSerialized"].(bool),
			IsSigningField: fi["isSigningField"].(bool),
			Type:           fi["type"].(string),
		}, nil
	}
	return FieldInfo{}, ErrUnableToCastFieldInfo
}

func addFieldHeadersAndOrdinals() {
	for k := range definitions.Fields {
		t, _ := definitions.GetTypeCodeByTypeName(definitions.Fields[k].Type)

		if fi, ok := definitions.Fields[k]; ok {
			fi.FieldHeader = &FieldHeader{
				TypeCode:  t,
				FieldCode: definitions.Fields[k].Nth,
			}
			fi.Ordinal = (t<<16 | definitions.Fields[k].Nth)
		}
	}
}

func createFieldIDNameMap() {
	definitions.FieldIDNameMap = make(map[FieldHeader]string, len(definitions.Fields))
	for k := range definitions.Fields {
		fh, _ := definitions.GetFieldHeaderByFieldName(k)

		definitions.FieldIDNameMap[*fh] = k
	}
}

// Initializes granular permissions and delegatable permissions mappings for account permission delegation.
func initializePermissions() {
	definitions.GranularPermissions = map[string]int32{
		"TrustlineAuthorize":     65537,
		"TrustlineFreeze":        65538,
		"TrustlineUnfreeze":      65539,
		"AccountDomainSet":       65540,
		"AccountEmailHashSet":    65541,
		"AccountMessageKeySet":   65542,
		"AccountTransferRateSet": 65543,
		"AccountTickSizeSet":     65544,
		"PaymentMint":            65545,
		"PaymentBurn":            65546,
		"MPTokenIssuanceLock":    65547,
		"MPTokenIssuanceUnlock":  65548,
	}

	definitions.DelegatablePermissions = make(map[string]int32)

	for name, value := range definitions.GranularPermissions {
		definitions.DelegatablePermissions[name] = value
	}

	for txType, value := range definitions.TransactionTypes {
		definitions.DelegatablePermissions[txType] = value + 1
	}
}

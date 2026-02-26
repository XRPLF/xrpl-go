package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestOfferDirectoryNode(t *testing.T) {
	var s Object = &DirectoryNode{
		Flags: 0,
		Indexes: []types.Hash256{
			"AD7EAE148287EF12D213A251015F86E6D4BD34B3C4A0A1ED9A17198373F908AD",
		},
		LedgerEntryType:   DirectoryNodeEntry,
		RootIndex:         "1BBEF97EDE88D40CEE2ADE6FEF121166AFE80D99EBADB01A4F069BA8FF484000",
		TakerGetsCurrency: "0000000000000000000000000000000000000000",
		TakerGetsIssuer:   "0000000000000000000000000000000000000000",
		TakerPaysCurrency: "0000000000000000000000004A50590000000000",
		TakerPaysIssuer:   "5BBC0F22F61D9224A110650CFE21CC0C4BE13098",
		DomainID:          "5BBC0F22F61D9224A110650CFE21CC0C4BE13098",
	}

	j := `{
	"Flags": 0,
	"Indexes": [
		"AD7EAE148287EF12D213A251015F86E6D4BD34B3C4A0A1ED9A17198373F908AD"
	],
	"LedgerEntryType": "DirectoryNode",
	"RootIndex": "1BBEF97EDE88D40CEE2ADE6FEF121166AFE80D99EBADB01A4F069BA8FF484000",
	"TakerGetsCurrency": "0000000000000000000000000000000000000000",
	"TakerGetsIssuer": "0000000000000000000000000000000000000000",
	"TakerPaysCurrency": "0000000000000000000000004A50590000000000",
	"TakerPaysIssuer": "5BBC0F22F61D9224A110650CFE21CC0C4BE13098",
	"DomainID": "5BBC0F22F61D9224A110650CFE21CC0C4BE13098"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestOwnerDirectoryNode(t *testing.T) {
	var s Object = &DirectoryNode{
		Flags: 0,
		Indexes: []types.Hash256{
			"AD7EAE148287EF12D213A251015F86E6D4BD34B3C4A0A1ED9A17198373F908AD",
			"E83BBB58949A8303DF07172B16FB8EFBA66B9191F3836EC27A4568ED5997BAC5",
		},
		LedgerEntryType: DirectoryNodeEntry,
		Owner:           "rpR95n1iFkTqpoy1e878f4Z1pVHVtWKMNQ",
		RootIndex:       "193C591BF62482468422313F9D3274B5927CA80B4DD3707E42015DD609E39C94",
	}

	j := `{
	"Flags": 0,
	"Indexes": [
		"AD7EAE148287EF12D213A251015F86E6D4BD34B3C4A0A1ED9A17198373F908AD",
		"E83BBB58949A8303DF07172B16FB8EFBA66B9191F3836EC27A4568ED5997BAC5"
	],
	"LedgerEntryType": "DirectoryNode",
	"Owner": "rpR95n1iFkTqpoy1e878f4Z1pVHVtWKMNQ",
	"RootIndex": "193C591BF62482468422313F9D3274B5927CA80B4DD3707E42015DD609E39C94"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNFTTokenBuyOffersDirectoryNode(t *testing.T) {
	var s Object = &DirectoryNode{
		Flags: LsfNFTokenBuyOffers,
		Indexes: []types.Hash256{
			"68227B203065DED9EEB8B73FC952494A1DA6A69CEABEAA99923836EB5E77C95A",
		},
		LedgerEntryType:   DirectoryNodeEntry,
		NFTokenID:         "000822603EA060FD1026C04B2D390CC132D07D600DA9B082CB5CE9AC0487E50B",
		PreviousTxnID:     "EF8A9AD51E7CC6BBD219C3C980EC3145C7B0814ED3184471FD952D9D23A1918D",
		PreviousTxnLgrSeq: 91448417,
		RootIndex:         "0EC5802BD1AB56527A9DE524CCA2A2BA25E1085CCE7EA112940ED115FFF91EE2",
	}

	j := `{
	"Flags": 1,
	"Indexes": [
		"68227B203065DED9EEB8B73FC952494A1DA6A69CEABEAA99923836EB5E77C95A"
	],
	"LedgerEntryType": "DirectoryNode",
	"NFTokenID": "000822603EA060FD1026C04B2D390CC132D07D600DA9B082CB5CE9AC0487E50B",
	"PreviousTxnID": "EF8A9AD51E7CC6BBD219C3C980EC3145C7B0814ED3184471FD952D9D23A1918D",
	"PreviousTxnLgrSeq": 91448417,
	"RootIndex": "0EC5802BD1AB56527A9DE524CCA2A2BA25E1085CCE7EA112940ED115FFF91EE2"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNFTTokenSellOffersDirectoryNode(t *testing.T) {
	var s Object = &DirectoryNode{
		Flags: LsfNFTokenSellOffers,
		Indexes: []types.Hash256{
			"68227B203065DED9EEB8B73FC952494A1DA6A69CEABEAA99923836EB5E77C95A",
		},
		LedgerEntryType:   DirectoryNodeEntry,
		NFTokenID:         "000822603EA060FD1026C04B2D390CC132D07D600DA9B082CB5CE9AC0487E50B",
		PreviousTxnID:     "EF8A9AD51E7CC6BBD219C3C980EC3145C7B0814ED3184471FD952D9D23A1918D",
		PreviousTxnLgrSeq: 91448417,
		RootIndex:         "0EC5802BD1AB56527A9DE524CCA2A2BA25E1085CCE7EA112940ED115FFF91EE2",
	}

	j := `{
	"Flags": 2,
	"Indexes": [
		"68227B203065DED9EEB8B73FC952494A1DA6A69CEABEAA99923836EB5E77C95A"
	],
	"LedgerEntryType": "DirectoryNode",
	"NFTokenID": "000822603EA060FD1026C04B2D390CC132D07D600DA9B082CB5CE9AC0487E50B",
	"PreviousTxnID": "EF8A9AD51E7CC6BBD219C3C980EC3145C7B0814ED3184471FD952D9D23A1918D",
	"PreviousTxnLgrSeq": 91448417,
	"RootIndex": "0EC5802BD1AB56527A9DE524CCA2A2BA25E1085CCE7EA112940ED115FFF91EE2"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestDirectoryNode_LedgerEntryType(t *testing.T) {
	var s Object = &DirectoryNode{}
	require.Equal(t, s.EntryType(), DirectoryNodeEntry)
}

func TestDirectoryNode_SetNFTokenBuyOffers(t *testing.T) {
	s := &DirectoryNode{}
	s.SetNFTokenBuyOffers()
	require.Equal(t, s.Flags, LsfNFTokenBuyOffers)
}

func TestDirectoryNode_SetNFTokenSellOffers(t *testing.T) {
	s := &DirectoryNode{}
	s.SetNFTokenSellOffers()
	require.Equal(t, s.Flags, LsfNFTokenSellOffers)
}

package mpt

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

func testIntegrationMptTokenCreationMetadata() (string, error) {
	assetSubclass := "treasury"
	desc := "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments."
	metadata := types.ParsedMPTokenMetadata{
		Ticker:        "TBILL",
		Name:          "T-Bill Yield Token",
		Desc:          &desc,
		Icon:          "example.org/tbill-icon.png",
		AssetClass:    "rwa",
		AssetSubclass: &assetSubclass,
		IssuerName:    "Example Yield Co.",
		URIs: []types.ParsedMPTokenMetadataURI{
			{
				URI:      "exampleyield.co/tbill",
				Category: "website",
				Title:    "Product Page",
			},
			{
				URI:      "exampleyield.co/docs",
				Category: "docs",
				Title:    "Yield Token Docs",
			},
		},
		AdditionalInfo: map[string]any{
			"interest_rate": "5.00%",
			"interest_type": "variable",
			"yield_source":  "U.S. Treasury Bills",
			"maturity_date": "2045-06-30",
			"cusip":         "912796RX0",
		},
	}

	return types.EncodeMPTokenMetadata(metadata)
}

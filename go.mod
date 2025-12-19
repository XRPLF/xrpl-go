module github.com/Peersyst/xrpl-go

go 1.24.3

require (
	github.com/bsv-blockchain/go-sdk v1.2.9
	github.com/btcsuite/btcd/btcec/v2 v2.3.4
	github.com/decred/dcrd/crypto/ripemd160 v1.0.2
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/stretchr/testify v1.10.0
	github.com/ugorji/go/codec v1.2.11
)

require github.com/pkg/errors v0.9.1 // indirect

require (
	github.com/golang/mock v1.6.0 // direct
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace golang.org/x/crypto => golang.org/x/crypto v0.45.0

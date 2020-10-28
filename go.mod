module github.com/polynetwork/fabric-relayer

go 1.14

require (
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/cmars/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/ethereum/go-ethereum v1.9.15
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta3.0.20201006151309-9c426dcc5096
	github.com/ontio/ontology-crypto v1.0.9
	github.com/pkg/errors v0.9.1
	github.com/polynetwork/eth-contracts v0.0.0-20200814062128-70f58e22b014
	github.com/polynetwork/poly v0.0.0-20200722075529-eea88acb37b2
	github.com/polynetwork/poly-go-sdk v0.0.0-20200729103825-af447ef53ef0
	github.com/urfave/cli v1.22.4
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/go-kit/kit v0.10.0 => github.com/go-kit/kit v0.8.0
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
)

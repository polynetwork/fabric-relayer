module github.com/polynetwork/fabric-relayer

go 1.14

require (
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/Shopify/sarama v1.27.2 // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/cloudflare/cfssl v1.4.1
	github.com/cmars/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/ethereum/go-ethereum v1.9.15 // indirect
	github.com/fsouza/go-dockerclient v1.6.6 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/golang/protobuf v1.4.1
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/hyperledger/fabric v1.4.3 // indirect
	github.com/hyperledger/fabric-amcl v0.0.0-20200424173818-327c9e2cf77a // indirect
	github.com/hyperledger/fabric-lib-go v1.0.0
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta3.0.20201006151309-9c426dcc5096
	github.com/miekg/pkcs11 v1.0.3
	github.com/mitchellh/mapstructure v1.3.2
	github.com/ontio/ontology-crypto v1.0.9 // indirect
	github.com/pkg/errors v0.9.1
	github.com/polynetwork/eth-contracts v0.0.0-20200814062128-70f58e22b014 // indirect
	github.com/polynetwork/poly v0.0.0-20200722075529-eea88acb37b2
	github.com/polynetwork/poly-go-sdk v0.0.0-20200729103825-af447ef53ef0
	github.com/prometheus/client_golang v1.5.0
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.3.0
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/go-kit/kit v0.10.0 => github.com/go-kit/kit v0.8.0
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
)

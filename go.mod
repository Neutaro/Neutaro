module github.com/Neutaro/Neutaro-Chain

go 1.16

require (
	github.com/CosmWasm/wasmd v0.28.0
	github.com/cosmos/cosmos-sdk v0.45.6
	github.com/cosmos/ibc-go/v3 v3.0.1
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/ignite-hq/cli v0.22.1
	github.com/prometheus/client_golang v1.12.2
	github.com/spf13/cast v1.5.0
	github.com/stretchr/testify v1.8.1
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.7
	google.golang.org/genproto v0.0.0-20230223222841-637eb2293923 // indirect
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)

package interchaintest

import (
	"context"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"testing"
)

// Pretty much a direct copy from Juno: https://github.com/CosmosContracts/juno/blob/main/interchaintest/setup.go

var (
	VotingPeriod     = "15s"
	MaxDepositPeriod = "10s"
	Denom            = "uneutaro"

	IBCRelayerImage   = "ghcr.io/cosmos/relayer"
	IBCRelayerVersion = "main"

	// Just using local here (build with docker build -t neutaro:local .)
	NeutaroImage = ibc.DockerImage{
		Repository: "neutaro",
		Version:    "local",
		UidGid:     "1025:1025",
	}

	// SDK v45 Genesis
	defaultGenesisKV = []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.voting_params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.deposit_params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.deposit_params.min_deposit.0.denom",
			Value: Denom,
		},
		{
			Key:   "app_state.gov.deposit_params.min_deposit.0.amount",
			Value: "1",
		},
	}

	neutaroConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "neutaro",
		ChainID:             "neutaro-1",
		Images:              []ibc.DockerImage{NeutaroImage},
		Bin:                 "Neutaro",
		Bech32Prefix:        "neutaro",
		Denom:               Denom,
		CoinType:            "118",
		GasPrices:           fmt.Sprintf("0%s", Denom),
		GasAdjustment:       2.0,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ConfigFileOverrides: nil,
		EncodingConfig:      neutaroEncoding(),
		ModifyGenesis:       cosmos.ModifyGenesis(defaultGenesisKV),
	}

	genesisWalletAmount = int64(10_000_000)
)

// neutaroEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func neutaroEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasm.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

// CreateChain generates a new chain with a custom image (useful for upgrades)
func CreateChain(t *testing.T, numVals, numFull int, img ibc.DockerImage) []ibc.Chain {
	cfg := neutaroConfig
	cfg.Images = []ibc.DockerImage{img}
	return CreateChainWithCustomConfig(t, numVals, numFull, cfg)
}

// CreateThisBranchChain generates this branch's chain (ex: from the commit)
func CreateThisBranchChain(t *testing.T, numVals, numFull int) []ibc.Chain {
	return CreateChain(t, numVals, numFull, NeutaroImage)
}

func CreateChainWithCustomConfig(t *testing.T, numVals, numFull int, config ibc.ChainConfig) []ibc.Chain {
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "neutaro",
			ChainName:     "neutaro",
			Version:       config.Images[0].Version,
			ChainConfig:   config,
			NumValidators: &numVals,
			NumFullNodes:  &numFull,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// chain := chains[0].(*cosmos.CosmosChain)
	return chains
}

func BuildInitialChain(t *testing.T, chains []ibc.Chain) (*interchaintest.Interchain, context.Context, *client.Client, string) {
	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()

	for _, chain := range chains {
		ic = ic.AddChain(chain)
	}

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err := ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	return ic, ctx, client, network
}
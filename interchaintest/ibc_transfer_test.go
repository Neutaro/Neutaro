package interchaintest

import (
	"context"
	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// Copying a lot of this from Juno (https://github.com/CosmosContracts/juno/blob/main/interchaintest/ibc_transfer_test.go) to have a base test to start from
func TestNeutaroGaiaIBCTransfer(t *testing.T) {
	// Create chain factory with Neutaro and Gaia
	numVals := 1
	numFullNodes := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "Neutaro",
			ChainConfig:   neutaroConfig,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "gaia",
			Version:       "v9.1.0",
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	const (
		path = "ibc-path"
	)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	client, network := interchaintest.DockerSetup(t)

	neutaro, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	relayerType, relayerName := ibc.CosmosRly, "relay"

	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		relayerType,
		zaptest.NewLogger(t),
		interchaintestrelayer.CustomDockerImage(IBCRelayerImage, IBCRelayerVersion, "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)

	r := rf.Build(t, client, network)

	ic := interchaintest.NewInterchain().
		AddChain(neutaro).
		AddChain(gaia).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  neutaro,
			Chain2:  gaia,
			Relayer: r,
			Path:    path,
		})

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Create some user accounts on both chains
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, neutaro, gaia)

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, neutaro, gaia)
	require.NoError(t, err)

	// Get our Bech32 encoded user addresses
	neutaroUser, gaiaUser := users[0], users[1]

	neutaroUserAddr := neutaroUser.FormattedAddress()
	gaiaUserAddr := gaiaUser.FormattedAddress()

	// Get original account balances
	neutaroOrigBal, err := neutaro.GetBalance(ctx, neutaroUserAddr, neutaro.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletAmount, neutaroOrigBal.Int64())

	gaiaOrigBal, err := gaia.GetBalance(ctx, gaiaUserAddr, gaia.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletAmount, gaiaOrigBal.Int64())

	// Compose an IBC transfer and send from Neutaro -> Gaia
	var transferAmount = math.NewInt(1_000)
	transfer := ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   neutaro.Config().Denom,
		Amount:  transferAmount,
	}

	channel, err := ibc.GetTransferChannel(ctx, r, eRep, neutaro.Config().ChainID, gaia.Config().ChainID)
	require.NoError(t, err)

	neutaroHeight, err := neutaro.Height(ctx)
	require.NoError(t, err)

	transferTx, err := neutaro.SendIBCTransfer(ctx, channel.ChannelID, neutaroUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	err = r.StartRelayer(ctx, eRep, path)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, neutaro, neutaroHeight, neutaroHeight+50, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 10, neutaro)
	require.NoError(t, err)

	// Get the IBC denom for uneutaro on Gaia
	neutaroTokenDenom := transfertypes.GetPrefixedDenom(channel.Counterparty.PortID, channel.Counterparty.ChannelID, neutaro.Config().Denom)
	neutaroIBCDenom := transfertypes.ParseDenomTrace(neutaroTokenDenom).IBCDenom()

	// Assert that the funds are no longer present in user acc on Neutaro and are in the user acc on Gaia
	neutaroUpdateBal, err := neutaro.GetBalance(ctx, neutaroUserAddr, neutaro.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, neutaroOrigBal.Sub(transferAmount), neutaroUpdateBal)

	gaiaUpdateBal, err := gaia.GetBalance(ctx, gaiaUserAddr, neutaroIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, gaiaUpdateBal)

	// Compose an IBC transfer and send from Gaia -> Neutaro
	transfer = ibc.WalletAmount{
		Address: neutaroUserAddr,
		Denom:   neutaroIBCDenom,
		Amount:  transferAmount,
	}

	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	transferTx, err = gaia.SendIBCTransfer(ctx, channel.Counterparty.ChannelID, gaiaUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on Neutaro and not on Gaia
	neutaroUpdateBal, err = neutaro.GetBalance(ctx, neutaroUserAddr, neutaro.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, neutaroOrigBal, neutaroUpdateBal)

	gaiaUpdateBal, err = gaia.GetBalance(ctx, gaiaUserAddr, neutaroIBCDenom)
	require.NoError(t, err)
	require.Equal(t, int64(0), gaiaUpdateBal.Int64())
}

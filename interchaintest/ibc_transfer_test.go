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
	neutaroUser, gaiaUser := users[0], users[1]

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

	IBCTransferWorksTest(t, ctx, neutaro, gaia, neutaroUser, gaiaUser, r, eRep, path)

}

func IBCTransferWorksTest(
	t *testing.T,
	ctx context.Context,
	srcChain *cosmos.CosmosChain,
	dstChain *cosmos.CosmosChain,
	srcUser ibc.Wallet,
	dstUser ibc.Wallet, 
	r ibc.Relayer, 
	eRep *testreporter.RelayerExecReporter,
	path string) {
	// Wait a few blocks for relayer to start and for user accounts to be created
	err := testutil.WaitForBlocks(ctx, 5, srcChain, dstChain)
	require.NoError(t, err)

	srcUserAddr := srcUser.FormattedAddress()
	dstUserAddr := dstUser.FormattedAddress()

	// Get original account balances
	srcOrigBal, err := srcChain.GetBalance(ctx, srcUserAddr, srcChain.Config().Denom)
	require.NoError(t, err)

	// Compose an IBC transfer and send from srcChain -> dstChain
	var transferAmount = math.NewInt(1_000)
	transfer := ibc.WalletAmount{
		Address: dstUserAddr,
		Denom:   srcChain.Config().Denom,
		Amount:  transferAmount,
	}

	channel, err := ibc.GetTransferChannel(ctx, r, eRep, srcChain.Config().ChainID, dstChain.Config().ChainID)
	require.NoError(t, err)

	srcHeight, err := srcChain.Height(ctx)
	require.NoError(t, err)

	transferTx, err := srcChain.SendIBCTransfer(ctx, channel.ChannelID, srcUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, srcChain, srcHeight, srcHeight+50, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 10, srcChain)
	require.NoError(t, err)

	// Get the IBC denom for srcChain on dstChain
	srcTokenDenom := transfertypes.GetPrefixedDenom(channel.Counterparty.PortID, channel.Counterparty.ChannelID, srcChain.Config().Denom)
	srcIBCDenom := transfertypes.ParseDenomTrace(srcTokenDenom).IBCDenom()

	// Assert that the funds are no longer present in user acc on srcChain and are in the user acc on dstChain
	srcUpdateBal, err := srcChain.GetBalance(ctx, srcUserAddr, srcChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, srcOrigBal.Sub(transferAmount), srcUpdateBal)

	dstUpdateBal, err := dstChain.GetBalance(ctx, dstUserAddr, srcIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, dstUpdateBal)

	// Compose an IBC transfer and send from dstChain -> srcChain
	transfer = ibc.WalletAmount{
		Address: srcUserAddr,
		Denom:   srcIBCDenom,
		Amount:  transferAmount,
	}

	dstHeight, err := dstChain.Height(ctx)
	require.NoError(t, err)

	transferTx, err = dstChain.SendIBCTransfer(ctx, channel.Counterparty.ChannelID, dstUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, dstChain, dstHeight, dstHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on srcChain and not on dstChain
	srcUpdateBal, err = srcChain.GetBalance(ctx, srcUserAddr, srcChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, srcOrigBal, srcUpdateBal)

	dstUpdateBal, err = dstChain.GetBalance(ctx, dstUserAddr, srcIBCDenom)
	require.NoError(t, err)
	require.Equal(t, int64(0), dstUpdateBal.Int64())
}

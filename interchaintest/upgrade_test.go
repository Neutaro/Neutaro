package interchaintest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
)

const (
	chainName   = "neutaro"
	upgradeName = "v2"

	haltHeightDelta    = uint64(9) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = uint64(7)
)

var (
	// baseChain is the current version of the chain that will be upgraded from
	baseChain = ibc.DockerImage{
		Repository: "neutaro",
		Version:    "v1.0.0",
		UidGid:     "1025:1025",
	}
)

func TestBasicNeutaroUpgrade(t *testing.T) {
	repo, version := "neutaro", "local"
	CosmosChainUpgradeTest(t, chainName, version, repo, upgradeName)
}

func CosmosChainUpgradeTest(t *testing.T, chainName, upgradeBranchVersion, upgradeRepo, upgradeName string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	t.Log(chainName, upgradeBranchVersion, upgradeRepo, upgradeName)

	previousVersionGenesis := []cosmos.GenesisKV{
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

	cfg := neutaroConfig
	cfg.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)
	cfg.Images = []ibc.DockerImage{baseChain}

	numVals, numNodes := 4, 0
	chains := CreateChainWithCustomConfig(t, numVals, numNodes, cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	ic, ctx, client, _ := BuildInitialChain(t, chains)
	t.Cleanup(func() {
		_ = ic.Close()
	})

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	// upgrade
	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta
	proposalID := SubmitUpgradeProposal(t, ctx, chain, chainUser, upgradeName, haltHeight)

	ValidatorVoting(t, ctx, chain, proposalID, height, haltHeight)

	UpgradeNodes(t, ctx, chain, client, haltHeight, upgradeRepo, upgradeBranchVersion)

	// Post Upgrade: Conformance Validation
	StdExecute(t, ctx, chain, chainUser)
	subMsg(t, ctx, chain, chainUser)
	// TODO: ibc conformance test
}

func UpgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, haltHeight uint64, upgradeRepo, upgradeBranchVersion string) {
	// bring down nodes to prepare for upgrade
	t.Log("stopping node(s)")
	err := chain.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version on all nodes
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(ctx, client, upgradeRepo, upgradeBranchVersion)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	t.Log("starting node(s)")
	err = chain.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*60)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")
}

func ValidatorVoting(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, height uint64, haltHeight uint64) {
	err := chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+haltHeightDelta, proposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height), chain)

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")
}

func SubmitUpgradeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string, haltHeight uint64) string {
	tx, err := chain.UpgradeProposal(ctx, user.KeyName(), cosmos.SoftwareUpgradeProposal{
		Deposit:     fmt.Sprintf(`500000000%s`, chain.Config().Denom),
		Title:       "Chain Upgrade 1",
		Name:        upgradeName,
		Description: "Summary desc",
		Height:      haltHeight,
		Info:        "",
	})
	require.NoError(t, err, "error creating upgrade proposal")

	return tx.ProposalID
}
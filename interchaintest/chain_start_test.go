package interchaintest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Pretty much a direct copy from Juno: https://github.com/CosmosContracts/juno/blob/main/interchaintest/chain_start_test.go

// TestBasicNeutaroStart is a basic test to assert that spinning up a Neutaro network with one validator works as expected
func TestBasicNeutaroStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// TODO: Add this back when we've upgraded wasmd:
	/*
	chain := chains[0].(*cosmos.CosmosChain)

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	neutaroconformance.ConformanceCosmWasm(t, ctx, chain, chainUser)
	 */

	require.NotNil(t, ic)
	require.NotNil(t, ctx)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
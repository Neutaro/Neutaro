package interchaintest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Initial version from Juno: https://github.com/CosmosContracts/juno/blob/main/interchaintest/chain_start_test.go

// TestBasicNeutaroStart is a basic test to assert that spinning up a Neutaro network with one validator
func TestBasicNeutaroStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	require.NotNil(t, ic)
	require.NotNil(t, ctx)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
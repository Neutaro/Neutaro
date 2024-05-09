package interchaintest

import (
	"context"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/Neutaro/Neutaro/interchaintest/helpers"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCosmWasmConformance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	chain := chains[0].(*cosmos.CosmosChain)

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	StdExecute(t, ctx, chain, chainUser)
	subMsg(t, ctx, chain, chainUser)

	require.NotNil(t, ic)
	require.NotNil(t, ctx)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}

func StdExecute(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) (contractAddr string) {
	_, contractAddr = helpers.SetupContract(t, ctx, chain, user.KeyName(), "contracts/cw_template.wasm", `{"count":0}`)
	helpers.ExecuteMsgWithFee(t, ctx, chain, user, contractAddr, "", "10000"+chain.Config().Denom, `{"increment":{}}`)

	var res helpers.GetCountResponse
	err := helpers.SmartQueryString(t, ctx, chain, contractAddr, `{"get_count":{}}`, &res)
	require.NoError(t, err)

	require.Equal(t, int64(1), res.Data.Count)

	return contractAddr
}

func subMsg(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) {
	// ref: https://github.com/CosmWasm/wasmd/issues/1735

	// === execute a contract sub message ===
	_, senderContractAddr := helpers.SetupContract(t, ctx, chain, user.KeyName(), "contracts/cw721_base.wasm.gz", fmt.Sprintf(`{"name":"Reece #00001", "symbol":"juno-reece-test-#00001", "minter":"%s"}`, user.FormattedAddress()))
	_, receiverContractAddr := helpers.SetupContract(t, ctx, chain, user.KeyName(), "contracts/cw721_receiver.wasm.gz", `{}`)

	// mint a token
	res, err := helpers.ExecuteMsgWithFeeReturn(t, ctx, chain, user, senderContractAddr, "", "10000"+chain.Config().Denom, fmt.Sprintf(`{"mint":{"token_id":"00000", "owner":"%s"}}`, user.FormattedAddress()))
	fmt.Println("First", res)
	require.NoError(t, err)

	// this purposely will fail with the current, we are just validating the messsage is not unknown.
	// sub message of unknown means the `wasmkeeper.WithMessageHandlerDecorator` is not setup properly.
	fail := "ImZhaWwi"
	res2, err := helpers.ExecuteMsgWithFeeReturn(t, ctx, chain, user, senderContractAddr, "", "10000"+chain.Config().Denom, fmt.Sprintf(`{"send_nft": { "contract": "%s", "token_id": "00000", "msg": "%s" }}`, receiverContractAddr, fail))
	require.NoError(t, err)
	fmt.Println("Second", res2)
	require.NotEqualValues(t, wasmtypes.ErrUnknownMsg.ABCICode(), res2.Code)
	require.NotContains(t, res2.RawLog, "unknown message from the contract")

	success := "InN1Y2NlZWQi"
	res3, err := helpers.ExecuteMsgWithFeeReturn(t, ctx, chain, user, senderContractAddr, "", "10000"+chain.Config().Denom, fmt.Sprintf(`{"send_nft": { "contract": "%s", "token_id": "00000", "msg": "%s" }}`, receiverContractAddr, success))
	require.NoError(t, err)
	fmt.Println("Third", res3)
	require.EqualValues(t, 0, res3.Code)
	require.NotContains(t, res3.RawLog, "unknown message from the contract")
}

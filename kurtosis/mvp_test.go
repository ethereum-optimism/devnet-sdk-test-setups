package kurtosis

import (
	"devnetsdktest/bindings/mockERC20"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/devnet-sdk/contracts/constants"
	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	sdktypes "github.com/ethereum-optimism/optimism/devnet-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestIsthmusInitiateWithdrawal(t *testing.T) {
	chainIdx := uint64(0) // We'll use the first L2 chain for this test

	walletGetter, walletValidator := validators.AcquireL2WalletWithFunds(chainIdx, sdktypes.NewBalance(big.NewInt(1.0*constants.ETH)))
	lowLevelSystemGetter, lowLevelSystemValidator := validators.AcquireLowLevelSystem()

	systest.SystemTest(t,
		func(t systest.T, sys system.System) {
			ctx := t.Context()

			user := walletGetter(ctx)

			lowLevelSystem := lowLevelSystemGetter(ctx)
			chain := lowLevelSystem.L2s()[chainIdx]

			// Glue code between devnet-sdk and abigen bindings
			client, err := chain.Client()
			require.NoError(t, err)

			// Ugly code for signing transactions
			signer := types.NewLondonSigner(chain.ID())
			signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
				return types.SignTx(tx, signer, user.PrivateKey())
			}

			// Ugly code for getting the fees
			block, err := client.BlockByNumber(ctx, nil)
			require.NoError(t, err)

			tipCap, err := client.SuggestGasTipCap(ctx)
			require.NoError(t, err)

			t.Log("Deploying an ERC20")

			// We need to deploy an ERC20 as a test prerequisite
			_, deployTx, mockERC20, err := mockERC20.DeployMockERC20(&bind.TransactOpts{
				Signer:    signerFn,
				From:      user.Address(),
				Context:   ctx,
				GasFeeCap: block.BaseFee(),
				GasTipCap: tipCap,
			}, client)
			require.NoError(t, err)

			_, err = bind.WaitMined(ctx, client, deployTx)
			require.NoError(t, err)

			// Log log log
			t.Log("Deployed an ERC20")

			// Mint some tokens to the user
			mintTx, err := mockERC20.Mint(&bind.TransactOpts{
				Signer:    signerFn,
				From:      user.Address(),
				GasFeeCap: block.BaseFee(),
				GasTipCap: tipCap,
			}, user.Address(), big.NewInt(9000000000))
			require.NoError(t, err)

			_, err = bind.WaitMined(ctx, client, mintTx)
			require.NoError(t, err)

			// Log log log
			t.Log("Minted some tokens")
		},
		walletValidator,
		lowLevelSystemValidator,
	)
}

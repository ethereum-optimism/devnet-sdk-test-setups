package kurtosis

import (
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/devnet-sdk/contracts/constants"
	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	sdktypes "github.com/ethereum-optimism/optimism/devnet-sdk/types"
	bindings "github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestIsthmusInitiateWithdrawal(t *testing.T) {
	chainIdx := uint64(0) // We'll use the first L2 chain for this test

	walletGetter, fundsValidator := validators.AcquireL2WalletWithFunds(chainIdx, sdktypes.NewBalance(big.NewInt(1.0*constants.ETH)))
	lowLevelSystemGetter, validator := validators.AcquireLowLevelSystem()

	systest.SystemTest(t,
		func(t systest.T, sys system.System) {
			ctx := t.Context()

			user := walletGetter(ctx)

			lowLevelSystem := lowLevelSystemGetter(ctx)
			chain := lowLevelSystem.L2s()[chainIdx]

			// Glue code between devnet-sdk and abigen bindings
			client, err := chain.Client()
			require.NoError(t, err)

			// Ugly code required to deploy an ERC20
			erc20factory, err := bindings.NewOptimismMintableERC20Factory(common.HexToAddress("0x4200000000000000000000000000000000000012"), client)
			require.NoError(t, err)

			// Ugly code for signing transactions
			signer := types.NewEIP155Signer(chain.ID())
			signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
				return types.SignTx(tx, signer, user.PrivateKey())
			}

			// block, err := client.BlockByNumber(ctx, nil)
			// require.NoError(t, err)

			// tipCap, err := client.SuggestGasTipCap(ctx)
			// require.NoError(t, err)

			t.Log("Creating an OptimismMintableERC20")

			// Now we create a new ERC20 token
			createTx, err := erc20factory.CreateOptimismMintableERC20(&bind.TransactOpts{
				Signer:  signerFn,
				From:    user.Address(),
				Context: ctx,
			}, common.HexToAddress("0x0000000000000000000000000000000000000007"), "Mock", "MCK")
			require.NoError(t, err)

			// Wait for it to land
			createReceipt, err := bind.WaitMined(ctx, client, createTx)
			require.NoError(t, err)

			// Log log log
			t.Log("Created an OptimismMintableERC20")

			// Now let's find the factory event that contains the address of the newly created ERC20
			var erc20CreatedEvent *bindings.OptimismMintableERC20FactoryOptimismMintableERC20Created
			for _, log := range createReceipt.Logs {
				erc20CreatedEvent, err = erc20factory.ParseOptimismMintableERC20Created(*log)
				if err == nil {
					break
				}
			}
			require.NotNil(t, erc20CreatedEvent)

			// We create a binding for the new token
			erc20, err := bindings.NewOptimismMintableERC20(erc20CreatedEvent.LocalToken, client)
			require.NoError(t, err)

			// Mint some tokens
			mintTx, err := erc20.Mint(&bind.TransactOpts{
				Signer:    signerFn,
				From:      user.Address(),
				GasFeeCap: big.NewInt(200),
				GasTipCap: big.NewInt(100),
			}, user.Address(), big.NewInt(9000000000))
			require.NoError(t, err)

			_, err = bind.WaitMined(ctx, client, mintTx)
			require.NoError(t, err)

			// Log log log
			t.Log("Minted some tokens")
		},
		fundsValidator,
		validator,
	)
}

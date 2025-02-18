package kurtosis

import (
	"log/slog"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/devnet-sdk/contracts/constants"
	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	sdktypes "github.com/ethereum-optimism/optimism/devnet-sdk/types"
	"github.com/stretchr/testify/require"
)

func smokeTestScenario(chainIdx uint64, walletGetter validators.WalletGetter) systest.SystemTestFunc {
	return func(t systest.T, sys system.System) {
		ctx := t.Context()
		logger := slog.With("test", "TestMinimal", "devnet", sys.Identifier())

		chain := sys.L2(chainIdx)
		logger = logger.With("chain", chain.ID())
		logger.InfoContext(ctx, "starting test")

		funds := sdktypes.NewBalance(big.NewInt(0.5 * constants.ETH))
		user := walletGetter(ctx)

		scw0Addr := constants.SuperchainWETH
		scw0, err := chain.ContractsRegistry().SuperchainWETH(scw0Addr)
		require.NoError(t, err)
		logger.InfoContext(ctx, "using SuperchainWETH", "contract", scw0Addr)

		initialBalance, err := scw0.BalanceOf(user.Address()).Call(ctx)
		require.NoError(t, err)
		logger = logger.With("user", user.Address())
		logger.InfoContext(ctx, "initial balance retrieved", "balance", initialBalance)

		logger.InfoContext(ctx, "sending ETH to contract", "amount", funds)
		require.NoError(t, user.SendETH(scw0Addr, funds).Send(ctx).Wait())

		balance, err := scw0.BalanceOf(user.Address()).Call(ctx)
		require.NoError(t, err)
		logger.InfoContext(ctx, "final balance retrieved", "balance", balance)

		require.Equal(t, initialBalance.Add(funds), balance)
	}
}

func TestSystemWrapETH(t *testing.T) {
	chainIdx := uint64(0) // We'll use the first L2 chain for this test

	walletGetter, fundsValidator := validators.AcquireL2WalletWithFunds(chainIdx, sdktypes.NewBalance(big.NewInt(1.0*constants.ETH)))

	systest.SystemTest(t,
		smokeTestScenario(chainIdx, walletGetter),
		fundsValidator,
	)
}

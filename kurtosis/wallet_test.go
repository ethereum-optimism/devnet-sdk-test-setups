package kurtosis

import (
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/devnet-sdk/contracts/constants"
	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	sdktypes "github.com/ethereum-optimism/optimism/devnet-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestAcquireWalletWithFundsExample is an example test that uses AcquireL2WalletWithFunds validator
//
// This is useful when a funded L2 account is needed to run the test
func TestAcquireWalletWithFundsExample(t *testing.T) {
	chainIdx := uint64(0)
	requiredBalance := sdktypes.NewBalance(big.NewInt(1.0 * constants.ETH))
	walletGetter, walletValidator := validators.AcquireL2WalletWithFunds(chainIdx, requiredBalance)

	systest.SystemTest(t,
		func(t systest.T, sys system.System) {
			ctx := t.Context()

			wallet := walletGetter(ctx)
			require.Greater(t, wallet.Balance().Int, sdktypes.NewBalance(big.NewInt(0)).Int)
		},
		walletValidator,
	)
}

package kurtosis

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	"github.com/stretchr/testify/require"
)

// TestSystemBuildsBlocks ensures that all L2 chains build 100 blocks within specified timeframe
//
// FIXME https://github.com/ethereum-optimism/platforms-team/issues/650 is required to access all RPC URLs in a system
func TestSystemBuildsBlocks(t *testing.T) {
	lowLevelSysGetter, lowLevelSysValidator := validators.AcquireLowLevelSystem()

	systest.SystemTest(t,
		func(t systest.T, sys system.System) {
			ctx := t.Context()

			lowLevelSys := lowLevelSysGetter(ctx)
			l2s := lowLevelSys.L2s()

			numAttempts := 100
			targetBlockNumber := uint64(100)
			for attempt := range 100 {
				t.Logf("Checking blocks, attempt %d/%d", attempt, numAttempts)

				// We'll accumulate all L2s that have not yet reached the block target in this list
				pendingL2s := []system.LowLevelChain{}

				for _, l2 := range l2s {
					// We get hold of a client
					client, err := l2.Client()
					require.NoError(t, err)

					// We ask the node for its block number
					blockNumber, err := client.BlockNumber(ctx)
					require.NoError(t, err)

					if blockNumber < targetBlockNumber {
						// If the chain has not yet reached the block target, we push it into the pending list
						t.Logf("Chain %d: Pending, block %d", l2.ID(), blockNumber)

						pendingL2s = append(pendingL2s, l2)
					} else {
						// If the chain has reached the target we log and move on
						t.Logf("Chain %d: Reached block target", l2.ID())
					}

				}

				// We update the list of pending chains
				l2s = pendingL2s

				// And if there are no pending chains, we happily exit
				if len(l2s) == 0 {
					t.Logf("Reached block targets for all chains")
					return
				}

				// If there are any pending chains left, we sleep and poll again
				time.Sleep(5 * time.Second)
			}

			require.FailNow(t, "Did not reach target block number in time")

		},
		lowLevelSysValidator,
	)
}

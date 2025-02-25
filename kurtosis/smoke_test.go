package kurtosis

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	"github.com/stretchr/testify/require"
)

func TestSystemBuildsBlocks(t *testing.T) {
	lowLevelSysGetter, lowLevelSysValidator := validators.AcquireLowLevelSystem()

	t.Run("system keeps building blocks", func(t *testing.T) {
		systest.SystemTest(t,
			func(t systest.T, sys system.System) {
				ctx := t.Context()

				lowLevelSys := lowLevelSysGetter(ctx)
				l2s := lowLevelSys.L2s()

				numAttempts := 100
				targetBlockNumber := uint64(100)
				for attempt := range 100 {
					t.Logf("Checking blocks, attempt %d/%d", attempt, numAttempts)

					pendingL2s := []system.LowLevelChain{}

					for _, l2 := range l2s {
						client, err := l2.Client()
						require.NoError(t, err)

						blockNumber, err := client.BlockNumber(ctx)
						require.NoError(t, err)

						t.Logf("Chain %d: Got block %d", l2.ID(), blockNumber)

						if blockNumber < targetBlockNumber {
							pendingL2s = append(pendingL2s, l2)
						}

						t.Logf("Chain %d: Reached block target", l2.ID())
					}

					l2s = pendingL2s
					if len(l2s) == 0 {
						t.Logf("Reached block targets for all chains")
						return
					}

					time.Sleep(5 * time.Second)
				}

				require.FailNow(t, "Did not reach target block number in time")

			},
			lowLevelSysValidator,
		)
	})
}

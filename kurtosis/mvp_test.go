package kurtosis

import (
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/devnet-sdk/contracts/constants"
	"github.com/ethereum-optimism/optimism/devnet-sdk/system"
	goethereum "github.com/ethereum-optimism/optimism/devnet-sdk/system/periphery/go-ethereum"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/systest"
	"github.com/ethereum-optimism/optimism/devnet-sdk/testing/testlib/validators"
	sdktypes "github.com/ethereum-optimism/optimism/devnet-sdk/types"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func TestIsthmusInitiateWithdrawal(t *testing.T) {
	chainIdx := uint64(0) // We'll use the first L2 chain for this test

	walletGetter, walletValidator := validators.AcquireL2WalletWithFunds(chainIdx, sdktypes.NewBalance(big.NewInt(1.0*constants.ETH)))
	lowLevelSystemGetter, lowLevelSystemValidator := validators.AcquireLowLevelSystem()

	// This test should only run on isthmus forks
	isthmusForkGetter, isthmusForkValidator := validators.AcquireL2WithFork(chainIdx, rollup.Isthmus)

	systest.SystemTest(t,
		func(t systest.T, sys system.System) {
			ctx := t.Context()

			// We show when isthmus was activated, for visibility purposes
			chainConfig := isthmusForkGetter(ctx)
			t.Logf("Isthmus has been activated at %d", chainConfig.IsthmusTime)

			// We'll need a wallet to sign transactions
			user := walletGetter(ctx)

			lowLevelSystem := lowLevelSystemGetter(ctx)
			chain := lowLevelSystem.L2s()[chainIdx]

			// AREA OF IMPROVEMENT
			//
			// Glue code between devnet-sdk and abigen bindings
			client, err := chain.Client()
			require.NoError(t, err)

			// AREA OF IMPROVEMENT
			//
			// Glue code between devnet-sdk and geth client
			rpcClient, err := rpc.Dial(chain.RPCURL())
			require.NoError(t, err)
			gethClient := gethclient.New(rpcClient)

			// AREA OF IMPROVEMENT
			//
			// Ugly code for signing transactions
			signer := types.NewLondonSigner(chain.ID())
			signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
				return types.SignTx(tx, signer, user.PrivateKey())
			}

			feeEstimator := goethereum.NewEIP1559FeeEstimator(client)

			// Create message passer
			l2ToL1MessagePasser, err := bindings.NewL2ToL1MessagePasser(constants.L2ToL1MessagePasser, client)
			require.NoError(t, err)

			// Get a reference to a recent block
			block, err := client.BlockByNumber(ctx, nil)
			require.NoError(t, err)

			// Get the storage root hash before the withdrawal
			preProof, err := gethClient.GetProof(ctx, constants.L2ToL1MessagePasser, nil, block.Number())
			require.NoError(t, err)

			// Now it's time to perform the withdrawal
			//
			// First we'll need to estimate the transaction fees
			t.Log("Estimating InitiateWithdrawal transaction fees")
			initiateWithdrawalTxOpts, err := feeEstimator.EstimateFees(ctx, &bind.TransactOpts{
				Signer: signerFn,
				From:   user.Address(),
			})
			require.NoError(t, err)

			// Now we submit the transaction
			t.Log("Submitting InitiateWithdrawal transaction")
			initiateWithdrawalTx, err := l2ToL1MessagePasser.InitiateWithdrawal(initiateWithdrawalTxOpts, user.Address(), big.NewInt(1_000_000), []byte{})
			require.NoError(t, err)

			t.Logf("Submitted InitiateWithdrawal transaction: %s", initiateWithdrawalTx.Hash().Hex())

			// And wait for it to mine
			t.Log("Waiting for InitiateWithdrawal transaction")
			initiateWithdrawalReceipt, err := bind.WaitMined(ctx, client, initiateWithdrawalTx)
			require.NoError(t, err)

			t.Log("Initiated a withdrawal")

			// Get the storage root hash after the withdrawal
			postProof, err := gethClient.GetProof(ctx, constants.L2ToL1MessagePasser, nil, initiateWithdrawalReceipt.BlockNumber)
			require.NoError(t, err)

			require.NotEqual(t, postProof.StorageHash, preProof.StorageHash)

			// Make sure the storage root has changed
			initiateWithdrawalBlock, err := client.BlockByHash(ctx, initiateWithdrawalReceipt.BlockHash)
			require.NoError(t, err)

			// Make sure the block contains the new withdrawals root
			require.Equal(t, *initiateWithdrawalBlock.WithdrawalsRoot(), postProof.StorageHash)
		},
		isthmusForkValidator,
		walletValidator,
		lowLevelSystemValidator,
	)
}

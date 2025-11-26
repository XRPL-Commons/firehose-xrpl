package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/streamingfast/cli/sflags"
	"github.com/xrpl-commons/firehose-xrpl/decoder"
	"github.com/xrpl-commons/firehose-xrpl/rpc"
	"go.uber.org/zap"
)

func NewToolCheckLedgerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool-check-ledger",
		Short: "Check connectivity and fetch a sample ledger from XRPL",
		Long: `Connects to an XRPL RPC endpoint and fetches ledger information
to verify connectivity and data format.

Examples:
  # Check latest ledger
  firexrpl tool-check-ledger --endpoint https://s1.ripple.com:51234/

  # Check specific ledger
  firexrpl tool-check-ledger --endpoint https://s1.ripple.com:51234/ --ledger 32570

  # Use testnet
  firexrpl tool-check-ledger --endpoint https://s.altnet.rippletest.net:51234/
`,
		RunE: runToolCheckLedger,
	}

	cmd.Flags().String("endpoint", "https://s1.ripple.com:51234/", "XRPL RPC endpoint URL")
	cmd.Flags().Uint64("ledger", 0, "Specific ledger index to fetch (0 = latest)")
	cmd.Flags().Bool("decode-transactions", false, "Decode and display transaction details")
	cmd.Flags().Int("max-transactions", 5, "Maximum number of transactions to display")

	return cmd
}

func runToolCheckLedger(cmd *cobra.Command, args []string) error {
	endpoint := sflags.MustGetString(cmd, "endpoint")
	ledgerIndex := sflags.MustGetUint64(cmd, "ledger")
	decodeTransactions := sflags.MustGetBool(cmd, "decode-transactions")
	maxTransactions := sflags.MustGetInt(cmd, "max-transactions")

	logger, _ := zap.NewDevelopment()

	fmt.Printf("Connecting to XRPL endpoint: %s\n\n", endpoint)

	// Create client
	client, err := rpc.NewClient(endpoint, logger)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get latest ledger if not specified
	if ledgerIndex == 0 {
		fmt.Println("Fetching latest validated ledger...")
		latest, err := client.GetLatestLedger(ctx)
		if err != nil {
			return fmt.Errorf("failed to get latest ledger: %w", err)
		}
		ledgerIndex = latest.LedgerIndex
		fmt.Printf("Latest validated ledger: %d\n", ledgerIndex)
		fmt.Printf("Ledger hash: %s\n\n", latest.LedgerHash)
	}

	// Fetch the ledger with transactions
	fmt.Printf("Fetching ledger %d with transactions...\n", ledgerIndex)
	ledgerResult, err := client.GetLedger(ctx, ledgerIndex)
	if err != nil {
		return fmt.Errorf("failed to get ledger: %w", err)
	}

	ledger := ledgerResult.Ledger

	fmt.Printf("\n=== Ledger %d ===\n", ledger.LedgerIndex)
	fmt.Printf("Hash:               %s\n", ledger.LedgerHash)
	fmt.Printf("Parent Hash:        %s\n", ledger.ParentHash)
	fmt.Printf("Close Time:         %d (XRPL epoch)\n", ledger.CloseTime)
	fmt.Printf("Total Coins:        %s drops\n", ledger.TotalCoins)
	fmt.Printf("Account Hash:       %s\n", ledger.AccountHash)
	fmt.Printf("Transaction Hash:   %s\n", ledger.TransactionHash)
	fmt.Printf("Validated:          %v\n", ledgerResult.Validated)
	fmt.Printf("Transaction Count:  %d\n", len(ledger.Transactions))

	if len(ledger.Transactions) > 0 {
		fmt.Printf("\n=== Transactions ===\n")

		dec := decoder.NewDecoder(logger)

		displayed := 0
		for i, tx := range ledger.Transactions {
			if displayed >= maxTransactions {
				fmt.Printf("\n... and %d more transactions\n", len(ledger.Transactions)-displayed)
				break
			}

			fmt.Printf("\n--- Transaction %d ---\n", i)
			fmt.Printf("Hash: %s\n", tx.Hash)

			if tx.TxBlob != "" {
				fmt.Printf("TxBlob length: %d bytes\n", len(tx.TxBlob)/2)
			}
			if tx.Meta != "" {
				fmt.Printf("Meta length: %d bytes\n", len(tx.Meta)/2)
			}

			// Decode if requested
			if decodeTransactions && tx.TxBlob != "" {
				decoded, err := dec.DecodeTransaction(tx.TxBlob)
				if err != nil {
					fmt.Printf("Failed to decode: %v\n", err)
				} else {
					fmt.Printf("Type:    %s\n", decoded.TransactionType)
					fmt.Printf("Account: %s\n", decoded.Account)
					if decoded.Destination != "" {
						fmt.Printf("Destination: %s\n", decoded.Destination)
					}
					fmt.Printf("Fee:     %d drops\n", decoded.Fee)
					fmt.Printf("Sequence: %d\n", decoded.Sequence)
				}

				if tx.Meta != "" {
					meta, err := dec.DecodeMetadata(tx.Meta)
					if err != nil {
						fmt.Printf("Failed to decode metadata: %v\n", err)
					} else {
						fmt.Printf("Result:  %s\n", meta.TransactionResult)
						fmt.Printf("Affected Nodes: %d\n", len(meta.AffectedNodes))
					}
				}
			}

			displayed++
		}
	}

	fmt.Printf("\nCheck completed successfully!\n")
	return nil
}

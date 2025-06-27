package cmd

import (
	"fmt"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Retrieves and prints secret value to STDOUT",
	Long: `Retrieves and prints secret value to STDOUT. This is mainly used by the Summon conceal-summon provider.

	Example Usage:
	$ conceal summon show aws/access_key_id`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := keychain.GetSecret(args[0], "stdout"); err != nil {
			conceal.PrintError(fmt.Sprintf("Failed to get secret value from credential store: %v", err))
		}
	},
}

func init() {
	summonCmd.AddCommand(showCmd)
}

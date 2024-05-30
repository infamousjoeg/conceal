package cmd

import (
	"fmt"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all concealed secret names",
	Long: `Returns a list of conceal set secret names from the secret provider.
	
	Example Usage:
	$ conceal list`,
	Run: func(cmd *cobra.Command, args []string) {
		// List all secrets in keychain with service label `summon`
		accounts := keychain.ListSecrets()

		conceal.PrintInfo("The following Summon/Conceal accounts are in keychain:")

		uniqueAccounts := make(map[string]bool)
		for _, account := range accounts {
			value := account.Account
			if !uniqueAccounts[value] {
				uniqueAccounts[value] = true
				fmt.Println(account.Account)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

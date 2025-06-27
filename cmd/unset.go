package cmd

import (
	"fmt"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
)

// unsetCmd represents the unset command
var unsetCmd = &cobra.Command{
	Use:     "unset",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a secret from secret provider",
	Long: `Unset removes a secret name and secret value entry from your secret provider.
	
	Example Usage:
	$ conceal unset aws/access_key_id`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := keychain.DeleteSecret(args[0]); err != nil {
			conceal.PrintError(fmt.Sprintf("Failed to delete secret from credential store: %v", err))
			return
		}

		conceal.PrintSuccess("Secret successfully deleted from credential store.")
	},
}

func init() {
	rootCmd.AddCommand(unsetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unsetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unsetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

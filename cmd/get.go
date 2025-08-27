package cmd

import (
	"os"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"cp", "retrieve"},
	Short:   "Retrieves and copies secret value to clipboard",
	Long: `Retrieves and copies the secret name provided's secret value.
The secret value is copied to the clipboard for 15 seconds.

	Example Usage:
	$ conceal get aws/access_key_id`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check platform support
		conceal.CheckPlatformSupport()

		secretName := conceal.GetSecretName(args)
		conceal.PrintInfo("Adding secret value to clipboard for 15 seconds...")
		err := keychain.GetSecret(secretName, "clipboard")
		if err != nil {
			conceal.PrintError("Failed to get secret value from keychain.")
			os.Exit(1)
		}
		conceal.PrintSuccess("Secret cleared from clipboard.")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

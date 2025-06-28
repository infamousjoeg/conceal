package cmd

import (
	"fmt"
	"os"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getStdout bool

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"cp", "retrieve"},
	Short:   "Retrieve a secret",
	Long: `Retrieves the secret value for the given name. By default the value
is copied to the clipboard for 15 seconds. Use --stdout to print the value
directly instead which enables piping to other commands.

        Example Usage:
        $ conceal get aws/access_key_id
        $ conceal get aws/access_key_id --stdout | other-command`,
	Run: func(cmd *cobra.Command, args []string) {
		secretName := conceal.GetSecretName(args)
		delivery := "clipboard"
		if getStdout {
			delivery = "stdout"
			conceal.PrintInfo("Printing secret value to STDOUT...")
		} else {
			conceal.PrintInfo("Adding secret value to clipboard for 15 seconds...")
		}
		if err := keychain.GetSecret(secretName, delivery); err != nil {
			conceal.PrintError(fmt.Sprintf("Failed to get secret value from credential store: %v", err))
			os.Exit(1)
		}
		if !getStdout {
			conceal.PrintSuccess("Secret cleared from clipboard.")
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVar(&getStdout, "stdout", false, "print secret to STDOUT instead of clipboard")
}

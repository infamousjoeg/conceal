package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// updateCmd represents the get command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates a secret value in the secret provider",
	Long: `Updates a secret value within the secret.

	Example Usage:
	$ conceal update
	$ conceal update aws/access_key_id
	$ echo "new_secret_value" | conceal update aws/access_key_id`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if secret name is empty
		secretName := conceal.GetSecretName(args)

		// Check stdin for secret value
		var byteSecretVal []byte
		info, err := os.Stdin.Stat()
		if err != nil {
			conceal.PrintError("An error occurred while checking stdin. Exiting...")
		}

		// Update secret value from STDIN
		if (info.Mode() & os.ModeCharDevice) == 0 {
			// Reading from STDIN
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				conceal.PrintError("An error occurred while reading stdin. Exiting...")
			}
			byteSecretVal = []byte(strings.TrimSpace(input))
		} else {
			// Get secret value from user
			fmt.Println("Please enter the secret value: ")
			byteSecretVal, err = term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				conceal.PrintError("An error occurred trying to read password. Exiting...")
			}
		}

		// Update secret and secret value in the credential store
		err = keychain.UpdateSecret(secretName, byteSecretVal)
		if err != nil {
			conceal.PrintError("Failed to update secret value in credential store.")
		}

		conceal.PrintSuccess("Secret value updated successfully.")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

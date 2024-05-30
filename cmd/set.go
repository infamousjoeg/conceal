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

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:     "set",
	Aliases: []string{"add", "create"},
	Short:   "Add a secret name and value to secret provider",
	Long: `Sets a given secret name and secret value within the secret provider.
	
	Example Usage:
	$ conceal set aws/access_key_id
	$ echo "my_secret_value" | conceal set aws/access_key_id`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if secret name is empty
		secretName := conceal.GetSecretName(args)

		// Check if secret already exists to save the user time
		if keychain.SecretExists(secretName) {
			conceal.PrintError("Secret already exists in keychain. Please use `conceal update` instead.")
		}

		var byteSecretVal []byte
		info, err := os.Stdin.Stat()
		if err != nil {
			conceal.PrintError("An error occurred while checking stdin. Exiting...")
		}

		// Get secret value from STDIN
		if (info.Mode() & os.ModeCharDevice) == 0 {
			// Reading from STDIN
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				conceal.PrintError("An error occurred while reading stdin. Exiting...")
			}
			byteSecretVal = []byte(strings.TrimSpace(input))
		} else {
			fmt.Println("Please enter the secret value: ")
			byteSecretVal, err = term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				conceal.PrintError("An error occurred trying to read password. Exiting...")
			}
		}

		// Add secret and secret value to keychain
		err = keychain.AddSecret(secretName, byteSecretVal)
		if err != nil {
			conceal.PrintError("An error occurred while adding secret to keychain.")
		}

		conceal.PrintSuccess("Secret added to keychain.")
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

package cmd

import (
	"fmt"
	"log"
	"syscall"

	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Add a secret name and value to secret provider",
	Long: `Sets a given secret name and secret value within the secret provider.
	
	Example Usage:
	$ conceal set aws/access_key_id`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get secret value from STDIN
		fmt.Println("Please enter the secret value: ")
		byteSecretVal, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalln("an error occurred trying to read password from Stdin")
		}

		// Add secret and secret value to keychain
		err = keychain.AddSecret(args[0], byteSecretVal)
		if err != nil {
			log.Fatalf("%s", err)
		}

		fmt.Printf("Added %s successfully to keychain.\n", args[0])
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

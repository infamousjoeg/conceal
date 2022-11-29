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
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a secret value for a secret name that exists in a secret provider",
	Long: `Updates a given secret value for a secret name within the secret provider.
	
	Example Usage:
	$ conceal update aws/access_key_id`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get secret value from STDIN
		fmt.Println("Please enter the secret value: ")
		byteSecretVal, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalln("an error occurred trying to read password from Stdin")
		}

		// Update secret value for given secret name in keychain
		err = keychain.UpdateSecret(args[0], byteSecretVal)
		if err != nil {
			log.Fatalf("%s", err)
		}
		fmt.Printf("Successfully updated secret value for %s in keychain.\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

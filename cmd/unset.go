package cmd

import (
	"log"
	"runtime"

	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/infamousjoeg/conceal/pkg/conceal/wincred"
	"github.com/spf13/cobra"
)

// unsetCmd represents the unset command
var unsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Remove a secret from secret provider",
	Long: `Unset removes a secret name and secret value entry from your secret provider.
	
	Example Usage:
	$ conceal unset aws/access_key_id`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			wincred.DeleteSecret(args[0])
		case "darwin":
			keychain.DeleteSecret(args[0])
		default:
			log.Fatalf("Unsupported Operating System: %s\n", runtime.GOOS)
		}
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

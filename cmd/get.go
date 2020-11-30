package cmd

import (
	"log"
	"runtime"

	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/infamousjoeg/conceal/pkg/conceal/wincred"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves and copies secret value to clipboard",
	Long: `Retrieves and copies the secret name provided's secret value.
The secret value is copied to the clipboard for 15 seconds.

	Example Usage:
	$ conceal get aws/access_key_id`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			wincred.GetSecret(args[0])
		case "darwin":
			keychain.GetSecret(args[0])
		default:
			log.Fatalf("Unsupported Operating System: %s\n", runtime.GOOS)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

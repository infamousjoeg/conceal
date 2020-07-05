package cmd

import (
	"fmt"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display current version",
	Long: `Display the current version of conceal.
	
	Example Usage:
	$ conceal version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("conceal v%s\n", conceal.FullVersionName)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

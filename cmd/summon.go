package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var summonCmd = &cobra.Command{
	Use:   "summon",
	Short: "Commands related to Summon integration",
	Long: `This command group includes commands for integrating Conceal with Summon.
    
    Example Usage:
    $ conceal summon install`,
}

func init() {
	rootCmd.AddCommand(summonCmd)
}

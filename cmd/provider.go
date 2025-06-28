package cmd

import (
	"fmt"
	"github.com/infamousjoeg/conceal/internal/migrate"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "List available migration providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := migrate.NewManager()
		names, err := mgr.List()
		if err != nil {
			return err
		}
		for _, n := range names {
			fmt.Println(n)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(providerCmd)
}

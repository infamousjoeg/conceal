package cmd

import (
	"context"
	"github.com/infamousjoeg/conceal/internal/migrate"
	"github.com/spf13/cobra"
	"os"
)

var configureCmd = &cobra.Command{
	Use:   "configure [provider]",
	Short: "Configure a migration provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := migrate.NewManager()
		return mgr.Configure(context.Background(), args[0], os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

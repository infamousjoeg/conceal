package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/infamousjoeg/conceal/internal/migrate"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage migration providers",
	RunE:  listProviders,
}

var providerInstallCmd = &cobra.Command{
	Use:   "install <path>",
	Short: "Install a migration provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		mgr := migrate.NewManager()
		if err := mgr.Install(src); err != nil {
			return err
		}
		name := filepath.Base(src)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Installed %s\n", name)
		return nil
	},
}

func listProviders(cmd *cobra.Command, args []string) error {
	mgr := migrate.NewManager()
	names, err := mgr.List()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(names) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No providers installed")
		return nil
	}
	for _, n := range names {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), n)
	}
	return nil
}

func init() {
	providerCmd.AddCommand(providerInstallCmd)
	rootCmd.AddCommand(providerCmd)
}

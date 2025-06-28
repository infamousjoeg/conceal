package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/infamousjoeg/conceal/internal/migrate"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
)

var (
	toProvider string
	selector   string
	rate       int
	contErr    bool
	dryRun     bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Export secrets to an external provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := migrate.NewManager()
		secrets := keychain.ListSecrets()
		src := make(map[string][]byte)
		for _, s := range secrets {
			if selector == "" || strings.Contains(s, selector) {
				v, err := keychain.ReadSecret(s)
				if err != nil {
					return err
				}
				src[s] = []byte(v)
			}
		}
		if dryRun {
			for k := range src {
				fmt.Println("would migrate", k)
			}
			return nil
		}
		return mgr.Migrate(context.Background(), toProvider, src, rate, contErr, os.Stdout)
	},
}

func init() {
	migrateCmd.Flags().StringVar(&toProvider, "to", "", "target provider")
	migrateCmd.Flags().StringVar(&selector, "selector", "", "filter secrets")
	migrateCmd.Flags().IntVar(&rate, "rate", 0, "writes per second")
	migrateCmd.Flags().BoolVar(&contErr, "continue-on-error", false, "keep going after errors")
	migrateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show actions without writing")
	if err := migrateCmd.MarkFlagRequired("to"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(migrateCmd)
}

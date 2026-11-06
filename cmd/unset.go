package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newUnsetCommand(store secretStore) *cobra.Command {
	return &cobra.Command{Use: "unset <secret-name>", Aliases: []string{"rm", "delete"}, Short: "Delete a secret", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateSecretName(args[0]); err != nil {
			return err
		}
		if err := store.DeleteSecret(args[0]); err != nil {
			return err
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Secret deleted.")
		return err
	}}
}

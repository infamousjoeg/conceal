package cmd

import (
	"fmt"
	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
	"github.com/spf13/cobra"
)

func newGetCommand(store secretStore) *cobra.Command {
	return &cobra.Command{Use: "get [secret-name]", Aliases: []string{"cp", "retrieve"}, Short: "Copy a secret to the clipboard for 15 seconds", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		id, err := secretName(cmd, args)
		if err != nil {
			return err
		}
		value, err := store.ReadSecret(id)
		if err != nil {
			return err
		}
		defer clear(value)
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "Copying secret for 15 seconds; newer clipboard content will be preserved."); err != nil {
			return err
		}
		return clipboard.Secret(string(value))
	}}
}

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
)

func writeSecret(store secretStore, id string, out io.Writer) error {
	if err := validateSecretName(id); err != nil {
		return err
	}
	value, err := store.ReadSecret(id)
	if err != nil {
		return err
	}
	defer clear(value)
	_, err = out.Write(value)
	return err
}

func newShowCommand(store secretStore) *cobra.Command {
	return &cobra.Command{Use: "show <secret-name>", Short: "Write the exact secret value to stdout for a trusted consumer", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return writeSecret(store, args[0], cmd.OutOrStdout())
	}}
}

func executeProvider(store secretStore, args []string, out io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: conceal_summon <secret-name>")
	}
	if args[0] == "--version" || args[0] == "-v" {
		_, err := fmt.Fprintln(out, versionString())
		return err
	}
	return writeSecret(store, args[0], out)
}

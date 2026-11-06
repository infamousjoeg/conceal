package cmd

import "github.com/spf13/cobra"

func newSetCommand(store secretStore) *cobra.Command { return writeSecretCommand(store, false) }

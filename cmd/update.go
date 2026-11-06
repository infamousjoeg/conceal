package cmd

import "github.com/spf13/cobra"

func newUpdateCommand(store secretStore) *cobra.Command { return writeSecretCommand(store, true) }

package cmd

import "github.com/spf13/cobra"

func newSummonCommand(store secretStore) *cobra.Command {
	command := &cobra.Command{Use: "summon", Short: "Optional Summon provider integration"}
	command.AddCommand(newShowCommand(store), newInstallCommand())
	return command
}

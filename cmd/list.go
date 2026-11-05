package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
	"sort"
)

func sortedSecrets(store secretStore) ([]keychain.QueryResult, error) {
	items, err := store.List()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	result := []keychain.QueryResult{}
	for _, item := range items {
		if !seen[item.Account] {
			result = append(result, item)
			seen[item.Account] = true
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Account < result[j].Account })
	return result, nil
}

func newListCommand(store secretStore) *cobra.Command {
	command := &cobra.Command{Use: "list", Aliases: []string{"ls"}, Short: "List secret names without exposing values", Args: cobra.NoArgs}
	command.Flags().Bool("json", false, "Emit versioned JSON metadata for tools and agents")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		items, err := sortedSecrets(store)
		if err != nil {
			return err
		}
		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
				SchemaVersion int                    `json:"schema_version"`
				Secrets       []keychain.QueryResult `json:"secrets"`
			}{1, items})
		}
		for _, item := range items {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), item.Account); err != nil {
				return err
			}
		}
		return nil
	}
	return command
}

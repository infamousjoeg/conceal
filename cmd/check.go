package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/infamousjoeg/conceal/pkg/conceal/platform"
	"github.com/spf13/cobra"
	"runtime"
)

type secretCheck struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}
type preflight struct {
	SchemaVersion int           `json:"schema_version"`
	Platform      string        `json:"platform"`
	Backend       string        `json:"backend"`
	Ready         bool          `json:"ready"`
	Secrets       []secretCheck `json:"secrets"`
	Error         string        `json:"error,omitempty"`
}

func newCheckCommand(store secretStore) *cobra.Command {
	command := &cobra.Command{Use: "check <secret-name>...", Short: "Check agent secret requirements without retrieving values", Args: cobra.MinimumNArgs(1)}
	command.Flags().Bool("json", false, "Emit JSON; exit 0 if ready, 2 if missing, 3 if backend unavailable")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		for _, id := range args {
			if err := validateSecretName(id); err != nil {
				return err
			}
		}
		result := preflight{SchemaVersion: 1, Platform: runtime.GOOS, Backend: platform.GetSecretStoreName(), Ready: true, Secrets: []secretCheck{}}
		items, err := store.List()
		code := 0
		if err != nil {
			result.Ready = false
			result.Error = "credential store unavailable"
			code = 3
		} else {
			available := map[string]bool{}
			for _, item := range items {
				available[item.Account] = true
			}
			for _, id := range args {
				result.Secrets = append(result.Secrets, secretCheck{id, available[id]})
				if !available[id] {
					result.Ready = false
					code = 2
				}
			}
		}
		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			err = json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		} else if result.Error != "" {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), result.Error)
		} else {
			for _, item := range result.Secrets {
				status := "available"
				if !item.Available {
					status = "missing"
				}
				if _, err = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", item.Name, status); err != nil {
					break
				}
			}
		}
		if err != nil {
			return err
		}
		if code != 0 {
			return &statusError{code: code}
		}
		return nil
	}
	return command
}

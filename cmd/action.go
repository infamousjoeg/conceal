package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/infamousjoeg/conceal/internal/contracts"
	"github.com/spf13/cobra"
)

func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func newActionCommand(store secretStore) *cobra.Command {
	action := &cobra.Command{
		Use:   "action",
		Short: "Prepare, approve, and execute exact GitHub actions without exposing credentials",
		Long:  "Contracts restrict repositories and actions. Approvals bind one exact request and expire.\nThe approval state directory must remain outside an untrusted agent's writable sandbox.",
	}
	action.PersistentFlags().String("state-dir", "", "Trusted approval directory (default: user's config directory/conceal/actions)")
	manager := func(cmd *cobra.Command) (contracts.Manager, error) {
		directory, _ := cmd.Flags().GetString("state-dir")
		if directory == "" {
			base, err := os.UserConfigDir()
			if err != nil {
				return contracts.Manager{}, err
			}
			directory = filepath.Join(base, "conceal", "actions")
		}
		return contracts.Manager{Directory: directory, ReadSecret: store.ReadSecret}, nil
	}
	example := &cobra.Command{Use: "example", Short: "Print a sample contract for review and customization", Args: cobra.NoArgs}
	example.RunE = func(cmd *cobra.Command, _ []string) error {
		return writeJSON(cmd.OutOrStdout(), contracts.Contract{
			SchemaVersion: 1, Name: "repository-review", Credential: "github/token",
			Repositories:       []string{"OWNER/REPOSITORY"},
			Actions:            []string{contracts.ListIssues, contracts.CreateDraft},
			ApprovalTTLSeconds: 300,
		})
	}
	validate := &cobra.Command{Use: "validate <contract.json>", Short: "Validate a contract without reading credentials", Args: cobra.ExactArgs(1)}
	validate.RunE = func(cmd *cobra.Command, args []string) error {
		contract, err := contracts.LoadContract(args[0])
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), contract)
	}
	prepare := &cobra.Command{Use: "prepare", Short: "Prepare an exact request for review; no credential access or API call", Args: cobra.NoArgs}
	prepare.Flags().String("contract", "", "Path to the reviewed contract")
	prepare.Flags().String("operation", "", "github.issues.list or github.pulls.create-draft")
	prepare.Flags().String("repo", "", "Exact owner/repository")
	prepare.Flags().String("input", "", "JSON input file, or - for stdin (default: {})")
	prepare.RunE = func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("contract")
		operation, _ := cmd.Flags().GetString("operation")
		repo, _ := cmd.Flags().GetString("repo")
		inputPath, _ := cmd.Flags().GetString("input")
		if path == "" || operation == "" || repo == "" {
			return fmt.Errorf("--contract, --operation, and --repo are required")
		}
		input := []byte(`{}`)
		if inputPath != "" {
			reader := cmd.InOrStdin()
			if inputPath != "-" {
				file, err := os.Open(inputPath)
				if err != nil {
					return err
				}
				defer func() { _ = file.Close() }()
				reader = file
			}
			var err error
			input, err = io.ReadAll(io.LimitReader(reader, maxInputSize+1))
			if err != nil {
				return err
			}
			if len(input) > maxInputSize {
				return fmt.Errorf("action input exceeds 1 MiB")
			}
		}
		m, err := manager(cmd)
		if err != nil {
			return err
		}
		proposal, err := m.Prepare(path, operation, repo, input)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), proposal)
	}
	inspect := &cobra.Command{Use: "inspect <request-id>", Short: "Show the exact request and digest for review", Args: cobra.ExactArgs(1)}
	inspect.RunE = func(cmd *cobra.Command, args []string) error {
		m, err := manager(cmd)
		if err != nil {
			return err
		}
		proposal, err := m.Inspect(args[0])
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), proposal)
	}
	approve := &cobra.Command{
		Use:   "approve <request-id> --digest <reviewed-sha256>",
		Short: "Approve a reviewed digest once; run only from the trusted approving context",
		Args:  cobra.ExactArgs(1),
	}
	approve.Flags().String("digest", "", "Exact SHA-256 digest shown by inspect; approval is an explicit privileged action")
	approve.RunE = func(cmd *cobra.Command, args []string) error {
		reviewed, _ := cmd.Flags().GetString("digest")
		if reviewed == "" {
			return fmt.Errorf("inspect the request, then approve its exact --digest from the trusted host")
		}
		m, err := manager(cmd)
		if err != nil {
			return err
		}
		approval, err := m.Approve(args[0], reviewed)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), approval)
	}
	execute := &cobra.Command{Use: "execute <request-id>", Short: "Consume approval and perform one authenticated request", Args: cobra.ExactArgs(1)}
	execute.RunE = func(cmd *cobra.Command, args []string) error {
		m, err := manager(cmd)
		if err != nil {
			return err
		}
		result, actionErr := m.Execute(cmd.Context(), args[0])
		if result.Receipt.ID != "" {
			if err := writeJSON(cmd.OutOrStdout(), result); err != nil {
				return err
			}
		}
		return actionErr
	}
	revoke := &cobra.Command{Use: "revoke <request-id>", Short: "Prevent an unexecuted request from being used", Args: cobra.ExactArgs(1)}
	revoke.RunE = func(cmd *cobra.Command, args []string) error {
		m, err := manager(cmd)
		if err != nil {
			return err
		}
		receipt, err := m.Revoke(args[0])
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), receipt)
	}
	action.AddCommand(example, validate, prepare, inspect, approve, execute, revoke)
	return action
}

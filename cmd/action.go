package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/infamousjoeg/conceal/internal/adapters"
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
		Short: "Prepare, review, and execute approved service actions",
		Long:  "Contracts restrict resources and actions. Approvals bind one exact request and expire.\nThe approval state directory must remain outside an untrusted agent's writable sandbox.",
	}
	action.PersistentFlags().String("state-dir", "", "Trusted approval directory (default: user's config directory/conceal/actions)")
	manager := func(cmd *cobra.Command) (contracts.Manager, error) {
		directory, _ := cmd.Flags().GetString("state-dir")
		if directory == "" {
			base, err := contracts.StateDirectory()
			if err != nil {
				return contracts.Manager{}, err
			}
			directory = base
		}
		return contracts.Manager{Directory: directory, ReadSecret: store.ReadSecret}, nil
	}
	example := &cobra.Command{Use: "example", Short: "Print a sample contract for review and customization", Args: cobra.NoArgs}
	example.Flags().String("adapter", "", "Adapter for a schema-2 example: github or aws (default: legacy GitHub contract)")
	example.RunE = func(cmd *cobra.Command, _ []string) error {
		selected, _ := cmd.Flags().GetString("adapter")
		if selected != "" {
			contract := contracts.Contract{SchemaVersion: 2, Name: "repository-review", Adapter: selected, ApprovalTTLSeconds: 300}
			switch selected {
			case "github":
				contract.Credentials = map[string]string{"token": "github/token"}
				contract.Actions = []string{contracts.ListIssues, contracts.CreateDraft}
				contract.Scope = json.RawMessage(`{"repositories":["OWNER/REPOSITORY"]}`)
			case "aws":
				contract.Name = "s3-file-list"
				contract.Credentials = map[string]string{"access_key_id": "aws/access-key-id", "secret_access_key": "aws/secret-access-key", "session_token": "aws/session-token"}
				contract.Actions = []string{"aws.s3.objects.list"}
				contract.Scope = json.RawMessage(`{"region":"us-east-1","buckets":[{"name":"example-backups","prefix":"reports/"}]}`)
			default:
				return fmt.Errorf("no example for adapter %q", selected)
			}
			return writeJSON(cmd.OutOrStdout(), contract)
		}
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
	prepare.Flags().String("operation", "", "Installed operation name; see action adapters")
	prepare.Flags().String("repo", "", "Legacy alias for --resource: exact owner/repository")
	prepare.Flags().String("resource", "", "Exact resource allowed by the contract, such as owner/repository or bucket name")
	prepare.MarkFlagsMutuallyExclusive("repo", "resource")
	prepare.Flags().String("input", "", "JSON input file, or - for stdin (default: {})")
	prepare.RunE = func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("contract")
		operation, _ := cmd.Flags().GetString("operation")
		repo, _ := cmd.Flags().GetString("repo")
		resource, _ := cmd.Flags().GetString("resource")
		if resource != "" {
			repo = resource
		}
		inputPath, _ := cmd.Flags().GetString("input")
		if path == "" || operation == "" || repo == "" {
			return fmt.Errorf("--contract, --operation, and --resource (or --repo) are required")
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
	catalog := &cobra.Command{Use: "adapters", Short: "Describe installed operations, schemas, and credential roles", Args: cobra.NoArgs}
	catalog.RunE = func(cmd *cobra.Command, _ []string) error {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"schema_version": 1, "adapters": adapters.NewRegistry().Descriptors()})
	}
	inbox := &cobra.Command{Use: "inbox", Short: "List recent requests and their approval state as JSON", Args: cobra.NoArgs}
	inbox.Flags().Int("limit", 100, "Maximum recent requests (1-200)")
	inbox.RunE = func(cmd *cobra.Command, _ []string) error {
		m, err := manager(cmd)
		if err != nil {
			return err
		}
		limit, _ := cmd.Flags().GetInt("limit")
		items, err := m.List(limit)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), items)
	}
	review := &cobra.Command{Use: "review <request-id>", Short: "Return complete human review fields and allowed transitions", Args: cobra.ExactArgs(1)}
	review.RunE = func(cmd *cobra.Command, args []string) error {
		m, err := manager(cmd)
		if err != nil {
			return err
		}
		document, err := m.Review(args[0])
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), document)
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
	action.AddCommand(example, validate, prepare, inspect, approve, execute, revoke, catalog, inbox, review)
	return action
}

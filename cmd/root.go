package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/infamousjoeg/conceal/pkg/conceal"
	"github.com/infamousjoeg/conceal/pkg/conceal/keychain"
	"github.com/spf13/cobra"
)

type secretStore interface {
	List() ([]keychain.QueryResult, error)
	ReadSecret(string) ([]byte, error)
	AddSecret(string, []byte) error
	UpdateSecret(string, []byte) error
	DeleteSecret(string) error
}

type nativeStore struct{}

func (nativeStore) List() ([]keychain.QueryResult, error)   { return keychain.List() }
func (nativeStore) ReadSecret(id string) ([]byte, error)    { return keychain.ReadSecret(id) }
func (nativeStore) AddSecret(id string, value []byte) error { return keychain.AddSecret(id, value) }
func (nativeStore) UpdateSecret(id string, value []byte) error {
	return keychain.UpdateSecret(id, value)
}
func (nativeStore) DeleteSecret(id string) error { return keychain.DeleteSecret(id) }

type statusError struct {
	code    int
	message string
}

func (e *statusError) Error() string { return e.message }

func newRootCommand(store secretStore) *cobra.Command {
	root := &cobra.Command{
		Use: "conceal", Short: "Manage local secrets for applications and agents",
		Long:          "Store secrets in macOS Keychain, Windows Credential Manager, or the Linux kernel keyring.\nRun applications with selected secrets, check agent requirements, or integrate with Summon.",
		SilenceErrors: true, SilenceUsage: true,
	}
	root.PersistentFlags().Bool("non-interactive", false, "Fail instead of prompting for input")
	root.AddCommand(newSetCommand(store), newUpdateCommand(store), newGetCommand(store), newListCommand(store), newUnsetCommand(store), newRunCommand(store), newCheckCommand(store), newActionCommand(store), newSummonCommand(store), newVersionCommand())
	return root
}

// Execute runs the CLI, or the single-argument Summon protocol when installed as
// conceal_summon. The provider is the same native binary, without a shell wrapper.
func Execute() {
	var err error
	name := strings.TrimSuffix(strings.ToLower(filepath.Base(os.Args[0])), ".exe")
	if name == "conceal_summon" {
		err = executeProvider(nativeStore{}, os.Args[1:], os.Stdout)
	} else {
		err = newRootCommand(nativeStore{}).Execute()
	}
	if err != nil {
		code := 1
		var status *statusError
		if errors.As(err, &status) {
			code = status.code
		}
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, "conceal:", err)
		}
		os.Exit(code)
	}
}

func validateSecretName(name string) error {
	if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\x00\r\n") {
		return fmt.Errorf("secret name must be nonempty and contain no NUL or newline characters")
	}
	return nil
}

func versionString() string { return conceal.FullVersionName }

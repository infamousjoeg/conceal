package keychain

import (
	"fmt"
	"os"
	"strings"

	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
)

// QueryResult contains metadata about a secret, never its value.
type QueryResult struct {
	Account string `json:"name"`
}

func validateID(id string) error {
	if strings.TrimSpace(id) == "" || strings.ContainsAny(id, "\x00\r\n") {
		return fmt.Errorf("secret name must be nonempty and contain no NUL or newline characters")
	}
	return nil
}

func SecretExists(id string) bool { return validateID(id) == nil && secretExists(id) }

// List returns metadata for Conceal secrets, preserving backend errors.
func List() ([]QueryResult, error) { return listSecrets() }

func AddSecret(id string, value []byte) error {
	if err := validateID(id); err != nil {
		return err
	}
	return addSecret(id, value)
}

func DeleteSecret(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	return deleteSecret(id)
}

// ReadSecret returns caller-owned bytes. Clear them after use.
func ReadSecret(id string) ([]byte, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	return readSecret(id)
}

func UpdateSecret(id string, value []byte) error {
	if err := validateID(id); err != nil {
		return err
	}
	return updateSecret(id, value)
}

// ListSecrets preserves the original package API. New callers should use List
// so an unavailable credential store cannot be mistaken for an empty one.
func ListSecrets() []QueryResult {
	results, err := List()
	if err != nil {
		return []QueryResult{}
	}
	return results
}

// GetSecret delivers a secret through the clipboard or standard output.
func GetSecret(secretID, delivery string) error {
	if delivery != "clipboard" && delivery != "stdout" {
		return fmt.Errorf("unsupported delivery method %q", delivery)
	}
	value, err := ReadSecret(secretID)
	if err != nil {
		return err
	}
	defer clear(value)
	if delivery == "clipboard" {
		return clipboard.Secret(string(value))
	}
	_, err = os.Stdout.Write(value)
	return err
}

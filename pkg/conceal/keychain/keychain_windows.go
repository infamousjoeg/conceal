//go:build windows
// +build windows

package keychain

import (
	"fmt"

	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
	"github.com/infamousjoeg/conceal/pkg/conceal/wincred"
)

// QueryResult represents a query result for cross-platform compatibility
type QueryResult struct {
	Account string
}

// SecretExists is a boolean function to verify a secret is present in credential manager
func SecretExists(secretID string) bool {
	return wincred.SecretExists(secretID)
}

// ListSecrets returns all secrets in Windows Credential Manager with the summon prefix
func ListSecrets() []QueryResult {
	secrets, err := wincred.ListSecrets()
	if err != nil {
		return []QueryResult{}
	}

	// Convert to QueryResult format for compatibility
	results := make([]QueryResult, 0, len(secrets))
	for _, secret := range secrets {
		results = append(results, QueryResult{
			Account: secret.Account,
		})
	}

	return results
}

// AddSecret adds a secret to Windows Credential Manager
func AddSecret(secretID string, secret []byte) error {
	return wincred.AddSecret(secretID, secret)
}

// DeleteSecret removes a secret from Windows Credential Manager
func DeleteSecret(secretID string) error {
	return wincred.DeleteSecret(secretID)
}

// GetSecret retrieves a secret from Windows Credential Manager
func GetSecret(secretID string, delivery string) error {
	secretBytes, err := wincred.GetSecret(secretID, delivery)
	if err != nil {
		return err
	}

	password := string(secretBytes)
	if delivery == "clipboard" {
		clipboard.Secret(password)
	} else if delivery == "stdout" {
		fmt.Printf("%s", password)
	}
	// Clear password from memory for security
	for i := range password {
		password = password[:i] + "X" + password[i+1:]
	}

	return nil
}

// UpdateSecret updates a secret in Windows Credential Manager
func UpdateSecret(secretID string, secret []byte) error {
	return wincred.UpdateSecret(secretID, secret)
}

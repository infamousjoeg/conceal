//go:build windows

package keychain

import (
	"fmt"

	"github.com/danieljoos/wincred"
	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
)

// SecretExists checks if the secretID exists in Windows Credential Manager
func SecretExists(secretID string) bool {
	cred, err := wincred.GetGenericCredential(secretID)
	return err == nil && cred != nil
}

// ReadSecret returns the secret value as a string.
func ReadSecret(secretID string) (string, error) {
	cred, err := wincred.GetGenericCredential(secretID)
	if err != nil || cred == nil {
		return "", fmt.Errorf("secret not found")
	}
	return string(cred.CredentialBlob), nil
}

// ListSecrets returns all secrets stored by conceal in Credential Manager
func ListSecrets() []string {
	creds, _ := wincred.List()
	var results []string
	for _, c := range creds {
		if c != nil && c.Comment == "summon" {
			results = append(results, c.TargetName)
		}
	}
	return results
}

// AddSecret adds a secret to Credential Manager
func AddSecret(secretID string, secret []byte) error {
	if SecretExists(secretID) {
		return fmt.Errorf("Secret %s already exists in credential manager. Please use `conceal update` instead.", secretID)
	}
	cred := wincred.NewGenericCredential(secretID)
	cred.CredentialBlob = secret
	cred.Comment = "summon"
	if err := cred.Write(); err != nil {
		return fmt.Errorf("An unexpected error occurred trying to add secret %s to credential manager.", secretID)
	}
	if !SecretExists(secretID) {
		return fmt.Errorf("Secret %s was set but is not found in credential manager.", secretID)
	}
	return nil
}

// DeleteSecret removes a secret from Credential Manager
func DeleteSecret(secretID string) error {
	cred := wincred.NewGenericCredential(secretID)
	if err := cred.Delete(); err != nil {
		return fmt.Errorf("An error occurred trying to remove secret from credential manager. Secret '%s' not found.", secretID)
	}
	return nil
}

// GetSecret retrieves a secret and delivers it via clipboard or stdout
func GetSecret(secretID string, delivery string) error {
	cred, err := wincred.GetGenericCredential(secretID)
	if err != nil || cred == nil {
		return fmt.Errorf("An error occurred trying to get secret from credential manager. Secret '%s' not found.", secretID)
	}
	password := string(cred.CredentialBlob)
	if delivery == "clipboard" {
		clipboard.Secret(password)
	} else if delivery == "stdout" {
		fmt.Printf("%s", password)
	}
	password = ""
	return nil
}

// UpdateSecret updates an existing secret in Credential Manager
func UpdateSecret(secretID string, secret []byte) error {
	cred, err := wincred.GetGenericCredential(secretID)
	if err != nil || cred == nil {
		return fmt.Errorf("The secret %s does not exist in the credential manager. Please use `conceal set` instead.", secretID)
	}
	cred.CredentialBlob = secret
	if err := cred.Write(); err != nil {
		return fmt.Errorf("An unexpected error occurred trying to update secret %s in the credential manager.", secretID)
	}
	return nil
}

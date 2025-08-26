//go:build windows
// +build windows

package wincred

import (
	"fmt"

	"github.com/danieljoos/wincred"
)

// SecretExists checks if a secret is present in Windows Credential Manager
func SecretExists(secretID string) bool {
	cred, err := wincred.GetGenericCredential(getTargetName(secretID))
	if err != nil {
		return false
	}
	return cred != nil
}

// ListSecrets returns all secrets in Windows Credential Manager with the summon prefix
func ListSecrets() ([]SecretInfo, error) {
	creds, err := wincred.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	var secrets []SecretInfo
	prefix := "summon/"

	for _, cred := range creds {
		if len(cred.TargetName) > len(prefix) && cred.TargetName[:len(prefix)] == prefix {
			secretID := cred.TargetName[len(prefix):]
			secrets = append(secrets, SecretInfo{
				Account: secretID,
			})
		}
	}

	return secrets, nil
}

// AddSecret adds a secret to Windows Credential Manager
func AddSecret(secretID string, secret []byte) error {
	targetName := getTargetName(secretID)

	// Check if credential already exists
	if SecretExists(secretID) {
		return fmt.Errorf("Secret %s already exists in credential manager. Please use `conceal update` instead.", secretID)
	}

	cred := wincred.NewGenericCredential(targetName)
	cred.CredentialBlob = secret
	cred.Comment = "Summon secret managed by Conceal"

	err := cred.Write()
	if err != nil {
		return fmt.Errorf("An unexpected error occurred trying to add secret %s to the credential manager: %w", secretID, err)
	}

	// Verify the secret was set successfully
	if !SecretExists(secretID) {
		return fmt.Errorf("Secret %s was set but is not found in credential manager.", secretID)
	}

	return nil
}

// DeleteSecret removes a secret from Windows Credential Manager
func DeleteSecret(secretID string) error {
	targetName := getTargetName(secretID)

	cred, err := wincred.GetGenericCredential(targetName)
	if err != nil {
		return fmt.Errorf("An error occurred trying to remove secret from credential manager. Secret '%s' not found in credential manager.", secretID)
	}

	err = cred.Delete()
	if err != nil {
		return fmt.Errorf("An error occurred trying to remove secret from credential manager: %w", err)
	}

	return nil
}

// GetSecret retrieves a secret from Windows Credential Manager
func GetSecret(secretID string, delivery string) ([]byte, error) {
	targetName := getTargetName(secretID)

	cred, err := wincred.GetGenericCredential(targetName)
	if err != nil {
		return nil, fmt.Errorf("An error occurred trying to get secret from credential manager. Secret '%s' not found in credential manager.", secretID)
	}

	return cred.CredentialBlob, nil
}

// UpdateSecret updates a secret in Windows Credential Manager
func UpdateSecret(secretID string, secret []byte) error {
	targetName := getTargetName(secretID)

	// Check if credential exists
	if !SecretExists(secretID) {
		return fmt.Errorf("The secret %s does not exist in the credential manager. Please use `conceal set` instead.", secretID)
	}

	cred := wincred.NewGenericCredential(targetName)
	cred.CredentialBlob = secret
	cred.Comment = "Summon secret managed by Conceal"

	err := cred.Write()
	if err != nil {
		return fmt.Errorf("An unexpected error occurred trying to update secret %s in the credential manager: %w", secretID, err)
	}

	return nil
}

// SecretInfo represents basic information about a stored secret
type SecretInfo struct {
	Account string
}

// getTargetName creates a target name with summon prefix for Windows Credential Manager
func getTargetName(secretID string) string {
	return "summon/" + secretID
}

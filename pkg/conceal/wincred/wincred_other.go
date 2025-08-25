//go:build !windows
// +build !windows

package wincred

import "fmt"

// SecretExists is not supported on non-Windows platforms
func SecretExists(secretID string) bool {
	return false
}

// ListSecrets is not supported on non-Windows platforms
func ListSecrets() ([]SecretInfo, error) {
	return nil, fmt.Errorf("Windows Credential Manager is not supported on this platform")
}

// AddSecret is not supported on non-Windows platforms
func AddSecret(secretID string, secret []byte) error {
	return fmt.Errorf("Windows Credential Manager is not supported on this platform")
}

// DeleteSecret is not supported on non-Windows platforms
func DeleteSecret(secretID string) error {
	return fmt.Errorf("Windows Credential Manager is not supported on this platform")
}

// GetSecret is not supported on non-Windows platforms
func GetSecret(secretID string, delivery string) ([]byte, error) {
	return nil, fmt.Errorf("Windows Credential Manager is not supported on this platform")
}

// UpdateSecret is not supported on non-Windows platforms
func UpdateSecret(secretID string, secret []byte) error {
	return fmt.Errorf("Windows Credential Manager is not supported on this platform")
}

// SecretInfo represents basic information about a stored secret
type SecretInfo struct {
	Account string
}
//go:build !darwin && !windows
// +build !darwin,!windows

package keychain

import (
	"fmt"
)

// QueryResult represents a query result for cross-platform compatibility
type QueryResult struct {
	Account string
}

// SecretExists is not supported on this platform
func SecretExists(secretID string) bool {
	return false
}

// ListSecrets is not supported on this platform
func ListSecrets() []QueryResult {
	return []QueryResult{}
}

// AddSecret is not supported on this platform
func AddSecret(secretID string, secret []byte) error {
	return fmt.Errorf("Secret management is not supported on this platform. Only macOS and Windows are currently supported.")
}

// DeleteSecret is not supported on this platform
func DeleteSecret(secretID string) error {
	return fmt.Errorf("Secret management is not supported on this platform. Only macOS and Windows are currently supported.")
}

// GetSecret is not supported on this platform
func GetSecret(secretID string, delivery string) error {
	return fmt.Errorf("Secret management is not supported on this platform. Only macOS and Windows are currently supported.")
}

// UpdateSecret is not supported on this platform
func UpdateSecret(secretID string, secret []byte) error {
	return fmt.Errorf("Secret management is not supported on this platform. Only macOS and Windows are currently supported.")
}
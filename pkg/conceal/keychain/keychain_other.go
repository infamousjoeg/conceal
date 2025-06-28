//go:build !darwin && !windows && !linux

package keychain

import "fmt"

// SecretExists returns false because this platform has no credential store implementation.
func SecretExists(secretID string) bool {
	return false
}

// ListSecrets returns an empty slice on unsupported platforms.
func ListSecrets() []string {
	return []string{}
}

func ReadSecret(secretID string) (string, error) {
	return "", fmt.Errorf("unsupported platform")
}

// AddSecret always returns an error on unsupported platforms.
func AddSecret(secretID string, secret []byte) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

// DeleteSecret always returns an error on unsupported platforms.
func DeleteSecret(secretID string) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

// GetSecret always returns an error on unsupported platforms.
func GetSecret(secretID string, delivery string) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

// UpdateSecret always returns an error on unsupported platforms.
func UpdateSecret(secretID string, secret []byte) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

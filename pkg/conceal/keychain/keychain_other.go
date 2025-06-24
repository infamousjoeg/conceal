//go:build !darwin && !windows && !linux

package keychain

import "fmt"

func SecretExists(secretID string) bool {
	return false
}

func ListSecrets() []string {
	return []string{}
}

func AddSecret(secretID string, secret []byte) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

func DeleteSecret(secretID string) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

func GetSecret(secretID string, delivery string) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

func UpdateSecret(secretID string, secret []byte) error {
	return fmt.Errorf("conceal: credential store not supported on this OS")
}

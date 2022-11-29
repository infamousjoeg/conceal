package keychain

import "fmt"

// UpdateSecret is a non-return function that updates the secret value in keychain.
func UpdateSecret(secretID string, secret []byte) error {
	// Update generic password item in keychain
	err := DeleteSecret(secretID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %s", err)
	}

	err = AddSecret(secretID, secret)
	if err != nil {
		return fmt.Errorf("failed to add secret: %s", err)
	}

	return nil
}

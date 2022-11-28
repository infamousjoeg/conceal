package keychain

import (
	"fmt"

	"github.com/keybase/go-keychain"
)

// AddSecret is a non-return function that adds the secret and secret value to
// keychain.
func AddSecret(secretID string, secret []byte) error {
	// Add new generic password item to keychain
	item := keychain.NewGenericPassword(
		"summon", secretID, "summon", secret, "",
	)
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleAfterFirstUnlock)

	err := keychain.AddItem(item)

	// Duplicate item error
	if err == keychain.ErrorDuplicateItem {
		return fmt.Errorf("an error occurred trying to add a secret to keychain. secret '%s' already exists", secretID)
	}

	// Unexpected error
	if err != nil {
		return fmt.Errorf("an unexpected error occurred trying to add a secret to the keychain: %s", err)
	}

	// Verify the secret was set in keychain successfully
	secretExists, err := SecretExists(secretID)
	if err != nil {
		return fmt.Errorf("an error occurred trying to verify the secret was added to keychain: %s", err)
	}
	if !secretExists {
		return fmt.Errorf("secret %s not found in keychain", secret)
	}

	return nil
}

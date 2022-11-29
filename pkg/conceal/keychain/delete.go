package keychain

import (
	"fmt"

	"github.com/keybase/go-keychain"
)

// DeleteSecret is a non-return function that removes the secret from keychain.
func DeleteSecret(secretID string) error {
	err := keychain.DeleteGenericPasswordItem("summon", secretID)
	if err != nil {
		return fmt.Errorf("an error occurred trying to remove secret from keychain. secret '%s' not found in keychain", secretID)
	}

	return nil
}

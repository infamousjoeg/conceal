package keychain

import (
	"fmt"

	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
	"github.com/keybase/go-keychain"
)

// GetSecret is a non-return function that retrieves a secret and immediately
// adds it to the host clipboard for 15 seconds.
func GetSecret(secretID string) error {
	// Build query for secret retrieval from Keychain
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService("summon")
	query.SetAccount(secretID)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)
	if err != nil {
		// Error occurred
		return fmt.Errorf("an error occurred trying to get secret from keychain")
	} else if len(results) != 1 {
		// Not found
		return fmt.Errorf("an error occurred trying to get secret from keychain. secret '%s' not found in keychain", secretID)
	} else {
		password := string(results[0].Data)
		clipboard.Secret(password)
	}

	return nil
}

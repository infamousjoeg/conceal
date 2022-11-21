package keychain

import "github.com/keybase/go-keychain"

// ListSecrets is a string array function that returns all secrets in keychain
// with the label `summon`.
func ListSecrets() ([]string, error) {
	// Note: OSX use the term "account" to refer to the secret id.
	secretIDs, err := keychain.GetGenericPasswordAccounts("summon")
	if err != nil {
		return nil, err
	}

	return secretIDs, nil
}

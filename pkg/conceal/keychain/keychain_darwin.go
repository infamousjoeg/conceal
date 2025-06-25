//go:build darwin

package keychain

import (
	"fmt"
	"log"

	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
	gokeychain "github.com/keybase/go-keychain"
)

// SecretExists is a boolean function to verify a secret is present in keychain
func SecretExists(secretID string) bool {
	allSecretIDs := ListSecrets()

	// Search all the available secretIDs for this one
	for _, account := range allSecretIDs {
		if account == secretID {
			return true
		}
	}

	return false
}

// ListSecrets is a string array function that returns all secrets in keychain
// with the label `summon`.
func ListSecrets() []string {
	query := gokeychain.NewItem()
	query.SetSecClass(gokeychain.SecClassGenericPassword)
	query.SetService("summon")
	query.SetMatchLimit(gokeychain.MatchLimitAll)
	query.SetReturnAttributes(true)
	// Note: OSX use the term "account" to refer to the secret id.
	secretIDs, err := gokeychain.QueryItem(query)
	if err != nil {
		log.Fatalln(err)
	}
	var accounts []string
	for _, r := range secretIDs {
		accounts = append(accounts, r.Account)
	}
	return accounts
}

// AddSecret is a boolean function that adds the secret and secret value to
// keychain.
func AddSecret(secretID string, secret []byte) error {
	// Create a new keychain item
	item := gokeychain.NewItem()
	item.SetSecClass(gokeychain.SecClassGenericPassword)
	item.SetService("summon")
	item.SetAccount(secretID)
	item.SetLabel("summon")
	item.SetData(secret)
	item.SetSynchronizable(gokeychain.SynchronizableNo)
	item.SetAccessible(gokeychain.AccessibleAfterFirstUnlock)

	// Add new password item to keychain
	err := gokeychain.AddItem(item)

	// Duplicate item error
	if err == gokeychain.ErrorDuplicateItem {
		return fmt.Errorf("Secret %s already exists in keychain. Please use `conceal update` instead.", secretID)
	}

	// Unexpected error
	if err != nil {
		return fmt.Errorf("An unexpected error occurred trying to add secret %s to the keychain. Exiting...", secretID)
	}

	// Verify the secret was set in keychain successfully
	if !SecretExists(secretID) {
		return fmt.Errorf("Secret %s was set but is not found in keychain.", secretID)
	}

	return nil
}

// DeleteSecret is a boolean function that removes a secret from keychain.
func DeleteSecret(secretID string) error {
	item := gokeychain.NewItem()
	item.SetSecClass(gokeychain.SecClassGenericPassword)
	item.SetService("summon")
	item.SetAccount(secretID)

	err := gokeychain.DeleteItem(item)
	if err != nil {
		return fmt.Errorf("An error occurred trying to remove secret from keychain. Secret '%s' not found in keychain.", secretID)
	}

	return nil
}

// GetSecret is a boolean function that retrieves a secret and immediately
// adds it to the host clipboard for 15 seconds.
func GetSecret(secretID string, delivery string) error {
	// Build query for secret retrieval from Keychain
	query := gokeychain.NewItem()
	query.SetSecClass(gokeychain.SecClassGenericPassword)
	query.SetService("summon")
	query.SetAccount(secretID)
	query.SetMatchLimit(gokeychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := gokeychain.QueryItem(query)
	if err != nil {
		// Error occurred
		return fmt.Errorf("An error occurred trying to get secret from keychain.")
	} else if len(results) != 1 {
		// Not found
		return fmt.Errorf("An error occurred trying to get secret from keychain. Secret '%s' not found in keychain.", secretID)
	} else {
		password := string(results[0].Data)
		if delivery == "clipboard" {
			clipboard.Secret(password)
		} else if delivery == "stdout" {
			fmt.Printf("%s", password)
		}
		password = ""
	}

	return nil
}

// UpdateSecret is a non-return function that updates the secret value in keychain.
func UpdateSecret(secretID string, secret []byte) error {
	// Build query for secret retrieval from Keychain
	query := gokeychain.NewItem()
	query.SetSecClass(gokeychain.SecClassGenericPassword)
	query.SetService("summon")
	query.SetAccount(secretID)
	query.SetMatchLimit(gokeychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := gokeychain.QueryItem(query)
	if err != nil {
		// Error occurred
		return fmt.Errorf("An error occurred trying to get secret from keychain.")
	} else if len(results) != 1 {
		return fmt.Errorf("The secret %s does not exist in the keychain. Please use `conceal set` instead.", secretID)
	} else {
		// Create a new keychain item
		item := gokeychain.NewItem()
		item.SetSecClass(gokeychain.SecClassGenericPassword)
		item.SetService("summon")
		item.SetAccount(secretID)
		item.SetLabel("summon")
		item.SetData(secret)
		item.SetSynchronizable(gokeychain.SynchronizableNo)
		item.SetAccessible(gokeychain.AccessibleAfterFirstUnlock)

		// Update password item in keychain
		err := gokeychain.UpdateItem(query, item)
		if err != nil {
			return fmt.Errorf("An unexpected error occurred trying to update secret %s in the keychain.", secretID)
		}
	}

	return nil
}

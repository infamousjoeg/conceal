package keychain

import (
	"fmt"
	"log"
	"strings"
	"syscall"

	"github.com/keybase/go-keychain"
	"golang.org/x/crypto/ssh/terminal"
)

// SecretExists is a boolean function to verify a secret is present in keychain
func SecretExists(secretID string) bool {
	allsecretIDs := ListSecrets()

	// Search all the available secretIDs for this one
	for _, id := range allsecretIDs {
		if id == secretID {
			return true
		}
	}
	return false
}

// GetSecret is a string function that securely gets the secret value from user
func GetSecret() string {
	fmt.Println("Please enter the secret value: ")
	byteSecretVal, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalln("An error occurred trying to read password from " +
			"Stdin. Exiting...")
	}
	secret := string(byteSecretVal)

	return strings.TrimSpace(secret)
}

// ListSecrets is a string array function that returns all secrets in keychain
// with the label `summon`.
func ListSecrets() []string {
	// Note: OSX use the term "account" to refer to the secret id.
	secretIDs, err := keychain.GetGenericPasswordAccounts("summon")
	if err != nil {
		log.Fatalln(err)
	}

	return secretIDs
}

// AddSecret is a non-return function that adds the secret and secret value to
// keychain.
func AddSecret(secretID string) {
	// Add new generic password item to keychain
	secret := []byte(GetSecret())
	item := keychain.NewGenericPassword(
		"summon", secretID, "summon", secret, "",
	)
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleAfterFirstUnlock)

	err := keychain.AddItem(item)

	// Duplicate item error
	if err == keychain.ErrorDuplicateItem {
		log.Fatalf(
			"An error occurred trying to add a secret to keychain.\n"+
				"Secret '%s' already exists. Exiting...\n",
			secretID,
		)
	}

	// Unexpected error
	if err != nil {
		log.Fatalf(
			"An unexpected error occurred trying to add a secret to "+
				"the keychain:\n%s\nExiting...",
			err,
		)
	}

	// Verify the secret was set in keychain successfully
	if !SecretExists(secretID) {
		log.Fatalf("Secret %s not found in keychain. Exiting...\n", secret)
	}

	fmt.Printf("Added %s successfully to keychain.\n", secretID)
	return
}

// DeleteSecret is a non-return function that removes the secret from keychain
func DeleteSecret(secretID string) {
	err := keychain.DeleteGenericPasswordItem("summon", secretID)
	if err != nil {
		log.Fatalf(
			"An error occurred trying to remove secret from "+
				"keychain.\n  Secret '%s' not found in keychain. Exiting...\n",
			secretID,
		)
	}

	fmt.Printf("Removed %s successfully from keychain.\n", secretID)
	return
}

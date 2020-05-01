package keychain

import (
	"fmt"
	"log"
	"strings"
	"syscall"

	"github.com/keybase/go-keychain"
	"golang.org/x/crypto/ssh/terminal"
)

// CheckSecret is a boolean function to verify a secret is present in keychain
func CheckSecret(account string) {
	accounts := ListSecrets()

	// Search all the available accounts for this one
	found := false
	for _, acc := range accounts {
		if acc == account {
			found = true
		}
	}

	if found == false {
		log.Fatalf("Account %s not found in keychain. Exiting...\n", account)
	}
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
// with the label `summon`
func ListSecrets() []string {
	accounts, err := keychain.GetGenericPasswordAccounts("summon")
	if err != nil {
		log.Fatalln(err)
	}

	return accounts
}

// AddSecret is a non-return function that adds the secret and secret value to
// keychain
func AddSecret(account string) {
	// Add new generic password item to keychain
	item := keychain.NewGenericPassword("summon", account, "summon", []byte(GetSecret()), "")
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleAfterFirstUnlock)
	err := keychain.AddItem(item)
	if err == keychain.ErrorDuplicateItem {
		log.Fatalf(
			"An error occurred trying to add secret to keychain.\n"+
				"Account %s was already found. Cannot add duplicate secret. "+
				"Exiting...\n",
			account,
		)
	}

	// Verify the secret was set in keychain successfully
	CheckSecret(account)

	fmt.Printf("Added %s successfully to keychain.\n", account)
	return
}

// DeleteSecret is a non-return function that removes the secret from keychain
func DeleteSecret(account string) {
	err := keychain.DeleteGenericPasswordItem("summon", account)
	if err != nil {
		log.Fatalf(
			"An error occurred trying to remove secret from "+
				"keychain.\n  Account %s not found in keychain. Exiting...\n",
			account,
		)
	}

	fmt.Printf("Removed %s successfully from keychain.\n", account)
	return
}

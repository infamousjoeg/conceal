package keychain

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/infamousjoeg/go-conceal/pkg/conceal/util"
	"github.com/keybase/go-keychain"
	"golang.org/x/crypto/ssh/terminal"
)

// CheckSecret is a boolean function to verify a secret is present in keychain
func CheckSecret(account string) {
	accounts := ListSecrets()

	found := util.Contains(account, accounts)
	if found == false {
		fmt.Printf("Account %s was not found in keychain. Exiting...\n", account)
		os.Exit(1)
	}

	return
}

// GetSecret is a string function that securely gets the secret value from user
func GetSecret() string {
	fmt.Println("Please enter the secret value: ")
	byteSecretVal, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("An error occurred trying to read password from Stdin. Exiting...")
		os.Exit(1)
	}
	secret := string(byteSecretVal)

	return strings.TrimSpace(secret)
}

// ListSecrets is a string array function that returns all secrets in keychain with the label `Summon`
func ListSecrets() []string {
	accounts, err := keychain.GetGenericPasswordAccounts("summon")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	return accounts
}

// AddSecret is a non-return function that adds the secret and secret value to keychain
func AddSecret(account string) {
	// Get secret security from user via Stdin
	// secret := GetSecret()
	// if secret == "" {
	// 	fmt.Println("An error occurred trying to add secret to keychain.")
	// 	fmt.Println("The secret value entered was empty. A value is required. Try again...")
	// }

	// Add new generic password item to keychain
	item := keychain.NewGenericPassword("summon", account, "summon", []byte(GetSecret()), "")
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleAfterFirstUnlock)
	err := keychain.AddItem(item)
	if err == keychain.ErrorDuplicateItem {
		fmt.Println("An error occurred trying to add secret to keychain.")
		fmt.Printf("Account %s was already found. Cannot add duplicate secret. Exiting...\n", account)
		os.Exit(1)
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
		fmt.Println("An error occurred trying to remove secret from keychain.")
		fmt.Printf("Account %s could not be found in keychain. Exiting...\n", account)
		os.Exit(1)
	}

	fmt.Printf("Removed %s successfully from keychain.\n", account)
	return
}

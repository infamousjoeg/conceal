package wincred

import (
	"fmt"
	"log"

	"github.com/danieljoos/wincred"
	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
)

// SecretExists is a boolean function to verify a secret is present in Windows Credential Manager
func SecretExists(secretID string) bool {
	allsecretIDs, err := wincred.List()
	if err != nil {
		fmt.Println(err)
	}

	// Search all the available secretID TargetNames for this one
	for i := range allsecretIDs {
		if allsecretIDs[i].TargetName == secretID {
			return true
		}
	}
	return false
}

// ListSecrets is a string array function that returns all secrets in Windows Credetial Manager
func ListSecrets() []*wincred.Credential {
	secretIDs, err := wincred.List()
	if err != nil {
		log.Fatalln(err)
	}

	return secretIDs
}

// AddSecret is a non-return function that adds the secret and secret value to
// Windows Credential Manager
func AddSecret(secretID string, secret []byte) {
	// Add new generic credential to Windows Credential Manager
	item := wincred.NewGenericCredential(secretID)
	item.CredentialBlob = secret

	err := item.Write()

	// Unexpected error
	if err != nil {
		log.Fatalf(
			"An unexpected error occurred trying to add a secret to "+
				"the keychain:\n%s\nExiting...",
			err,
		)
	}

	// Verify the secret was set in Windows Credential Manager successfully
	if !SecretExists(secretID) {
		log.Fatalf("Secret %s not found in Windows Credential Manager. Exiting...\n", secret)
	}

	fmt.Printf("Added %s successfully to Windows Credential Manager.\n", secretID)
}

// DeleteSecret is a non-return function that removes the secret from Windows Credential Manager.
func DeleteSecret(secretID string) {
	cred, err := wincred.GetGenericCredential(secretID)
	if err != nil {
		log.Fatalf(
			"An error occurred trying to remove secret from "+
				"keychain.\n  Secret '%s' not found in keychain. Exiting...\n",
			secretID,
		)
	}

	cred.Delete()

	fmt.Printf("Removed %s successfully from Windows Credential Manager.\n", secretID)
}

// GetSecret is a non-return function that retrieves a secret and immediately
// adds it to the host clipboard for 15 seconds.
func GetSecret(secretID string) {
	results, err := wincred.GetGenericCredential(secretID)
	if err != nil {
		// Error occurred
		log.Fatalf(
			"An error occurred trying to get secret from " +
				"Windows Credential Manager.\n Exiting...\n",
		)
	} else {
		password := string(results.CredentialBlob)
		clipboard.Secret(password)
	}
}

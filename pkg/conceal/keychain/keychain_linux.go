//go:build linux

package keychain

import (
	"fmt"

	libsecret "github.com/gsterjov/go-libsecret"
	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
)

func getCollection() (*libsecret.Service, *libsecret.Collection, *libsecret.Session, error) {
	svc, err := libsecret.NewService()
	if err != nil {
		return nil, nil, nil, err
	}
	session, err := svc.Open()
	if err != nil {
		return nil, nil, nil, err
	}
	cols, err := svc.Collections()
	if err != nil || len(cols) == 0 {
		return nil, nil, nil, fmt.Errorf("no credential collections available")
	}
	col := &cols[0]
	locked, _ := col.Locked()
	if locked {
		if err := svc.Unlock(col); err != nil {
			return nil, nil, nil, err
		}
	}
	return svc, col, session, nil
}

// SecretExists checks if the secretID exists in the collection
func SecretExists(secretID string) bool {
	_, col, _, err := getCollection()
	if err != nil {
		return false
	}
	items, err := col.SearchItems(secretID)
	return err == nil && len(items) > 0
}

// ListSecrets returns all secrets stored by conceal
func ListSecrets() []string {
	_, col, _, err := getCollection()
	if err != nil {
		return []string{}
	}
	items, err := col.Items()
	if err != nil {
		return []string{}
	}
	var results []string
	for _, it := range items {
		lbl, err := it.Label()
		if err == nil {
			results = append(results, lbl)
		}
	}
	return results
}

// AddSecret adds a secret to libsecret
func AddSecret(secretID string, secret []byte) error {
	_, col, session, err := getCollection()
	if err != nil {
		return fmt.Errorf("An unexpected error occurred trying to add secret %s to the credential store", secretID)
	}
	if SecretExists(secretID) {
		return fmt.Errorf("Secret %s already exists in credential store. Please use `conceal update` instead.", secretID)
	}
	sec := libsecret.NewSecret(session, nil, secret, "text/plain")
	if _, err := col.CreateItem(secretID, sec, false); err != nil {
		return fmt.Errorf("An unexpected error occurred trying to add secret %s to the credential store", secretID)
	}
	return nil
}

// DeleteSecret removes a secret from libsecret
func DeleteSecret(secretID string) error {
	_, col, _, err := getCollection()
	if err != nil {
		return fmt.Errorf("An error occurred trying to remove secret from credential store. Secret '%s' not found.", secretID)
	}
	items, err := col.SearchItems(secretID)
	if err != nil || len(items) == 0 {
		return fmt.Errorf("An error occurred trying to remove secret from credential store. Secret '%s' not found.", secretID)
	}
	if err := items[0].Delete(); err != nil {
		return fmt.Errorf("An error occurred trying to remove secret from credential store. Secret '%s' not found.", secretID)
	}
	return nil
}

// GetSecret retrieves a secret and delivers it via clipboard or stdout
func GetSecret(secretID string, delivery string) error {
	_, col, session, err := getCollection()
	if err != nil {
		return fmt.Errorf("An error occurred trying to get secret from credential store. Secret '%s' not found.", secretID)
	}
	items, err := col.SearchItems(secretID)
	if err != nil || len(items) == 0 {
		return fmt.Errorf("An error occurred trying to get secret from credential store. Secret '%s' not found.", secretID)
	}
	sec, err := items[0].GetSecret(session)
	if err != nil {
		return fmt.Errorf("An error occurred trying to get secret from credential store. Secret '%s' not found.", secretID)
	}
	password := string(sec.Value)
	if delivery == "clipboard" {
		clipboard.Secret(password)
	} else if delivery == "stdout" {
		fmt.Printf("%s", password)
	}
	password = ""
	return nil
}

// UpdateSecret updates an existing secret in libsecret
func UpdateSecret(secretID string, secret []byte) error {
	_, col, session, err := getCollection()
	if err != nil {
		return fmt.Errorf("An unexpected error occurred trying to update secret %s in the credential store", secretID)
	}
	items, err := col.SearchItems(secretID)
	if err != nil || len(items) == 0 {
		return fmt.Errorf("The secret %s does not exist in the credential store. Please use `conceal set` instead.", secretID)
	}
	sec := libsecret.NewSecret(session, nil, secret, "text/plain")
	if _, err := col.CreateItem(secretID, sec, true); err != nil {
		return fmt.Errorf("An unexpected error occurred trying to update secret %s in the credential store", secretID)
	}
	return nil
}

//go:build darwin

package keychain

import (
	"fmt"
	"github.com/keybase/go-keychain"
)

func queryFor(id string) keychain.Item {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService("summon")
	if id != "" {
		item.SetAccount(id)
	}
	return item
}

func secretExists(id string) bool {
	query := queryFor(id)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnAttributes(true)
	results, err := keychain.QueryItem(query)
	return err == nil && len(results) == 1
}

func listSecrets() ([]QueryResult, error) {
	query := queryFor("")
	query.SetMatchLimit(keychain.MatchLimitAll)
	query.SetReturnAttributes(true)
	items, err := keychain.QueryItem(query)
	if err != nil && err != keychain.ErrorItemNotFound {
		return nil, fmt.Errorf("list keychain: %w", err)
	}
	results := make([]QueryResult, 0, len(items))
	for _, item := range items {
		results = append(results, QueryResult{Account: item.Account})
	}
	return results, nil
}

func addSecret(id string, value []byte) error {
	item := queryFor(id)
	item.SetLabel("summon")
	item.SetData(value)
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleAfterFirstUnlock)
	if err := keychain.AddItem(item); err != nil {
		return fmt.Errorf("add secret %q: %w", id, err)
	}
	return nil
}

func deleteSecret(id string) error {
	if err := keychain.DeleteItem(queryFor(id)); err != nil {
		return fmt.Errorf("delete secret %q: %w", id, err)
	}
	return nil
}

func readSecret(id string) ([]byte, error) {
	query := queryFor(id)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)
	items, err := keychain.QueryItem(query)
	if err != nil {
		return nil, fmt.Errorf("read secret %q: %w", id, err)
	}
	if len(items) != 1 {
		return nil, fmt.Errorf("secret %q not found", id)
	}
	return items[0].Data, nil
}

func updateSecret(id string, value []byte) error {
	attributes := keychain.NewItem()
	attributes.SetData(value)
	if err := keychain.UpdateItem(queryFor(id), attributes); err != nil {
		return fmt.Errorf("update secret %q: %w", id, err)
	}
	return nil
}

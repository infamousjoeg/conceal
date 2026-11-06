//go:build windows

package keychain

import "github.com/infamousjoeg/conceal/pkg/conceal/wincred"

func secretExists(id string) bool { return wincred.SecretExists(id) }

// List returns secret metadata, or an error if the store cannot be read.
func listSecrets() ([]QueryResult, error) {
	secrets, err := wincred.ListSecrets()
	if err != nil {
		return nil, err
	}
	results := make([]QueryResult, 0, len(secrets))
	for _, secret := range secrets {
		results = append(results, QueryResult{Account: secret.Account})
	}
	return results, nil
}

func addSecret(id string, value []byte) error    { return wincred.AddSecret(id, value) }
func deleteSecret(id string) error               { return wincred.DeleteSecret(id) }
func readSecret(id string) ([]byte, error)       { return wincred.GetSecret(id, "") }
func updateSecret(id string, value []byte) error { return wincred.UpdateSecret(id, value) }

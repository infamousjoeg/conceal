//go:build windows

package wincred

import (
	"errors"
	"fmt"
	"strings"

	"github.com/danieljoos/wincred"
)

func SecretExists(id string) bool {
	credential, err := wincred.GetGenericCredential(getTargetName(id))
	if credential != nil {
		clear(credential.CredentialBlob)
	}
	return err == nil && credential != nil
}

// ListSecrets enumerates only Conceal's namespace. Windows returns complete
// records; discard their blobs immediately and return names only.
func ListSecrets() ([]SecretInfo, error) {
	credentials, err := wincred.FilteredList("summon/*")
	if err != nil {
		return nil, fmt.Errorf("list credentials: %w", err)
	}
	result := make([]SecretInfo, 0, len(credentials))
	for _, credential := range credentials {
		clear(credential.CredentialBlob)
		if id, ok := strings.CutPrefix(credential.TargetName, "summon/"); ok && id != "" {
			result = append(result, SecretInfo{Account: id})
		}
	}
	return result, nil
}

func AddSecret(id string, value []byte) error {
	existing, err := wincred.GetGenericCredential(getTargetName(id))
	if existing != nil {
		clear(existing.CredentialBlob)
	}
	if err == nil {
		return fmt.Errorf("secret %q already exists; use conceal update", id)
	}
	if !errors.Is(err, wincred.ErrElementNotFound) {
		return fmt.Errorf("check secret %q: %w", id, err)
	}
	credential := wincred.NewGenericCredential(getTargetName(id))
	credential.CredentialBlob = value
	credential.Comment = "Summon secret managed by Conceal"
	if err := credential.Write(); err != nil {
		return fmt.Errorf("add secret %q: %w", id, err)
	}
	return nil
}

func DeleteSecret(id string) error {
	credential := wincred.NewGenericCredential(getTargetName(id))
	if err := credential.Delete(); err != nil {
		return fmt.Errorf("delete secret %q: %w", id, err)
	}
	return nil
}

func GetSecret(id string, _ string) ([]byte, error) {
	credential, err := wincred.GetGenericCredential(getTargetName(id))
	if err != nil {
		return nil, fmt.Errorf("read secret %q: %w", id, err)
	}
	return credential.CredentialBlob, nil
}

func UpdateSecret(id string, value []byte) error {
	credential, err := wincred.GetGenericCredential(getTargetName(id))
	if err != nil {
		return fmt.Errorf("find secret %q: %w", id, err)
	}
	clear(credential.CredentialBlob)
	credential.CredentialBlob = value
	if err := credential.Write(); err != nil {
		return fmt.Errorf("update secret %q: %w", id, err)
	}
	return nil
}

type SecretInfo struct{ Account string }

func getTargetName(id string) string { return "summon/" + id }

//go:build linux
// +build linux

package keychain

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/infamousjoeg/conceal/pkg/conceal/clipboard"
	"golang.org/x/sys/unix"
)

// Linux kernel keyring constants
const (
	keyringService = "summon"
	keyDescPrefix  = "summon:"
)

// QueryResult represents a query result for cross-platform compatibility
type QueryResult struct {
	Account string
}

// getSessionKeyring returns the user's session keyring ID
func getSessionKeyring() (int, error) {
	keyring, err := unix.KeyctlGetKeyringID(unix.KEY_SPEC_USER_KEYRING, true)
	if err != nil {
		return 0, fmt.Errorf("failed to get user keyring: %w", err)
	}
	return keyring, nil
}

// keyDescription builds the key description from secretID
func keyDescription(secretID string) string {
	return keyDescPrefix + secretID
}

// SecretExists checks if a secret exists in the Linux kernel keyring
func SecretExists(secretID string) bool {
	keyring, err := getSessionKeyring()
	if err != nil {
		return false
	}

	desc := keyDescription(secretID)
	_, err = unix.KeyctlSearch(keyring, "user", desc, 0)
	return err == nil
}

// ListSecrets returns all secrets in the keyring with the summon prefix
func ListSecrets() []QueryResult {
	keyring, err := getSessionKeyring()
	if err != nil {
		return []QueryResult{}
	}

	// Read all key IDs from the keyring
	// First call with nil buffer to get size
	size, err := unix.KeyctlBuffer(unix.KEYCTL_READ, keyring, nil, 0)
	if err != nil || size == 0 {
		return []QueryResult{}
	}

	// Allocate buffer and read key IDs
	buf := make([]byte, size)
	_, err = unix.KeyctlBuffer(unix.KEYCTL_READ, keyring, buf, 0)
	if err != nil {
		return []QueryResult{}
	}

	// Parse key IDs (4 bytes each, little endian)
	results := []QueryResult{}
	for i := 0; i+4 <= len(buf); i += 4 {
		keyID := int(*(*int32)(unsafe.Pointer(&buf[i])))
		if keyID == 0 {
			continue
		}

		// Get key description
		descBuf := make([]byte, 256)
		n, err := unix.KeyctlBuffer(unix.KEYCTL_DESCRIBE, keyID, descBuf, 0)
		if err != nil || n == 0 {
			continue
		}

		// Description format: "type;uid;gid;perm;description"
		desc := string(descBuf[:n-1]) // trim null terminator
		parts := strings.SplitN(desc, ";", 5)
		if len(parts) < 5 {
			continue
		}

		keyType := parts[0]
		keyDesc := parts[4]

		// Only include user keys with our prefix
		if keyType == "user" && strings.HasPrefix(keyDesc, keyDescPrefix) {
			secretID := strings.TrimPrefix(keyDesc, keyDescPrefix)
			results = append(results, QueryResult{Account: secretID})
		}
	}

	return results
}

// AddSecret adds a secret to the Linux kernel keyring
func AddSecret(secretID string, secret []byte) error {
	if SecretExists(secretID) {
		return fmt.Errorf("secret %s already exists in keyring, please use `conceal update` instead", secretID)
	}

	keyring, err := getSessionKeyring()
	if err != nil {
		return err
	}

	desc := keyDescription(secretID)
	_, err = unix.AddKey("user", desc, secret, keyring)
	if err != nil {
		return fmt.Errorf("failed to add secret %s to keyring: %w", secretID, err)
	}

	// Verify the secret was added
	if !SecretExists(secretID) {
		return fmt.Errorf("secret %s was set but is not found in keyring", secretID)
	}

	return nil
}

// DeleteSecret removes a secret from the Linux kernel keyring
func DeleteSecret(secretID string) error {
	keyring, err := getSessionKeyring()
	if err != nil {
		return err
	}

	desc := keyDescription(secretID)
	keyID, err := unix.KeyctlSearch(keyring, "user", desc, 0)
	if err != nil {
		return fmt.Errorf("secret '%s' not found in keyring", secretID)
	}

	// Invalidate/unlink the key
	_, err = unix.KeyctlInt(unix.KEYCTL_INVALIDATE, keyID, 0, 0, 0)
	if err != nil {
		// Fallback to unlink if invalidate not supported
		_, err = unix.KeyctlInt(unix.KEYCTL_UNLINK, keyID, keyring, 0, 0)
		if err != nil {
			return fmt.Errorf("failed to delete secret '%s' from keyring: %w", secretID, err)
		}
	}

	return nil
}

// GetSecret retrieves a secret and delivers it via clipboard or stdout
func GetSecret(secretID string, delivery string) error {
	keyring, err := getSessionKeyring()
	if err != nil {
		return err
	}

	desc := keyDescription(secretID)
	keyID, err := unix.KeyctlSearch(keyring, "user", desc, 0)
	if err != nil {
		return fmt.Errorf("secret '%s' not found in keyring", secretID)
	}

	// Get the secret data size first
	size, err := unix.KeyctlBuffer(unix.KEYCTL_READ, keyID, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to read secret '%s' from keyring: %w", secretID, err)
	}

	// Read the secret data
	buf := make([]byte, size)
	_, err = unix.KeyctlBuffer(unix.KEYCTL_READ, keyID, buf, 0)
	if err != nil {
		return fmt.Errorf("failed to read secret '%s' from keyring: %w", secretID, err)
	}

	password := string(buf)
	switch delivery {
	case "clipboard":
		clipboard.Secret(password)
	case "stdout":
		fmt.Printf("%s", password)
	}

	return nil
}

// UpdateSecret updates an existing secret in the Linux kernel keyring
func UpdateSecret(secretID string, secret []byte) error {
	keyring, err := getSessionKeyring()
	if err != nil {
		return err
	}

	desc := keyDescription(secretID)
	keyID, err := unix.KeyctlSearch(keyring, "user", desc, 0)
	if err != nil {
		return fmt.Errorf("secret %s does not exist in keyring, please use `conceal set` instead", secretID)
	}

	// Update the key payload
	// Ensure we have write permission (ignore error, try anyway)
	_ = unix.KeyctlSetperm(keyID, 0x3f3f0000)

	// Linux keyring doesn't have a direct update; we delete and re-add
	_, err = unix.KeyctlInt(unix.KEYCTL_INVALIDATE, keyID, 0, 0, 0)
	if err != nil {
		_, err = unix.KeyctlInt(unix.KEYCTL_UNLINK, keyID, keyring, 0, 0)
		if err != nil {
			return fmt.Errorf("failed to update secret '%s': could not remove old value: %w", secretID, err)
		}
	}

	// Add the new value
	_, err = unix.AddKey("user", desc, secret, keyring)
	if err != nil {
		return fmt.Errorf("failed to update secret %s in keyring: %w", secretID, err)
	}

	return nil
}

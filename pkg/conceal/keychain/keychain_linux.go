//go:build linux

package keychain

import (
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"strings"
)

const keyDescPrefix = "summon:"

func userKeyring() (int, error) {
	id, err := unix.KeyctlGetKeyringID(unix.KEY_SPEC_USER_KEYRING, true)
	if err != nil {
		return 0, fmt.Errorf("user keyring unavailable on Linux (check session permissions and container restrictions): %w", err)
	}
	return id, nil
}

func findKey(id string) (int, error) {
	ring, err := userKeyring()
	if err != nil {
		return 0, err
	}
	key, err := unix.KeyctlSearch(ring, "user", keyDescPrefix+id, 0)
	if err != nil {
		return 0, fmt.Errorf("find secret %q: %w", id, err)
	}
	return key, nil
}

func secretExists(id string) bool { _, err := findKey(id); return err == nil }

// readKeyringBuffer retries when the kernel reports that a concurrently changed
// keyring or payload no longer fits the supplied buffer.
func readKeyringBuffer(command, id int) ([]byte, error) {
	var buffer []byte
	for attempt := 0; attempt < 8; attempt++ {
		n, err := unix.KeyctlBuffer(command, id, buffer, 0)
		if err != nil {
			return nil, err
		}
		if n <= len(buffer) {
			return buffer[:n], nil
		}
		buffer = make([]byte, n)
	}
	return nil, errors.New("keyring changed repeatedly while reading; retry the operation")
}

func listSecrets() ([]QueryResult, error) {
	ring, err := userKeyring()
	if err != nil {
		return nil, err
	}
	data, err := readKeyringBuffer(unix.KEYCTL_READ, ring)
	if err != nil {
		return nil, fmt.Errorf("list keyring: %w", err)
	}
	results := []QueryResult{}
	for i := 0; i+4 <= len(data); i += 4 {
		id := int(int32(binary.NativeEndian.Uint32(data[i : i+4])))
		description, err := unix.KeyctlString(unix.KEYCTL_DESCRIBE, id)
		if errors.Is(err, unix.ENOKEY) || errors.Is(err, unix.EKEYREVOKED) || errors.Is(err, unix.EKEYEXPIRED) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("describe keyring entry: %w", err)
		}
		parts := strings.SplitN(strings.TrimRight(description, "\x00"), ";", 5)
		if len(parts) == 5 && parts[0] == "user" && strings.HasPrefix(parts[4], keyDescPrefix) {
			results = append(results, QueryResult{Account: strings.TrimPrefix(parts[4], keyDescPrefix)})
		}
	}
	return results, nil
}

func addSecret(id string, value []byte) error {
	ring, err := userKeyring()
	if err != nil {
		return err
	}
	_, err = unix.KeyctlSearch(ring, "user", keyDescPrefix+id, 0)
	if err == nil {
		return fmt.Errorf("secret %q already exists; use conceal update", id)
	}
	if !errors.Is(err, unix.ENOKEY) {
		return fmt.Errorf("check secret %q: %w", id, err)
	}
	_, err = unix.AddKey("user", keyDescPrefix+id, value, ring)
	if err != nil {
		return fmt.Errorf("add secret %q: %w", id, err)
	}
	return nil
}

func deleteSecret(id string) error {
	key, err := findKey(id)
	if err != nil {
		return err
	}
	_, err = unix.KeyctlInt(unix.KEYCTL_INVALIDATE, key, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("delete secret %q: %w", id, err)
	}
	return nil
}

func readSecret(id string) ([]byte, error) {
	key, err := findKey(id)
	if err != nil {
		return nil, err
	}
	value, err := readKeyringBuffer(unix.KEYCTL_READ, key)
	if err != nil {
		return nil, fmt.Errorf("read secret %q: %w", id, err)
	}
	return value, nil
}

func updateSecret(id string, value []byte) error {
	key, err := findKey(id)
	if err != nil {
		return err
	}
	// KEYCTL_UPDATE preserves the old value if the update fails.
	_, err = unix.KeyctlBuffer(unix.KEYCTL_UPDATE, key, value, 0)
	if err != nil {
		return fmt.Errorf("update secret %q: %w", id, err)
	}
	return nil
}

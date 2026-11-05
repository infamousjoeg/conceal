package keychain

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"os"
	"runtime"
	"strings"
	"testing"
)

// Native tests are explicit because they write to the operating system's store.
// They fail, rather than skip, when an opted-in backend is unavailable.
func TestNativeSecretLifecycle(t *testing.T) {
	if os.Getenv("CONCEAL_INTEGRATION") != "1" {
		t.Skip("set CONCEAL_INTEGRATION=1 to test the native credential store")
	}
	entropy := make([]byte, 16)
	if _, err := rand.Read(entropy); err != nil {
		t.Fatal(err)
	}
	id := "conceal-integration/" + hex.EncodeToString(entropy)
	// A long account exercises dynamically sized Linux key descriptions.
	id += "/" + strings.Repeat("x", 300)
	if SecretExists(id) {
		t.Fatal("unexpected integration name collision")
	}
	original := []byte("fixture-only\x00value\nwith whitespace \n")
	if err := AddSecret(id, original); err != nil {
		t.Fatalf("native add: %v", err)
	}
	t.Cleanup(func() { _ = DeleteSecret(id) })
	if !SecretExists(id) {
		t.Fatal("new secret does not exist")
	}
	assertValue := func(want []byte) {
		t.Helper()
		actual, err := ReadSecret(id)
		if err != nil {
			t.Fatalf("native read: %v", err)
		}
		defer clear(actual)
		if !bytes.Equal(actual, want) {
			t.Fatal("stored bytes differ from input")
		}
	}
	assertValue(original)
	if err := AddSecret(id, []byte("replacement")); err == nil {
		t.Fatal("duplicate add succeeded")
	}
	assertValue(original)
	items, err := List()
	if err != nil {
		t.Fatalf("native list: %v", err)
	}
	found := 0
	for _, item := range items {
		if item.Account == id {
			found++
		}
	}
	if found != 1 {
		t.Fatalf("own integration entry appears %d times", found)
	}
	updated := []byte("updated fixture\r\n")
	if err := UpdateSecret(id, updated); err != nil {
		t.Fatalf("native update: %v", err)
	}
	assertValue(updated)
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		if err := UpdateSecret(id, bytes.Repeat([]byte("x"), 1024*1024)); err == nil {
			t.Fatal("oversized update unexpectedly succeeded")
		}
		assertValue(updated)
	}
	if err := DeleteSecret(id); err != nil {
		t.Fatalf("native delete: %v", err)
	}
	if SecretExists(id) {
		t.Fatal("deleted secret still exists")
	}
	if _, err := ReadSecret(id); err == nil {
		t.Fatal("read of missing secret succeeded")
	}
	if err := UpdateSecret(id, original); err == nil {
		t.Fatal("update created a missing secret")
	}
	if err := DeleteSecret(id); err == nil {
		t.Fatal("delete of missing secret succeeded")
	}
}

func TestInvalidDelivery(t *testing.T) {
	if err := GetSecret("not-accessed", "invalid"); err == nil {
		t.Fatal("unsupported delivery accepted")
	}
}

func TestInvalidNamesNeverReachNativeStore(t *testing.T) {
	for _, id := range []string{"", "  ", "bad\nname", "bad\rname", "bad\x00name"} {
		if SecretExists(id) {
			t.Errorf("invalid name %q exists", id)
		}
		if err := AddSecret(id, []byte("fixture")); err == nil {
			t.Errorf("invalid add %q succeeded", id)
		}
		if err := UpdateSecret(id, []byte("fixture")); err == nil {
			t.Errorf("invalid update %q succeeded", id)
		}
		if err := DeleteSecret(id); err == nil {
			t.Errorf("invalid delete %q succeeded", id)
		}
		if _, err := ReadSecret(id); err == nil {
			t.Errorf("invalid read %q succeeded", id)
		}
	}
}

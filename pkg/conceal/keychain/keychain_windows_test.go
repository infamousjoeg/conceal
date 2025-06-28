//go:build windows

package keychain

import "testing"

func TestAddGetDeleteWindows(t *testing.T) {
	secretID := "conceal_test_secret"
	_ = DeleteSecret(secretID)
	if err := AddSecret(secretID, []byte("val")); err != nil {
		t.Fatal(err)
	}
	if !SecretExists(secretID) {
		t.Fatal("secret should exist")
	}
	if err := GetSecret(secretID, "stdout"); err != nil {
		t.Fatalf("get error: %v", err)
	}
	if err := UpdateSecret(secretID, []byte("new")); err != nil {
		t.Fatalf("update error: %v", err)
	}
	if err := DeleteSecret(secretID); err != nil {
		t.Fatalf("delete error: %v", err)
	}
}

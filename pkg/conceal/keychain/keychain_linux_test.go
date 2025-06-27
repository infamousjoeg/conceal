//go:build linux

package keychain

import (
	"strings"
	"testing"
)

func TestSecretExistsNoService(t *testing.T) {
	if SecretExists("does/not/exist") {
		t.Fatal("expected false when service unavailable")
	}
}

func TestListSecretsNoService(t *testing.T) {
	if secrets := ListSecrets(); len(secrets) != 0 {
		t.Fatalf("expected no secrets, got %v", secrets)
	}
}

func TestAddSecretNoService(t *testing.T) {
	err := AddSecret("test/key", []byte("value"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "credential store") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetSecretNoService(t *testing.T) {
	if err := GetSecret("test/key", "stdout"); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateSecretNoService(t *testing.T) {
	err := UpdateSecret("test/key", []byte("value"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteSecretNoService(t *testing.T) {
	err := DeleteSecret("test/key")
	if err == nil {
		t.Fatal("expected error")
	}
}

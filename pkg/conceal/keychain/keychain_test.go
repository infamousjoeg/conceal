package keychain

import (
	"testing"
)

func TestAddSecret(t *testing.T) {
	// Test case where secret does not exist
	err := AddSecret("secret1", []byte("password1"))
	if err != nil {
		t.Errorf("Expected AddSecret to return nil error, but got: %v", err)
	}

	// Test case where secret already exists
	err = AddSecret("secret1", []byte("password1"))
	if err == nil {
		t.Errorf("Expected AddSecret to return non-nil error, but got nil")
	}
}

func TestSecretExists(t *testing.T) {
	// Test case where secret exists
	exists := SecretExists("secret1")
	if !exists {
		t.Errorf("Expected SecretExists to return true, but got false")
	}

	// Test case where secret does not exist
	exists = SecretExists("secret3")
	if exists {
		t.Errorf("Expected SecretExists to return false, but got true")
	}
}

func TestListSecrets(t *testing.T) {
	// Test case where secrets exist
	secrets := ListSecrets()
	if len(secrets) == 0 {
		t.Errorf("Expected ListSecrets to return non-empty list, but got empty list")
	}
}

func TestGetSecret(t *testing.T) {
	// Test case where secret exists and delivery is clipboard
	err := GetSecret("secret1", "clipboard")
	if err != nil {
		t.Errorf("Expected GetSecret to return nil error, but got: %v", err)
	}

	// Test case where secret exists and delivery is stdout
	err = GetSecret("secret1", "stdout")
	if err != nil {
		t.Errorf("Expected GetSecret to return nil error, but got: %v", err)
	}

	// Test case where secret does not exist and delivery is clipboard
	err = GetSecret("secret3", "clipboard")
	if err == nil {
		t.Errorf("Expected GetSecret to return non-nil error, but got nil")
	}

	// Test case where secret does not exist and delivery is stdout
	err = GetSecret("secret3", "stdout")
	if err == nil {
		t.Errorf("Expected GetSecret to return non-nil error, but got nil")
	}
}

func TestUpdateSecret(t *testing.T) {
	// Test case where secret exists
	err := UpdateSecret("secret1", []byte("newpassword"))
	if err != nil {
		t.Errorf("Expected UpdateSecret to return nil error, but got: %v", err)
	}

	// Test case where secret does not exist
	err = UpdateSecret("secret3", []byte("password3"))
	if err == nil {
		t.Errorf("Expected UpdateSecret to return non-nil error, but got nil")
	}
}

func TestDeleteSecret(t *testing.T) {
	// Test case where secret exists
	err := DeleteSecret("secret1")
	if err != nil {
		t.Errorf("Expected DeleteSecret to return nil error, but got: %v", err)
	}

	// Test case where secret does not exist
	err = DeleteSecret("secret3")
	if err == nil {
		t.Errorf("Expected DeleteSecret to return non-nil error, but got nil")
	}
}

func TestListSecretsNone(t *testing.T) {
	// Test case where no secrets exist
	// (Assuming there are no secrets in the keychain)
	secrets := ListSecrets()
	if len(secrets) != 0 {
		t.Errorf("Expected ListSecrets to return empty list, but got non-empty list")
	}
}

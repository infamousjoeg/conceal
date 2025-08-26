package keychain

import (
	"runtime"
	"strings"
	"testing"
)

// isSupported returns true if the current platform supports secret management
func isSupported() bool {
	platform := runtime.GOOS
	return platform == "darwin" || platform == "windows"
}

func TestAddSecret(t *testing.T) {
	if !isSupported() {
		// On unsupported platforms, operations should return errors
		err := AddSecret("secret1", []byte("password1"))
		if err == nil {
			t.Errorf("Expected AddSecret to return error on unsupported platform, but got nil")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("Expected error message to contain 'not supported', but got: %v", err)
		}
		return
	}

	// Test case where secret does not exist
	err := AddSecret("test_secret_add", []byte("password1"))
	if err != nil {
		t.Errorf("Expected AddSecret to return nil error, but got: %v", err)
	}

	// Test case where secret already exists
	err = AddSecret("test_secret_add", []byte("password1"))
	if err == nil {
		t.Errorf("Expected AddSecret to return non-nil error, but got nil")
	}

	// Cleanup
	DeleteSecret("test_secret_add")
}

func TestSecretExists(t *testing.T) {
	if !isSupported() {
		// On unsupported platforms, secrets don't exist
		exists := SecretExists("secret1")
		if exists {
			t.Errorf("Expected SecretExists to return false on unsupported platform, but got true")
		}
		return
	}

	// Add a test secret first
	AddSecret("test_secret_exists", []byte("password1"))

	// Test case where secret exists
	exists := SecretExists("test_secret_exists")
	if !exists {
		t.Errorf("Expected SecretExists to return true, but got false")
	}

	// Test case where secret does not exist
	exists = SecretExists("non_existent_secret")
	if exists {
		t.Errorf("Expected SecretExists to return false, but got true")
	}

	// Cleanup
	DeleteSecret("test_secret_exists")
}

func TestListSecrets(t *testing.T) {
	if !isSupported() {
		// On unsupported platforms, list should be empty
		secrets := ListSecrets()
		if len(secrets) != 0 {
			t.Errorf("Expected ListSecrets to return empty list on unsupported platform, but got %d secrets", len(secrets))
		}
		return
	}

	// Add a test secret first
	AddSecret("test_secret_list", []byte("password1"))

	// Test case where secrets exist
	secrets := ListSecrets()
	if len(secrets) == 0 {
		t.Errorf("Expected ListSecrets to return non-empty list, but got empty list")
	}

	// Cleanup
	DeleteSecret("test_secret_list")
}

func TestGetSecret(t *testing.T) {
	if !isSupported() {
		// On unsupported platforms, operations should return errors
		err := GetSecret("secret1", "clipboard")
		if err == nil {
			t.Errorf("Expected GetSecret to return error on unsupported platform, but got nil")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("Expected error message to contain 'not supported', but got: %v", err)
		}

		err = GetSecret("secret1", "stdout")
		if err == nil {
			t.Errorf("Expected GetSecret to return error on unsupported platform, but got nil")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("Expected error message to contain 'not supported', but got: %v", err)
		}
		return
	}

	// Add a test secret first
	AddSecret("test_secret_get", []byte("password1"))

	// Test case where secret exists and delivery is stdout (safer for CI)
	err := GetSecret("test_secret_get", "stdout")
	if err != nil {
		t.Errorf("Expected GetSecret to return nil error, but got: %v", err)
	}

	// Test case where secret does not exist and delivery is stdout
	err = GetSecret("non_existent_secret", "stdout")
	if err == nil {
		t.Errorf("Expected GetSecret to return non-nil error, but got nil")
	}

	// Cleanup
	DeleteSecret("test_secret_get")
}

func TestUpdateSecret(t *testing.T) {
	if !isSupported() {
		// On unsupported platforms, operations should return errors
		err := UpdateSecret("secret1", []byte("newpassword"))
		if err == nil {
			t.Errorf("Expected UpdateSecret to return error on unsupported platform, but got nil")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("Expected error message to contain 'not supported', but got: %v", err)
		}
		return
	}

	// Add a test secret first
	AddSecret("test_secret_update", []byte("password1"))

	// Test case where secret exists
	err := UpdateSecret("test_secret_update", []byte("newpassword"))
	if err != nil {
		t.Errorf("Expected UpdateSecret to return nil error, but got: %v", err)
	}

	// Test case where secret does not exist
	err = UpdateSecret("non_existent_secret", []byte("password3"))
	if err == nil {
		t.Errorf("Expected UpdateSecret to return non-nil error, but got nil")
	}

	// Cleanup
	DeleteSecret("test_secret_update")
}

func TestDeleteSecret(t *testing.T) {
	if !isSupported() {
		// On unsupported platforms, operations should return errors
		err := DeleteSecret("secret1")
		if err == nil {
			t.Errorf("Expected DeleteSecret to return error on unsupported platform, but got nil")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("Expected error message to contain 'not supported', but got: %v", err)
		}
		return
	}

	// Add a test secret first
	AddSecret("test_secret_delete", []byte("password1"))

	// Test case where secret exists
	err := DeleteSecret("test_secret_delete")
	if err != nil {
		t.Errorf("Expected DeleteSecret to return nil error, but got: %v", err)
	}

	// Test case where secret does not exist
	err = DeleteSecret("non_existent_secret")
	if err == nil {
		t.Errorf("Expected DeleteSecret to return non-nil error, but got nil")
	}
}

func TestListSecretsNone(t *testing.T) {
	if !isSupported() {
		// On unsupported platforms, list should be empty
		secrets := ListSecrets()
		if len(secrets) != 0 {
			t.Errorf("Expected ListSecrets to return empty list on unsupported platform, but got %d secrets", len(secrets))
		}
		return
	}

	// Test case where no summon secrets exist
	// Note: This test may fail if other tests leave secrets behind
	// or if there are existing summon secrets in the keychain
	secrets := ListSecrets()
	// We can't assume the keychain is empty, so we just check that the function works
	if secrets == nil {
		t.Errorf("Expected ListSecrets to return non-nil slice, but got nil")
	}
}

//go:build windows
// +build windows

package wincred

import (
	"testing"
)

func TestGetTargetName(t *testing.T) {
	secretID := "test/secret"
	expected := "summon/test/secret"
	result := getTargetName(secretID)

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestSecretInfo(t *testing.T) {
	info := SecretInfo{
		Account: "test/account",
	}

	if info.Account != "test/account" {
		t.Errorf("Expected test/account, got %s", info.Account)
	}
}

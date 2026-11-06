//go:build !darwin && !windows && !linux

package keychain

import "errors"

var errUnsupported = errors.New("secret management requires macOS, Windows, or Linux")

func secretExists(string) bool            { return false }
func listSecrets() ([]QueryResult, error) { return nil, errUnsupported }
func addSecret(string, []byte) error      { return errUnsupported }
func deleteSecret(string) error           { return errUnsupported }
func readSecret(string) ([]byte, error)   { return nil, errUnsupported }
func updateSecret(string, []byte) error   { return errUnsupported }

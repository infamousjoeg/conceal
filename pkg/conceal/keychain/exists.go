package keychain

// SecretExists is a boolean function to verify a secret is present in keychain
func SecretExists(secretID string) (bool, error) {
	allsecretIDs, err := ListSecrets()
	if err != nil {
		return false, err
	}

	// Search all the available secretIDs for this one
	for _, id := range allsecretIDs {
		if id == secretID {
			return true, nil
		}
	}
	return false, nil
}

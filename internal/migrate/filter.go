package migrate

import "path/filepath"

func Filter(secrets map[string][]byte, pattern string) map[string][]byte {
	if pattern == "" {
		return secrets
	}
	out := make(map[string][]byte)
	for k, v := range secrets {
		match, _ := filepath.Match(pattern, k)
		if match {
			out[k] = v
		}
	}
	return out
}

package platform

import (
	"runtime"
)

// GetPlatform returns the current operating system platform
func GetPlatform() string {
	return runtime.GOOS
}

// IsMacOS checks if the current platform is macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsWindows checks if the current platform is Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsSupported checks if the current platform supports secret management
func IsSupported() bool {
	return IsMacOS() || IsWindows()
}

// GetSecretStoreName returns the name of the secret store for the current platform
func GetSecretStoreName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain"
	case "windows":
		return "Windows Credential Manager"
	default:
		return "Unsupported Platform"
	}
}
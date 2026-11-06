package platform

import "runtime"

func GetPlatform() string { return runtime.GOOS }
func IsMacOS() bool       { return runtime.GOOS == "darwin" }
func IsWindows() bool     { return runtime.GOOS == "windows" }
func IsSupported() bool   { return IsMacOS() || IsWindows() || runtime.GOOS == "linux" }
func GetSecretStoreName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain"
	case "windows":
		return "Windows Credential Manager"
	case "linux":
		return "Linux kernel keyring"
	default:
		return "Unsupported platform"
	}
}

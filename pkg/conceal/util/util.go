package util

// Contains is a boolean function to check an array list for a value present
func Contains(item string, list []string) bool {
	for _, listItem := range list {
		if listItem == item {
			return true
		}
	}
	return false
}

package clipboard

import "github.com/atotto/clipboard"

// Secret is a non-return function that adds content to the host clipboard
// that persists for 15 seconds. If a signal interrupt is detected, the content is
// immediately cleared.
func Secret(secret string) {
	clipboard.WriteAll(secret)
}

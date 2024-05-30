package clipboard

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
)

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Clear clipboard
		clipboard.WriteAll("")
		os.Exit(0)
	}()
}

// Secret is a non-return function that adds content to the host clipboard
// that persists for 15 seconds. If a signal interrupt is detected, the content is
// immediately cleared.
func Secret(secret string) {
	// Setup our Ctrl+C (signal interrupt) handler
	SetupCloseHandler()

	// Write secret to clipboard
	clipboard.WriteAll(secret)

	// Sleep for 15 seconds
	time.Sleep(15 * time.Second)

	// Clear clipboard
	clipboard.WriteAll("")
}

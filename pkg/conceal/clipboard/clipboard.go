package clipboard

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
)

// these vars allow tests to stub clipboard and time operations
var (
	writeAll = clipboard.WriteAll
	sleep    = time.Sleep
	exitFunc = os.Exit
)

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := writeAll(""); err != nil {
			log.Printf("failed clearing clipboard: %v", err)
		}
		exitFunc(0)
	}()
}

// Secret is a non-return function that adds content to the host clipboard
// that persists for 15 seconds. If a signal interrupt is detected, the content is
// immediately cleared.
func Secret(secret string) {
	// Setup our Ctrl+C (signal interrupt) handler
	SetupCloseHandler()

	// Write secret to clipboard
	if err := writeAll(secret); err != nil {
		log.Printf("failed setting clipboard: %v", err)
	}

	// Sleep for 15 seconds
	sleep(15 * time.Second)

	// Clear clipboard
	if err := writeAll(""); err != nil {
		log.Printf("failed clearing clipboard: %v", err)
	}
}

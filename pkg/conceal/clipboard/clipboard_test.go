package clipboard

import (
	"os"
	"testing"
	"time"

	"github.com/atotto/clipboard"
)

func TestSecret(t *testing.T) {
	// Save the current clipboard content
	originalContent, _ := clipboard.ReadAll()

	// Set up a mock clipboard
	mockClipboard := ""

	// Replace the clipboard.WriteAll function with a mock implementation
	clipObject := clipboard.WriteAll
	clipObject = func(text string) error {
		mockClipboard = text
		return nil
	}

	// Call the Secret function with a secret message
	secret := "This is a secret message"
	go Secret(secret)

	// Wait for 15 seconds
	time.Sleep(15 * time.Second)

	// Check if the clipboard was cleared after 15 seconds
	if mockClipboard != "" {
		t.Errorf("Expected clipboard to be cleared, but got: %s", mockClipboard)
	}

	// Restore the original clipboard content
	clipObject(originalContent)
}

func TestSetupCloseHandler(t *testing.T) {
	// Save the current clipboard content
	originalContent, _ := clipboard.ReadAll()

	// Set up a mock clipboard
	mockClipboard := ""

	// Replace the clipboard.WriteAll function with a mock implementation
	clipObject := clipboard.WriteAll
	clipObject = func(text string) error {
		mockClipboard = text
		return nil
	}

	// Call the SetupCloseHandler function
	SetupCloseHandler()

	// Send an interrupt signal to the program
	signalChan := make(chan os.Signal, 1)
	signalChan <- os.Interrupt

	// Wait for a short time to allow the goroutine to handle the signal
	time.Sleep(100 * time.Millisecond)

	// Check if the clipboard was cleared after receiving the interrupt signal
	if mockClipboard != "" {
		t.Errorf("Expected clipboard to be cleared, but got: %s", mockClipboard)
	}

	// Restore the original clipboard content
	clipObject(originalContent)
}

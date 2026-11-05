package clipboard

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCopySecret(t *testing.T) {
	for _, test := range []struct {
		name, newer                                          string
		cancel, writeError, readError, clearError, wantError bool
	}{
		{name: "expires"},
		{name: "preserves newer clipboard", newer: "copied later"},
		{name: "clears on cancellation", cancel: true, wantError: true},
		{name: "copy failure", writeError: true, wantError: true},
		{name: "read failure", readError: true, wantError: true},
		{name: "clear failure", clearError: true, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			current := ""
			var writes []string
			read := func() (string, error) {
				if test.readError {
					return "", errors.New("read failed")
				}
				if test.newer != "" {
					current = test.newer
				}
				return current, nil
			}
			write := func(value string) error {
				writes = append(writes, value)
				if test.writeError || (value == "" && test.clearError) {
					return errors.New("write failed")
				}
				current = value
				if test.cancel {
					cancel()
				}
				return nil
			}
			err := copySecret(ctx, "test secret", time.Millisecond, read, write)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v", err)
			}
			if len(writes) == 0 || writes[0] != "test secret" {
				t.Fatalf("secret was not copied: %q", writes)
			}
			if test.newer != "" && (current != test.newer || len(writes) != 1) {
				t.Fatal("overwrote newer clipboard content")
			}
			if !test.writeError && !test.readError && !test.clearError && test.newer == "" && current != "" {
				t.Fatal("secret was not cleared")
			}
		})
	}
}

func TestCancelledCopyDoesNotTouchClipboard(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := copySecret(ctx, "secret", time.Hour, nil, func(string) error { t.Fatal("wrote after cancellation"); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}

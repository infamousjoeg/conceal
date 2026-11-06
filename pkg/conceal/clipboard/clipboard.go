package clipboard

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
)

// SetupCloseHandler is retained for source compatibility.
// Deprecated: Secret now manages its own signal handler and releases it on return.
func SetupCloseHandler() {}

// Secret copies a value for 15 seconds, or until interrupted. A newer clipboard
// value is left alone, and failures to copy or clear are returned to the caller.
func Secret(secret string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return copySecret(ctx, secret, 15*time.Second, clipboard.ReadAll, clipboard.WriteAll)
}

func copySecret(ctx context.Context, secret string, duration time.Duration, read func() (string, error), write func(string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := write(secret); err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
	current, err := read()
	if err != nil {
		return fmt.Errorf("read clipboard before clearing: %w", err)
	}
	if current == secret {
		if err := write(""); err != nil {
			return fmt.Errorf("clear clipboard: %w", err)
		}
	}
	return ctx.Err()
}

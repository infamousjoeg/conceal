package clipboard

import (
	"sync"
	"testing"
	"time"
)

func TestSecret(t *testing.T) {
	var mu sync.Mutex
	var calls []string
	writeAll = func(s string) error {
		mu.Lock()
		calls = append(calls, s)
		mu.Unlock()
		return nil
	}
	sleep = func(d time.Duration) {}
	exitFunc = func(code int) {}

	Secret("val")
	mu.Lock()
	defer mu.Unlock()
	if len(calls) != 2 || calls[0] != "val" || calls[1] != "" {
		t.Fatalf("unexpected calls %v", calls)
	}
}

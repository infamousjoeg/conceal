package migrate

import (
	"context"
	"io"
	"testing"
	"time"
)

type nopMigrator struct{}

func (nopMigrator) Name() string                                                     { return "nop" }
func (nopMigrator) Configure(ctx context.Context, in io.Reader, out io.Writer) error { return nil }
func (nopMigrator) DefaultRate() int                                                 { return 2 }
func (nopMigrator) Put(ctx context.Context, key string, v []byte, m map[string]string) error {
	return nil
}

func TestRateLimit(t *testing.T) {
	mgr := &Manager{}
	mig := nopMigrator{}
	start := time.Now()
	if err := mgr.MigrateWith(context.Background(), mig, map[string][]byte{"k": {}}, 2, false, io.Discard); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Fatalf("no rate limiting")
	}
}

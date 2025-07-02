package migrate

import (
	"context"
	"io"
	"os"
	"path/filepath"
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

func TestInstallAndList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	mgr := NewManager()

	pluginFile := filepath.Join(dir, "plug")
	if err := os.WriteFile(pluginFile, []byte("test"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Install(pluginFile); err != nil {
		t.Fatalf("install: %v", err)
	}

	names, err := mgr.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := "conceal-migrate-plug"
	if len(names) != 1 || names[0] != want {
		t.Fatalf("unexpected names %v", names)
	}

	info, err := os.Stat(filepath.Join(mgr.Dir, want))
	if err != nil {
		t.Fatalf("stat installed: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("wrong perm %v", info.Mode().Perm())
	}
}

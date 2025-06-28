package migrate

import (
	"context"
	"io"
)

// Migrator defines an external provider plugin.
type Migrator interface {
	Name() string
	Configure(ctx context.Context, in io.Reader, out io.Writer) error
	DefaultRate() int
	Put(ctx context.Context, key string, value []byte, meta map[string]string) error
}

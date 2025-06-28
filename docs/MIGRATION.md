# Secret Migration

Conceal can copy secrets from your local keyring to external providers using
plugins.
A plugin is a standalone executable named `conceal-migrate-<provider>` placed in
`$XDG_CONFIG_HOME/conceal/plugins`.

## Quickstart

1. Build or download a plugin and move it into the plugin directory.
2. Verify it is detected:
   ```bash
   conceal provider
   ```
3. Configure the provider (runs once and stores credentials securely):
   ```bash
   conceal configure aws-secretsmanager
   ```
4. Migrate secrets:
   ```bash
   conceal migrate --to aws-secretsmanager
   ```
   Use `--dry-run` to preview and `--selector` to copy only matching keys.

## Writing plugins

Plugins implement the following interface:

```go
type Migrator interface {
    Name() string
    Configure(ctx context.Context, in io.Reader, out io.Writer) error
    DefaultRate() int
    Put(ctx context.Context, key string, value []byte, meta map[string]string) error
}
```

Copy `plugins/template` as a starting point or study
`plugins/aws-secretsmanager` for a production-ready example. After building your
plugin, place the binary into the plugin directory and run `conceal provider`
again to confirm it loads.

# Secret Migration

Conceal can export your local secrets into external vaults using small plugin
executables. Each plugin follows a common interface so you can write your own or
use one provided by the community.

## Quickstart

1. Build or download a plugin and install it:
   ```bash
   conceal provider install ./conceal-migrate-aws-secretsmanager
   ```
2. List available providers:
   ```bash
   conceal provider
   ```
3. Configure the provider (only once):
   ```bash
   conceal configure aws-secretsmanager
   ```
4. Migrate secrets:
   ```bash
   conceal migrate --to aws-secretsmanager
   ```
   Add `--dry-run` to preview changes or `--selector` to choose specific keys.

## Writing Plugins

A plugin is a simple Go program that implements the `Migrator` interface. Start
by copying `plugins/template` and filling in the `Put` method. Build your binary
and place it in `$XDG_CONFIG_HOME/conceal/plugins`.

The interface:

```go
type Migrator interface {
    Name() string
    Configure(ctx context.Context, in io.Reader, out io.Writer) error
    DefaultRate() int
    Put(ctx context.Context, key string, value []byte, meta map[string]string) error
}
```

See `plugins/aws-secretsmanager` for a complete example.

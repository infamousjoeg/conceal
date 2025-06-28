# AWS Secrets Manager Migrator

This plugin uploads secrets from Conceal to AWS Secrets Manager.

## Build

```bash
go build -o conceal-migrate-aws-secretsmanager .
```

Place the resulting binary in your Conceal plugin directory (usually `$XDG_CONFIG_HOME/conceal/plugins`).

## Usage

Configure the plugin once:

```bash
conceal configure aws-secretsmanager
```

Then run a migration:

```bash
conceal migrate --to aws-secretsmanager
```

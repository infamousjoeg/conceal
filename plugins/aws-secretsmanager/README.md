# AWS Secrets Manager Migrator

This plugin uploads secrets from Conceal to AWS Secrets Manager.

## Build

```bash
cd plugins/aws-secretsmanager
go build -o conceal-migrate-aws-secretsmanager .
```

Copy the binary to your Conceal plugin directory (usually
`$XDG_CONFIG_HOME/conceal/plugins`).

## Usage

Configure once:

```bash
conceal configure aws-secretsmanager
```

Then run a migration:

```bash
conceal migrate --to aws-secretsmanager
```

# Secret Migration

Conceal can export secrets from the local keyring to external providers via plugins.

```
conceal migrate --to aws-secretsmanager
```

Use `conceal provider` to list installed plugins. Configure a provider once with
`conceal configure <name>`.

To develop a new plugin copy `plugins/template` and see `plugins/aws-secretsmanager` for a full example.

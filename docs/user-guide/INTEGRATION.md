# Summon Integration

[Summon](https://cyberark.github.io/summon) is a tool that injects secrets into
processes. Conceal can act as its provider.

```bash
conceal summon install
```

Then reference secrets in your Summon YAML:

```yaml
DB_USER: !var db/user
DB_PASS: !var db/password
```

In CI/CD pipelines set `SUMMON_PROVIDER=conceal` and ensure the `conceal` binary
is on the `PATH`.

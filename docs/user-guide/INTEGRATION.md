# Summon Integration

Conceal can serve as a Summon provider. Install it:
```bash
conceal summon install
```

Then reference secrets in a Summon YAML map:
```yaml
DB_USER: !var db/user
DB_PASS: !var db/password
```

In CI pipelines, export `SUMMON_PROVIDER=conceal` and ensure the binary is in the `PATH`.
